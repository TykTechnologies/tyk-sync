package gateway

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
	uuid "github.com/satori/go.uuid"
)

type Client struct {
	url                string
	secret             string
	InsecureSkipVerify bool
}

const (
	endpointAPIs     string = "/tyk/apis/"
	endpointCerts    string = "/tyk/certs"
	reloadAPIs       string = "/tyk/reload/group"
	endpointPolicies string = "/tyk/policies"
)

var (
	UseUpdateError error = errors.New("Object seems to exist (same API ID, Listen Path or Slug), use update()")
	UseCreateError error = errors.New("Object does not exist, use create()")
)

type APIMessage struct {
	Key     string `json:"key"`
	Status  string `json:"status"`
	Action  string `json:"action"`
	Message string `json:"message"`
}

type APISList []objects.APIDefinition

func NewGatewayClient(url, secret string) (*Client, error) {
	return &Client{
		url:    url,
		secret: secret,
	}, nil
}

func (c *Client) SetInsecureTLS(val bool) {
	c.InsecureSkipVerify = val
}

func (c *Client) GetActiveID(def *objects.DBApiDefinition) string {
	return def.APIID
}

func (c *Client) FetchAPIs() ([]objects.DBApiDefinition, error) {
	fullPath := urljoin.Join(c.url, endpointAPIs)

	ro := &grequests.RequestOptions{
		Headers: map[string]string{
			"x-tyk-authorization": c.secret,
			"content-type":        "application/json",
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Get(fullPath, ro)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API Returned error: %v", resp.String())
	}

	apis := APISList{}
	if err := resp.JSON(&apis); err != nil {
		return nil, err
	}

	retList := make([]objects.DBApiDefinition, len(apis))
	for i := range apis {
		retList[i] = objects.DBApiDefinition{APIDefinition: &apis[i]}
	}

	return retList, nil
}

func getAPIsIdentifiers(apiDefs *[]objects.DBApiDefinition) (map[string]*objects.DBApiDefinition, map[string]*objects.DBApiDefinition, map[string]*objects.DBApiDefinition) {
	apiids := make(map[string]*objects.DBApiDefinition)
	slugs := make(map[string]*objects.DBApiDefinition)
	paths := make(map[string]*objects.DBApiDefinition)

	for i, apiDef := range *apiDefs {
		apiids[apiDef.APIID] = &(*apiDefs)[i]
		slugs[apiDef.Slug] = &(*apiDefs)[i]
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &(*apiDefs)[i]
	}

	return apiids, slugs, paths
}

func (c *Client) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	existingAPIs, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	apiids, slugs, paths := getAPIsIdentifiers(&existingAPIs)

	for i, apiDef := range *apiDefs {
		fmt.Printf("Creating API %v: %v\n", i, apiDef.Name)
		if nil != apiids[apiDef.APIID] {
			fmt.Println("Warning: API ID Exists")
			return UseUpdateError
		} else if nil != slugs[apiDef.Slug] {
			fmt.Println("Warning: Slug Exists")
			return UseUpdateError
		} else if nil != paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] {
			fmt.Println("Warning: Listen Path Exists")
			return UseUpdateError
		}

		data, err := json.Marshal(apiDef.APIDefinition)
		if err != nil {
			return err
		}

		// Create
		fullPath := urljoin.Join(c.url, endpointAPIs)
		createResp, err := grequests.Post(fullPath, &grequests.RequestOptions{
			JSON: data,
			Headers: map[string]string{
				"x-tyk-authorization": c.secret,
				"content-type":        "application/json",
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		})

		if err != nil {
			return err
		}

		if createResp.StatusCode != 200 {
			return fmt.Errorf("API Returned error: %v (code: %v)", createResp.String(), createResp.StatusCode)
		}

		var status APIMessage
		if err := createResp.JSON(&status); err != nil {
			return err
		}

		if status.Status != "ok" {
			return fmt.Errorf("API request completed, but with error: %v", status.Message)
		}

		// initiate a reload
		go c.Reload()

		// Add updated API to existing API list.
		apiids[apiDef.APIID] = &(*apiDefs)[i]
		slugs[apiDef.Slug] = &(*apiDefs)[i]
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &(*apiDefs)[i]

		fmt.Printf("--> Status: OK, ID:%v\n", apiDef.APIID)
	}

	return nil
}

func (c *Client) Reload() error {
	// Reload
	fmt.Println("Reloading...")
	fullPath := urljoin.Join(c.url, reloadAPIs)
	reloadREsp, err := grequests.Get(fullPath, &grequests.RequestOptions{
		Headers: map[string]string{
			"x-tyk-authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	})

	if err != nil {
		return err
	}

	if reloadREsp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v (code: %v)", reloadREsp.String(), reloadREsp.StatusCode)
	}

	var status APIMessage
	if err := reloadREsp.JSON(&status); err != nil {
		return err
	}

	if status.Status != "ok" {
		fmt.Errorf("API request completed, but with error: %v", status.Message)
	}

	return nil
}

func (c *Client) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	existingAPIs, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	apiids, slugs, paths := getAPIsIdentifiers(&existingAPIs)

	for i, apiDef := range *apiDefs {
		fmt.Printf("Updating API %v: %v\n", i, apiDef.Name)
		if nil == apiids[apiDef.APIID] {
			return UseCreateError
		}

		// Update
		if apiDef.APIID == "" {
			return errors.New("API ID must be set")
		}

		data, err := json.Marshal(apiDef.APIDefinition)
		if err != nil {
			return err
		}

		updatePath := urljoin.Join(c.url, endpointAPIs, apiDef.APIID)
		uResp, err := grequests.Put(updatePath, &grequests.RequestOptions{
			JSON: data,
			Headers: map[string]string{
				"x-tyk-authorization": c.secret,
				"content-type":        "application/json",
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		})

		if err != nil {
			return err
		}

		if uResp.StatusCode != 200 {
			return fmt.Errorf("API Returned error: %v (code: %v)", uResp.String(), uResp.StatusCode)
		}

		// initiate a reload
		go c.Reload()

		// Add updated API to existing API list.
		apiids[apiDef.APIID] = &(*apiDefs)[i]
		slugs[apiDef.Slug] = &(*apiDefs)[i]
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &(*apiDefs)[i]

		fmt.Printf("--> Status: OK, ID:%v\n", apiDef.APIID)
	}

	return nil
}

func (c *Client) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	deleteAPIs := []string{}
	updateAPIs := []objects.DBApiDefinition{}
	createAPIs := []objects.DBApiDefinition{}

	apis, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	GWIDMap := map[string]int{}
	GitIDMap := map[string]int{}

	// Build the gw ID map
	for i, api := range apis {
		// Lets get a full list of existing IDs
		GWIDMap[api.APIID] = i
	}

	// Build the Git ID Map
	for i, def := range apiDefs {
		if def.APIID != "" {
			GitIDMap[def.APIID] = i
		} else {
			created := fmt.Sprintf("temp-%v", uuid.NewV4().String())
			GitIDMap[created] = i
		}
	}

	// Updates are when we find items in git that are also in dash
	for key, index := range GitIDMap {
		_, ok := GWIDMap[key]
		if ok {
			updateAPIs = append(updateAPIs, apiDefs[index])
		}
	}

	// Deletes are when we find items in the dash that are not in git
	for key, _ := range GWIDMap {
		_, ok := GitIDMap[key]
		if !ok {
			deleteAPIs = append(deleteAPIs, key)
		}
	}

	// Create operations are when we find things in Git that are not in the dashboard
	for key, index := range GitIDMap {
		_, ok := GWIDMap[key]
		if !ok {
			createAPIs = append(createAPIs, apiDefs[index])
		}
	}

	fmt.Printf("Deleting: %v\n", len(deleteAPIs))
	fmt.Printf("Updating: %v\n", len(updateAPIs))
	fmt.Printf("Creating: %v\n", len(createAPIs))

	// Do the deletes
	for _, dbId := range deleteAPIs {
		fmt.Printf("SYNC Deleting: %v\n", dbId)
		if err := c.deleteAPI(dbId); err != nil {
			return err
		}
	}

	// Do the updates
	// TODO
	//fmt.Printf("SYNC Updating: %v\n", api.APIID)
	if err := c.UpdateAPIs(&updateAPIs); err != nil {
		fmt.Println("ERR:", err)
		return err
	}

	// Do the creates
	// TODO
	//	fmt.Printf("SYNC Creating: %v\n", api.Name)
	//	fmt.Printf("--> ID: %v\n", id)
	if err := c.CreateAPIs(&createAPIs); err != nil {
		return err
	}

	return nil
}

func (c *Client) DeleteAPI(id string) error {
	return c.deleteAPI(id)
}

func (c *Client) deleteAPI(id string) error {
	delPath := urljoin.Join(c.url, endpointAPIs)
	delPath += id

	delResp, err := grequests.Delete(delPath, &grequests.RequestOptions{
		Headers: map[string]string{
			"x-tyk-authorization": c.secret,
			"content-type":        "application/json",
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	})

	if err != nil {
		return err
	}

	if delResp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v", delResp.String())
	}

	// initiate a reload
	go c.Reload()

	return nil
}
