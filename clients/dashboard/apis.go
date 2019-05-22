package dashboard

import (
	"fmt"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
	"github.com/satori/go.uuid"
	"gopkg.in/mgo.v2/bson"
)

type APIResponse struct {
	Message string
	Meta    string
	Status  string
}

type APISResponse struct {
	Apis  []objects.DBApiDefinition `json:"apis"`
	Pages int                       `json:"pages"`
}

func (c *Client) fixDBDef(def *objects.DBApiDefinition) {
	if def.HookReferences == nil {
		def.HookReferences = make([]interface{}, 0)
	}
}

func (c *Client) SetInsecureTLS(val bool) {
	c.InsecureSkipVerify = val
}

func (c *Client) GetActiveID(def *apidef.APIDefinition) string {
	return def.Id.Hex()
}

func (c *Client) CreateAPI(def *apidef.APIDefinition) (string, error) {
	fullPath := urljoin.Join(c.url, endpointAPIs)

	ro := &grequests.RequestOptions{
		Params: map[string]string{"p": "-2"},
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Get(fullPath, ro)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API Returned error: %v for %v", resp.String(), fullPath)
	}

	apis := APISResponse{}
	if err := resp.JSON(&apis); err != nil {
		return "", err
	}

	retainedIDs := false

	for _, api := range apis.Apis {
		if api.APIID == def.APIID {
			fmt.Println("Warning: API ID Exists")
			return "", UseUpdateError
		}

		if api.Id == def.Id {
			fmt.Println("Warning: Object ID Exists")
			return "", UseUpdateError
		}

		if api.Slug == def.Slug {
			fmt.Println("Warning: Slug Exists")
			return "", UseUpdateError
		}

		if api.Proxy.ListenPath == def.Proxy.ListenPath {
			if api.Domain == def.Domain {
				fmt.Println("Warning: Listen Path Exists")
				return "", UseUpdateError
			}
		}
	}

	if def.APIID != "" {
		// Retain the API ID
		retainedIDs = true
	}

	// Create
	asDBDef := objects.DBApiDefinition{APIDefinition: *def}
	c.fixDBDef(&asDBDef)

	createResp, err := grequests.Post(fullPath, &grequests.RequestOptions{
		JSON: asDBDef,
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	})

	if err != nil {
		return "", err
	}

	if createResp.StatusCode != 200 {
		return "", fmt.Errorf("API Returned error: %v (code: %v)", createResp.String(), createResp.StatusCode)
	}

	var status APIResponse
	if err := createResp.JSON(&status); err != nil {
		return "", err
	}

	if status.Status != "OK" {
		return "", fmt.Errorf("API request completed, but with error: %v", status.Message)
	}

	// Create will always reset the API ID on dashboard, if we want to retain it, we must use UPDATE
	if retainedIDs {
		def.Id = bson.ObjectIdHex(status.Meta)
		if err := c.UpdateAPI(def); err != nil {
			fmt.Printf("Problem trying to retain API ID: %v\n", err)
		}

	}

	return status.Meta, nil

}

func (c *Client) FetchAPIs() ([]objects.DBApiDefinition, error) {
	fullPath := urljoin.Join(c.url, endpointAPIs)

	ro := &grequests.RequestOptions{
		Params: map[string]string{"p": "-2"},
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Get(fullPath, ro)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API Returned error: %v for %v", resp.String(), fullPath)
	}

	apis := APISResponse{}
	if err := resp.JSON(&apis); err != nil {
		return nil, err
	}

	return apis.Apis, nil
}

func (c *Client) UpdateAPI(def *apidef.APIDefinition) error {
	fullPath := urljoin.Join(c.url, endpointAPIs)

	ro := &grequests.RequestOptions{
		Params: map[string]string{"p": "-2"},
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Get(fullPath, ro)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v for %v", resp.String(), fullPath)
	}

	apis := APISResponse{}
	if err := resp.JSON(&apis); err != nil {
		return err
	}

	found := false
	for _, api := range apis.Apis {
		// For an update, prefer API IDs
		if api.APIID == def.APIID {
			// Lets make sure we target the internal ID of the matching API ID
			def.Id = api.Id
			found = true
			break
		}

		// Dashboard uses it's own IDs
		if api.Id == def.Id {
			if def.APIID == "" {
				def.APIID = api.APIID
			}
			found = true
			break
		}

		// We can also match on the slug
		if api.Slug == def.Slug {
			if def.APIID == "" {
				def.APIID = api.APIID
			}
			if def.Id == "" {
				def.Id = api.Id
			}

			found = true
			break
		}

		// We can also match on the listen_path
		if api.Proxy.ListenPath == def.Proxy.ListenPath {
			if def.APIID == "" {
				def.APIID = api.APIID
			}
			if def.Id == "" {
				def.Id = api.Id
			}

			found = true
			break
		}
	}

	if !found {
		return UseCreateError
	}

	// Update
	asDBDef := objects.DBApiDefinition{APIDefinition: *def}
	c.fixDBDef(&asDBDef)

	updatePath := urljoin.Join(c.url, endpointAPIs, def.Id.Hex())
	updateResp, err := grequests.Put(updatePath, &grequests.RequestOptions{
		JSON: asDBDef,
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	})

	if err != nil {
		return err
	}

	if updateResp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v", updateResp.String())
	}

	var status APIResponse
	if err := updateResp.JSON(&status); err != nil {
		return err
	}

	if status.Status != "OK" {
		return fmt.Errorf("API request completed, but with error: %v", status.Message)
	}

	return nil
}

func (c *Client) Sync(apiDefs []apidef.APIDefinition) error {
	deleteAPIs := []string{}
	updateAPIs := []apidef.APIDefinition{}
	createAPIs := []apidef.APIDefinition{}

	// Fetch the running API list
	fullPath := urljoin.Join(c.url, endpointAPIs)

	ro := &grequests.RequestOptions{
		Params: map[string]string{"p": "-2"},
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Get(fullPath, ro)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v", resp.String())
	}

	apis := APISResponse{}
	if err := resp.JSON(&apis); err != nil {
		return err
	}

	DashIDMap := map[string]int{}
	GitIDMap := map[string]int{}

	// Build the dash ID map
	for i, api := range apis.Apis {
		// Lets get a full list of existing IDs
		if c.isCloud {
			DashIDMap[api.Slug] = i
			continue
		}
		DashIDMap[api.APIID] = i
	}

	// Build the Git ID Map
	for i, def := range apiDefs {
		if c.isCloud {
			GitIDMap[def.Slug] = i
			continue
		}

		if def.APIID != "" {
			GitIDMap[def.APIID] = i
			continue
		} else if def.Id.Hex() != "" {
			// No API ID? Let's try the actual DB ID
			GitIDMap[def.Id.Hex()] = i
			continue
		} else {
			created := fmt.Sprintf("temp-%v", uuid.NewV4().String())
			GitIDMap[created] = i
		}
	}

	// Updates are when we find items in git that are also in dash
	for key, index := range GitIDMap {
		fmt.Printf("Checking: %v\n", key)
		dashIndex, ok := DashIDMap[key]
		if ok {
			// Make sure we are targeting the correct DB ID
			api := apiDefs[index]
			api.Id = apis.Apis[dashIndex].Id
			api.APIID = apis.Apis[dashIndex].APIID
			updateAPIs = append(updateAPIs, api)
		}
	}

	// Deletes are when we find items in the dash that are not in git
	for key, dashIndex := range DashIDMap {
		_, ok := GitIDMap[key]
		if !ok {
			// Make sure we always target the DB ID
			deleteAPIs = append(deleteAPIs, apis.Apis[dashIndex].Id.Hex())
		}
	}

	// Create operations are when we find things in Git that are not in the dashboard
	for key, index := range GitIDMap {
		_, ok := DashIDMap[key]
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
		if err := c.DeleteAPI(dbId); err != nil {
			return err
		}
	}

	// Do the updates
	for _, api := range updateAPIs {
		fmt.Printf("SYNC Updating: %v\n", api.Id.Hex())
		if err := c.UpdateAPI(&api); err != nil {
			return err
		}
	}

	// Do the creates
	for _, api := range createAPIs {
		fmt.Printf("SYNC Creating: %v\n", api.Name)
		var err error
		var id string
		if id, err = c.CreateAPI(&api); err != nil {
			return err
		}
		fmt.Printf("--> ID: %v\n", id)
	}

	return nil
}

func (c *Client) DeleteAPI(id string) error {
	delPath := urljoin.Join(c.url, endpointAPIs, id)
	delResp, err := grequests.Delete(delPath, &grequests.RequestOptions{
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	})

	if err != nil {
		return err
	}

	if delResp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v", delResp.String())
	}

	return nil
}
