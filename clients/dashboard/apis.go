package dashboard

import (
	"fmt"

	"encoding/json"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/gofrs/uuid"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
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

	if def.IsOAS && def.OAS != nil {
		tykExt := def.OAS.GetTykExtension()
		if tykExt != nil && def.Id != "" {
			tykExt.Info.DBID = def.Id
		}
	}
}

func (c *Client) SetInsecureTLS(val bool) {
	c.InsecureSkipVerify = val
}

func (c *Client) GetActiveID(def *objects.DBApiDefinition) string {
	return def.Id.Hex()
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

func (c *Client) FetchAPI(apiID string) (objects.DBApiDefinition, error) {
	api := objects.DBApiDefinition{}
	fullPath := urljoin.Join(c.url, endpointAPIs, apiID)

	ro := &grequests.RequestOptions{
		Params: map[string]string{"p": "-2"},
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Get(fullPath, ro)
	if err != nil {
		return api, err
	}

	if resp.StatusCode != 200 {
		return api, fmt.Errorf("API %v Returned error: %v for %v", apiID, resp.String(), fullPath)
	}

	if err := resp.JSON(&api); err != nil {
		return api, err
	}

	return api, nil
}

func getAPIsIdentifiers(apiDefs *[]objects.DBApiDefinition) (map[string]*objects.DBApiDefinition, map[string]*objects.DBApiDefinition, map[string]*objects.DBApiDefinition, map[string]*objects.DBApiDefinition) {
	apiids := make(map[string]*objects.DBApiDefinition)
	ids := make(map[string]*objects.DBApiDefinition)
	slugs := make(map[string]*objects.DBApiDefinition)
	paths := make(map[string]*objects.DBApiDefinition)

	for i := range *apiDefs {
		apiDef := (*apiDefs)[i]
		apiids[apiDef.APIID] = &apiDef
		ids[apiDef.Id.Hex()] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &apiDef
	}

	return apiids, ids, slugs, paths
}

func (c *Client) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	existingAPIs, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	apiids, ids, slugs, paths := getAPIsIdentifiers(&existingAPIs)

	retainAPIIdList := make([]objects.DBApiDefinition, 0)
	for i := range *apiDefs {
		apiDef := (*apiDefs)[i]
		fmt.Printf("Creating API %v: %v\n", i, apiDef.Name)
		if thisAPI, ok := apiids[apiDef.APIID]; ok && thisAPI != nil {
			fmt.Println("Warning: API ID Exists")
			return UseUpdateError
		} else if thisAPI, ok := ids[apiDef.Id.Hex()]; ok && thisAPI != nil {
			fmt.Println("Warning: Object ID Exists")
			return UseUpdateError
		} else if thisAPI, ok := slugs[apiDef.Slug]; ok && thisAPI != nil {
			fmt.Println("Warning: Slug Exists")
			return UseUpdateError
		} else if thisAPI, ok := paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain]; ok && thisAPI != nil {
			fmt.Println("Warning: Listen Path Exists")
			return UseUpdateError
		}

		// Create
		asDBDef := &apiDef
		c.fixDBDef(asDBDef)

		data, err := json.Marshal(asDBDef)
		if err != nil {
			return err
		}

		fullPath := urljoin.Join(c.url, endpointAPIs)
		createResp, err := grequests.Post(fullPath, &grequests.RequestOptions{
			JSON: data,
			Headers: map[string]string{
				"Authorization": c.secret,
			},
			Params: map[string]string{
				"accept_additional_properties": "true",
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		})

		if err != nil {
			return err
		}

		if createResp.StatusCode != 200 {
			return fmt.Errorf("API Returned error: %v (code: %v)", createResp.String(), createResp.StatusCode)
		}

		var status APIResponse
		if err := createResp.JSON(&status); err != nil {
			return err
		}

		if status.Status != "OK" {
			return fmt.Errorf("API request completed, but with error: %v", status.Message)
		}

		// Update apiDef with its ID before adding it to the existing APIs list.
		apiDef.Id = apidef.ObjectIdHex(status.Meta)

		// Create will always reset the API ID on dashboard, if we want to retain it, we must use UPDATE
		if apiDef.APIID != "" {
			retainAPIIdList = append(retainAPIIdList, apiDef)
		}

		// Add created API to existing API list.
		apiids[apiDef.APIID] = &apiDef
		ids[apiDef.Id.Hex()] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &apiDef

		fmt.Printf("--> Status: OK, ID:%v\n", apiDef.APIID)
	}

	if err := c.UpdateAPIs(&retainAPIIdList); err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	existingAPIs, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	apiids, ids, slugs, paths := getAPIsIdentifiers(&existingAPIs)

	for i := range *apiDefs {
		apiDef := (*apiDefs)[i]
		fmt.Printf("Updating API %v: %v\n", i, apiDef.Name)
		if thisAPI, ok := apiids[apiDef.APIID]; ok && thisAPI != nil {
			apiDef.Id = thisAPI.Id
		} else if thisAPI, ok := ids[apiDef.Id.Hex()]; ok && thisAPI != nil {
			if apiDef.APIID == "" {
				apiDef.APIID = thisAPI.APIID
			}
		} else if thisAPI, ok := slugs[apiDef.Slug]; ok && thisAPI != nil {
			if apiDef.APIID == "" {
				apiDef.APIID = thisAPI.APIID
			}
			if apiDef.Id == "" {
				apiDef.Id = thisAPI.Id
			}
		} else if thisAPI, ok := paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain]; ok && thisAPI != nil {
			if apiDef.APIID == "" {
				apiDef.APIID = thisAPI.APIID
			}
			if apiDef.Id == "" {
				apiDef.Id = thisAPI.Id
			}
		} else {
			return UseCreateError
		}

		// Update
		asDBDef := &apiDef
		c.fixDBDef(asDBDef)

		endpoint := endpointAPIs
		var payload interface{}
		payload = asDBDef
		if apiDef.IsOAS {
			endpoint = endpointOASAPIs
			payload = asDBDef.OAS
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		updatePath := urljoin.Join(c.url, endpoint, apiDef.Id.Hex())
		updateResp, err := grequests.Put(updatePath, &grequests.RequestOptions{
			JSON: data,
			Headers: map[string]string{
				"Authorization": c.secret,
			},
			Params: map[string]string{
				"accept_additional_properties": "true",
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		})

		if err != nil {
			return err
		}

		if updateResp.StatusCode != 200 {
			return fmt.Errorf("API Updating Returned error: %v", updateResp.String())
		}

		var status APIResponse
		if err := updateResp.JSON(&status); err != nil {
			return err
		}

		if status.Status != "OK" {
			return fmt.Errorf("API request completed, but with error: %v", status.Message)
		}

		// Add updated API to existing API list.
		apiids[apiDef.APIID] = &apiDef
		ids[apiDef.Id.Hex()] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.Proxy.ListenPath+"-"+apiDef.Domain] = &apiDef

		fmt.Printf("--> Status: OK, ID:%v\n", apiDef.APIID)
	}

	return nil
}

func (c *Client) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	deleteAPIs := []string{}
	updateAPIs := []objects.DBApiDefinition{}
	createAPIs := []objects.DBApiDefinition{}

	existingAPIs, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	DashIDMap := map[string]int{}
	GitIDMap := map[string]int{}

	// Build the dash ID map
	for i, api := range existingAPIs {
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
			uid, err := uuid.NewV4()
			if err != nil {
				fmt.Println("error generating UUID", err)
				return err
			}
			created := fmt.Sprintf("temp-%v", uid.String())
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
			api.Id = existingAPIs[dashIndex].Id
			api.APIID = existingAPIs[dashIndex].APIID
			updateAPIs = append(updateAPIs, api)
		}
	}

	// Deletes are when we find items in the dash that are not in git
	for key, dashIndex := range DashIDMap {
		_, ok := GitIDMap[key]
		if !ok {
			// Make sure we always target the DB ID
			deleteAPIs = append(deleteAPIs, existingAPIs[dashIndex].Id.Hex())
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
	if err := c.UpdateAPIs(&updateAPIs); err != nil {
		return err
	}
	for _, apiDef := range updateAPIs {
		fmt.Printf("SYNC Updated: %v\n", apiDef.Id.Hex())
	}

	// Do the creates
	if err := c.CreateAPIs(&createAPIs); err != nil {
		return err
	}
	for _, apiDef := range createAPIs {
		fmt.Printf("SYNC Created: %v\n", apiDef.Name)
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
