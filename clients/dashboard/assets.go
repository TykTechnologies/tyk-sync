package dashboard

import (
	"encoding/json"
	"fmt"

	"github.com/TykTechnologies/storage/persistent/model"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/gofrs/uuid"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
)

func (c *Client) FetchAssets() ([]objects.DBAssets, error) {
	fullPath := urljoin.Join(c.url, endpointAssets)

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

	var assets []objects.DBAssets
	if err := resp.JSON(&assets); err != nil {
		return nil, err
	}

	return assets, nil
}

func (c *Client) FetchAsset(assetID string) (objects.DBAssets, error) {
	asset := objects.DBAssets{}
	fullPath := urljoin.Join(c.url, endpointAssets, assetID)

	ro := &grequests.RequestOptions{
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Get(fullPath, ro)
	if err != nil {
		return asset, err
	}

	if resp.StatusCode != 200 {
		return asset, fmt.Errorf("API %v Returned error: %v for %v", assetID, resp.String(), fullPath)
	}

	if err := resp.JSON(&asset); err != nil {
		return asset, err
	}

	return asset, nil
}

func getAssetsIdentifiers(assets *[]objects.DBAssets) map[string]*objects.DBAssets {
	ids := make(map[string]*objects.DBAssets)

	for i := range *assets {
		asset := (*assets)[i]
		ids[asset.ID] = &asset
	}

	return ids
}

func (c *Client) CreateAssets(assetsDefs *[]objects.DBAssets) error {
	existingAssets, err := c.FetchAssets()
	if err != nil {
		return err
	}

	ids := getAssetsIdentifiers(&existingAssets)

	for i := range *assetsDefs {
		assetDef := (*assetsDefs)[i]
		fmt.Printf("Creating Assets %v: %v\n", i, assetDef.Name)
		if thisAPI, ok := ids[assetDef.ID]; ok && thisAPI != nil {
			fmt.Println("Warning: Asset ID already exists")
			return UseAssetUpdateError
		}

		// Create
		data, err := json.Marshal(assetDef)
		if err != nil {
			return err
		}

		fullPath := urljoin.Join(c.url, endpointAssets)
		createResp, err := grequests.Post(fullPath, &grequests.RequestOptions{
			JSON: data,
			Headers: map[string]string{
				"Authorization": c.secret,
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		})

		if err != nil {
			return err
		}

		if createResp.StatusCode != 201 {
			return fmt.Errorf("API Returned error: %v (code: %v)", createResp.String(), createResp.StatusCode)
		}

		var status APIResponse
		if err := createResp.JSON(&status); err != nil {
			return err
		}

		if status.Status != "success" {
			return fmt.Errorf("API request completed, but with error: %v", status.Message)
		}

		// Update assetDef with its ID before adding it to the existing APIs list.
		assetDef.DBId = model.ObjectIDHex(status.Meta)

		// Add created API to existing API list.
		ids[assetDef.ID] = &assetDef

		fmt.Printf("--> Status: OK, ID:%v\n", assetDef.ID)
	}

	return nil
}

func (c *Client) SyncAssets(assetsDefs []objects.DBAssets) error {
	deleteAssets := []string{}
	updateAssets := []objects.DBAssets{}
	createAssets := []objects.DBAssets{}

	existingAssets, err := c.FetchAssets()
	if err != nil {
		return err
	}

	DashIDMap := map[string]int{}
	GitIDMap := map[string]int{}

	// Build the dash ID map
	for i, asset := range existingAssets {
		DashIDMap[asset.ID] = i
	}

	// Build the Git ID Map
	for i, def := range assetsDefs {
		if def.ID != "" {
			GitIDMap[def.ID] = i
			continue
		} else if def.DBId.Hex() != "" {
			//  No Asset ID? Let's try the actual DB ID
			GitIDMap[def.DBId.Hex()] = i
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
			api := assetsDefs[index]
			api.DBId = existingAssets[dashIndex].DBId
			api.ID = existingAssets[dashIndex].ID
			updateAssets = append(updateAssets, api)
		}
	}

	// Deletes are when we find items in the dash that are not in git
	for key, dashIndex := range DashIDMap {
		_, ok := GitIDMap[key]
		if !ok {
			// Make sure we always target the DB ID
			deleteAssets = append(deleteAssets, existingAssets[dashIndex].DBId.Hex())
		}
	}

	// Create operations are when we find things in Git that are not in the dashboard
	for key, index := range GitIDMap {
		_, ok := DashIDMap[key]
		if !ok {
			createAssets = append(createAssets, assetsDefs[index])
		}
	}

	fmt.Printf("Deleting: %v\n", len(deleteAssets))
	fmt.Printf("Updating: %v\n", len(updateAssets))
	fmt.Printf("Creating: %v\n", len(createAssets))

	// Do the deletes
	for _, dbId := range deleteAssets {
		fmt.Printf("SYNC Deleting: %v\n", dbId)
		if err := c.DeleteAssets(dbId); err != nil {
			return err
		}
	}

	// Do the updates
	if err := c.UpdateAssets(&updateAssets); err != nil {
		return err
	}
	for _, apiDef := range updateAssets {
		fmt.Printf("SYNC Updated: %v\n", apiDef.DBId.Hex())
	}

	// Do the creates
	if err := c.CreateAssets(&createAssets); err != nil {
		return err
	}
	for _, apiDef := range createAssets {
		fmt.Printf("SYNC Created: %v\n", apiDef.Name)
	}

	return nil
}

func (c *Client) UpdateAssets(assetsDef *[]objects.DBAssets) error {
	existingAssets, err := c.FetchAssets()
	if err != nil {
		return err
	}

	ids := getAssetsIdentifiers(&existingAssets)

	for i := range *assetsDef {
		assetDef := (*assetsDef)[i]
		fmt.Printf("Updating Asset %v: %v\n", i, assetDef.Name)
		if thisAsset, ok := ids[assetDef.ID]; ok && thisAsset != nil {
			assetDef.ID = thisAsset.ID
			assetDef.DBId = thisAsset.DBId
		} else {
			return UseAssetUpdateError
		}

		// Update
		data, err := json.Marshal(assetDef)
		if err != nil {
			return err
		}

		updatePath := urljoin.Join(c.url, endpointAssets, assetDef.ID)
		updateResp, err := grequests.Put(updatePath, &grequests.RequestOptions{
			JSON: data,
			Headers: map[string]string{
				"Authorization": c.secret,
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		})

		if err != nil {
			return err
		}

		if updateResp.StatusCode != 200 {
			return fmt.Errorf("Assets Updating Returned error: %v", updateResp.String())
		}

		var status APIResponse
		if err := updateResp.JSON(&status); err != nil {
			return err
		}

		if status.Status != "success" {
			return fmt.Errorf("Assets request completed, but with error: %v", status.Message)
		}

		fmt.Printf("--> Status: OK, ID:%v\n", assetDef.ID)
	}

	return nil
}

func (c *Client) DeleteAssets(id string) error {
	delPath := urljoin.Join(c.url, endpointAssets, id)
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
