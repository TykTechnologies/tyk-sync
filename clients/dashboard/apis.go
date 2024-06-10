package dashboard

import (
	"encoding/json"
	"fmt"
	"github.com/gofrs/uuid"

	"github.com/TykTechnologies/storage/persistent/model"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk/apidef/oas"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
)

type APIResponse struct {
	Message string
	Meta    string
	Status  string
}

type APISResponse struct {
	OASApis []oas.OAS                 `json:"oasApis"`
	Apis    []objects.DBApiDefinition `json:"apis"`
	Pages   int                       `json:"pages"`
}

func (c *Client) fixDBDef(def *objects.DBApiDefinition) {
	if def.HookReferences == nil {
		def.HookReferences = make([]interface{}, 0)
	}

	if def.IsOAS && def.OAS != nil {
		tykExt := def.OAS.GetTykExtension()
		if tykExt != nil && def.GetDBID() != "" {
			tykExt.Info.DBID = def.GetDBID()
		}
	}
}

func (c *Client) SetInsecureTLS(val bool) {
	c.InsecureSkipVerify = val
}

func (c *Client) GetActiveID(def *objects.DBApiDefinition) string {
	return def.GetDBID().Hex()
}

func (c *Client) FetchOASCategory(id string) ([]string, error) {
	fullPath := urljoin.Join(c.url, endpointOASAPIs, "/"+id, endpointCategories)

	getResp, err := grequests.Get(fullPath, &grequests.RequestOptions{
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		Params: map[string]string{
			"accept_additional_properties": "true",
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	})
	if err != nil {
		return nil, err
	}

	if getResp.StatusCode != 200 {
		return nil, fmt.Errorf(
			"OAS Categories API Returned error: %v (code: %v)", getResp.String(), getResp.StatusCode,
		)
	}

	categoriesOutput := CategoriesPayload{}
	if err := getResp.JSON(&categoriesOutput); err != nil {
		return nil, fmt.Errorf("failed to read OAS Category API output, %v", getResp.String())
	}

	return categoriesOutput.Categories, nil
}

func (c *Client) FetchOASAPI(id string) (*oas.OAS, error) {
	fullPath := urljoin.Join(c.url, endpointOASAPIs, id, "export")

	ro := &grequests.RequestOptions{
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

	oasApi := &oas.OAS{}
	if err := oasApi.UnmarshalJSON(resp.Bytes()); err != nil {
		return nil, err
	}

	return oasApi, nil
}

func (c *Client) FetchAPIs() (*APISResponse, error) {
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

	apisResponse := APISResponse{}
	if err := resp.JSON(&apisResponse); err != nil {
		return nil, err
	}

	var oasApis []oas.OAS

	for i := range apisResponse.Apis {
		if apisResponse.Apis[i].IsOASAPI() {
			oasApi, err := c.FetchOASAPI(apisResponse.Apis[i].GetAPIID())
			if err != nil {
				fmt.Printf("Failed to fetch OAS API: %v, err: %v", apisResponse.Apis[i].GetAPIID(), err)
				continue
			}

			oasApis = append(oasApis, *oasApi)
		}
	}

	apisResponse.OASApis = oasApis

	return &apisResponse, nil
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
		apiids[apiDef.GetAPIID()] = &apiDef
		ids[apiDef.GetDBID().Hex()] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.GetListenPath()+"-"+apiDef.GetDomain()] = &apiDef
	}

	return apiids, ids, slugs, paths
}

func (c *Client) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	resp, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	existingAPIs := resp.Apis
	apiids, ids, slugs, paths := getAPIsIdentifiers(&existingAPIs)

	retainAPIIdList := make([]objects.DBApiDefinition, 0)

	for i := range *apiDefs {
		apiDef := (*apiDefs)[i]
		fmt.Printf("Creating API %v: %v\n", i, apiDef.GetAPIName())

		if thisAPI, ok := apiids[apiDef.GetAPIID()]; ok && thisAPI != nil {
			fmt.Println("Warning: API ID Exists")
			return UseUpdateError
		} else if thisAPI, ok := ids[apiDef.GetDBID().Hex()]; ok && thisAPI != nil {
			fmt.Println("Warning: Object ID Exists")
			return UseUpdateError
		} else if thisAPI, ok := slugs[apiDef.Slug]; apiDef.Slug != "" && ok && thisAPI != nil {
			fmt.Println("Warning: Slug Exists")
			return UseUpdateError
		} else if thisAPI, ok := paths[apiDef.GetListenPath()+"-"+apiDef.GetDomain()]; ok && thisAPI != nil {
			fmt.Println("Warning: Listen Path Exists")
			return UseUpdateError
		}

		// Create
		asDBDef := &apiDef
		c.fixDBDef(asDBDef)

		var data []byte

		fullPath := urljoin.Join(c.url, endpointAPIs)

		if asDBDef.APIDefinition != nil {
			switch asDBDef.IsOASAPI() {
			case true:
				fullPath = urljoin.Join(c.url, endpointOASAPIs)

				data, err = json.Marshal(asDBDef.OAS)
				if err != nil {
					return err
				}
			default:
				data, err = json.Marshal(asDBDef)
				if err != nil {
					return err
				}
			}
		}

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
		apiDef.SetDBID(model.ObjectIDHex(status.Meta))

		// Create will always reset the API ID on dashboard, if we want to retain it, we must use UPDATE
		if apiDef.GetAPIID() != "" {
			retainAPIIdList = append(retainAPIIdList, apiDef)
		}

		// Add created API to existing API list.
		apiids[apiDef.GetAPIID()] = &apiDef
		ids[apiDef.GetDBID().Hex()] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.GetListenPath()+"-"+apiDef.GetDomain()] = &apiDef

		if asDBDef.IsOASAPI() && len(asDBDef.Categories) > 0 {
			resp, err := c.UpdateOASCategory(asDBDef)
			if err != nil {
				return err
			}

			fmt.Printf("OAS API Categories updated, %v", resp.String())
		}

		fmt.Printf("--> Status: OK, ID:%v\n", apiDef.GetAPIID())
	}

	if err := c.UpdateAPIs(&retainAPIIdList); err != nil {
		return err
	}

	return nil
}

func (c *Client) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	resp, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	existingAPIs := resp.Apis
	apiids, ids, slugs, paths := getAPIsIdentifiers(&existingAPIs)

	for i := range *apiDefs {
		apiDef := (*apiDefs)[i]
		fmt.Printf("Updating API %v: %v\n", i, apiDef.GetAPIName())
		if thisAPI, ok := apiids[apiDef.GetAPIID()]; ok && thisAPI != nil {
			apiDef.SetDBID(thisAPI.GetDBID())
		} else if thisAPI, ok := ids[apiDef.GetDBID().Hex()]; ok && thisAPI != nil {
			if apiDef.GetAPIID() == "" {
				apiDef.SetAPIID(thisAPI.GetAPIID())
			}
		} else if thisAPI, ok := slugs[apiDef.Slug]; ok && thisAPI != nil {
			if apiDef.GetAPIID() == "" {
				apiDef.SetAPIID(thisAPI.GetAPIID())
			}
			if apiDef.GetDBID() == "" {
				apiDef.SetDBID(thisAPI.GetDBID())
			}
		} else if thisAPI, ok := paths[apiDef.GetListenPath()+"-"+apiDef.GetDomain()]; ok && thisAPI != nil {
			if apiDef.GetAPIID() == "" {
				apiDef.SetAPIID(thisAPI.GetAPIID())
			}
			if apiDef.GetDBID() == "" {
				apiDef.SetDBID(thisAPI.GetDBID())
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
		if apiDef.IsOASAPI() {
			endpoint = endpointOASAPIs
			payload = asDBDef.OAS
		}

		data, err := json.Marshal(payload)
		if err != nil {
			return err
		}

		updatePath := urljoin.Join(c.url, endpoint, apiDef.GetDBID().Hex())
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

		if asDBDef.IsOASAPI() && len(asDBDef.Categories) > 0 {
			resp, err := c.UpdateOASCategory(asDBDef)
			if err != nil {
				return err
			}

			fmt.Printf("OAS API Categories updated, %v", resp.String())
		}

		// Add updated API to existing API list.
		apiids[apiDef.GetAPIID()] = &apiDef
		ids[apiDef.GetDBID().Hex()] = &apiDef
		slugs[apiDef.Slug] = &apiDef
		paths[apiDef.GetListenPath()+"-"+apiDef.GetDomain()] = &apiDef

		fmt.Printf("--> Status: OK, ID:%v\n", apiDef.GetAPIID())
	}

	return nil
}

func (c *Client) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	deleteAPIs := []string{}
	updateAPIs := []objects.DBApiDefinition{}
	createAPIs := []objects.DBApiDefinition{}

	resp, err := c.FetchAPIs()
	if err != nil {
		return err
	}

	existingAPIs := resp.Apis

	DashIDMap := map[string]int{}
	GitIDMap := map[string]int{}

	// Build the dash ID map
	for i, api := range existingAPIs {
		id := ""

		if api.APIDefinition != nil && !api.APIDefinition.IsOAS {
			// Lets get a full list of existing IDs
			if c.isCloud {
				DashIDMap[api.Slug] = i
				continue
			}

			id, err = parseId(api.GetAPIID(), api.GetDBID().Hex())
			if err != nil {
				return err
			}
		} else if api.OAS != nil {
			if api.OAS.GetTykExtension() != nil {
				id, err = parseId(api.OAS.GetTykExtension().Info.ID, api.OAS.GetTykExtension().Info.DBID.Hex())
				if err != nil {
					return err
				}
			}
		}

		if id != "" {
			DashIDMap[id] = i
		}
	}

	// Build the Git ID Map
	for i, def := range apiDefs {
		id := ""
		if def.APIDefinition != nil && !def.APIDefinition.IsOAS {
			if c.isCloud {
				GitIDMap[def.Slug] = i
				continue
			}

			id, err = parseId(def.GetAPIID(), def.GetDBID().Hex())
			if err != nil {
				return err
			}
		} else if def.OAS != nil {
			if c.isCloud {
				GitIDMap[def.Slug] = i
				continue
			}

			if def.OAS.GetTykExtension() != nil {
				id, err = parseId(def.OAS.GetTykExtension().Info.ID, def.OAS.GetTykExtension().Info.DBID.Hex())
				if err != nil {
					return err
				}
			}
		} else {
			continue
		}

		if id != "" {
			GitIDMap[id] = i
		}
	}

	// Updates are when we find items in git that are also in dash
	for key, index := range GitIDMap {
		fmt.Printf("Checking: %v\n", key)
		dashIndex, ok := DashIDMap[key]
		if ok {
			// Make sure we are targeting the correct DB ID
			api := apiDefs[index]
			existingApi := existingAPIs[dashIndex]
			if existingApi.APIDefinition != nil && !existingApi.IsOAS {
				api.SetDBID(existingApi.GetDBID())
				api.SetAPIID(existingApi.GetAPIID())

			} else if existingApi.OAS != nil {
				if existingApi.OAS.GetTykExtension() == nil {
					return fmt.Errorf("invalid OAS doc, expected x-tyk-gateway field exists")
				}

				api.OAS.GetTykExtension().Info.ID = existingApi.OAS.GetTykExtension().Info.ID
				api.OAS.GetTykExtension().Info.DBID = existingApi.OAS.GetTykExtension().Info.DBID
			}

			updateAPIs = append(updateAPIs, api)
		}
	}

	// Deletes are when we find items in the dash that are not in git
	for key, dashIndex := range DashIDMap {
		_, ok := GitIDMap[key]
		if !ok {
			// Make sure we always target the DB ID
			deleteAPIs = append(deleteAPIs, existingAPIs[dashIndex].GetDBID().Hex())
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
		if apiDef.IsOASAPI() {
			fmt.Printf("SYNC Updated OAS API Definition %v\n", apiDef.GetDBID().Hex())
		} else {
			fmt.Printf("SYNC Updated Classic API Definition %v\n", apiDef.GetDBID().Hex())
		}
	}

	// Do the creates
	if err := c.CreateAPIs(&createAPIs); err != nil {
		return err
	}
	for _, apiDef := range createAPIs {
		fmt.Printf("SYNC Created: %v\n", apiDef.GetAPIName())
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

func parseId(apiID, dbIDHex string) (string, error) {
	id := ""
	if apiID != "" {
		id = apiID
	} else if dbIDHex != "" {
		// No API ID? Let's try the actual DB ID
		id = dbIDHex
	} else {
		uid, err := uuid.NewV4()
		if err != nil {
			fmt.Println("error generating UUID", err)
			return "", err
		}

		created := fmt.Sprintf("temp-%v", uid.String())
		id = created
	}

	return id, nil
}
