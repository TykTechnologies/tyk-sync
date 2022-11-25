package dashboard

import (
	"fmt"
	"gopkg.in/mgo.v2/bson"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/kataras/go-errors"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
	uuid "github.com/satori/go.uuid"
)

type PoliciesData struct {
	Data  []objects.Policy
	Pages int
}

func (c *Client) FetchPolicies() ([]objects.Policy, error) {
	fullPath := urljoin.Join(c.url, endpointPolicies)

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
		return nil, fmt.Errorf("API Returned error: %v", resp.String())
	}

	policies := PoliciesData{}
	if err := resp.JSON(&policies); err != nil {
		return nil, err
	}

	return policies.Data, nil
}

func getPoliciesIdentifiers(pols *[]objects.Policy) (map[string]*objects.Policy, map[string]*objects.Policy) {
	mids := make(map[string]*objects.Policy)
	ids := make(map[string]*objects.Policy)

	for i := range *pols {
		pol := (*pols)[i]
		mids[pol.MID.Hex()] = &pol
		ids[pol.ID] = &pol
	}

	return mids, ids
}

func (c *Client) CreatePolicies(pols *[]objects.Policy) error {
	existingPols, err := c.FetchPolicies()
	if err != nil {
		return err
	}

	mids, ids := getPoliciesIdentifiers(&existingPols)

	for i := range *pols {
		pol := (*pols)[i]
		fmt.Printf("Creating Policy %v: %v\n", i, pol.Name)
		if nil != mids[pol.MID.Hex()] {
			fmt.Println("Warning: Policy MID Exists")
			return UseUpdateError
		} else if nil != ids[pol.ID] {
			fmt.Println("Warning: Policy ID Exists")
			return UseUpdateError
		}

		fullPath := urljoin.Join(c.url, endpointPolicies)

		ro := &grequests.RequestOptions{
			JSON: pol,
			Headers: map[string]string{
				"Authorization": c.secret,
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		}

		resp, err := grequests.Post(fullPath, ro)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API Returned error: %v", resp.String())
		}

		dbResp := APIResponse{}
		if err := resp.JSON(&dbResp); err != nil {
			return err
		}

		if dbResp.Status != "OK" {
			return fmt.Errorf("API request completed, but with error: %v", dbResp.Message)
		}

		// Update pol with its ID before adding it to the existing policies list.
		pol.MID = bson.ObjectId(dbResp.Meta)

		// Add created Policy to existing policies
		mids[pol.MID.Hex()] = &pol
		ids[pol.ID] = &pol

		fmt.Printf("--> Status: OK, ID:%v\n", dbResp.Meta)
	}

	return nil
}

func (c *Client) DeletePolicy(id string) error {
	fullPath := urljoin.Join(c.url, endpointPolicies, id)

	ro := &grequests.RequestOptions{
		Headers: map[string]string{
			"Authorization": c.secret,
		},
		InsecureSkipVerify: c.InsecureSkipVerify,
	}

	resp, err := grequests.Delete(fullPath, ro)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("API Returned error: %v", resp.String())
	}

	return nil
}

func (c *Client) FetchPolicy(id string) (*objects.Policy, error) {
	fullPath := urljoin.Join(c.url, endpointPolicies, id)

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
		return nil, fmt.Errorf("API Returned error: %v", resp.String())
	}

	pol := objects.Policy{}
	if err := resp.JSON(&pol); err != nil {
		return nil, err
	}

	return &pol, nil
}

func (c *Client) UpdatePolicies(pols *[]objects.Policy) error {
	existingPols, err := c.FetchPolicies()
	if err != nil {
		return err
	}

	mids, ids := getPoliciesIdentifiers(&existingPols)

	for i := range *pols {
		pol := (*pols)[i]
		fmt.Printf("Updating Policy %v: %v\n", i, pol.Name)
		if pol.MID.Hex() == "" && pol.ID == "" {
			return errors.New("--> Can't update policy without an ID or explicit (legacy) ID")
		}

		if nil != ids[pol.ID] {
			fmt.Println("--> Found policy using explicit ID, substituting remote ID for update")

			pol.MID = ids[pol.ID].MID
		} else if nil == mids[pol.MID.Hex()] {
			return UseCreateError
		}

		fullPath := urljoin.Join(c.url, endpointPolicies, pol.MID.Hex())

		ro := &grequests.RequestOptions{
			JSON: pol,
			Headers: map[string]string{
				"Authorization": c.secret,
			},
			InsecureSkipVerify: c.InsecureSkipVerify,
		}

		resp, err := grequests.Put(fullPath, ro)
		if err != nil {
			return err
		}

		if resp.StatusCode != 200 {
			return fmt.Errorf("API Returned error: %v", resp.String())
		}

		dbResp := APIResponse{}
		if err := resp.JSON(&dbResp); err != nil {
			return err
		}

		if dbResp.Status != "OK" {
			return fmt.Errorf("API request completed, but with error: %v", dbResp.Message)
		}

		// Add updated Policy to existing policies
		mids[pol.MID.Hex()] = &pol
		ids[pol.ID] = &pol

		fmt.Printf("--> Status: OK, ID:%v\n", dbResp.Meta)
	}

	return nil
}

func (c *Client) SyncPolicies(pols []objects.Policy) error {
	deletePols := []string{}
	updatePols := []objects.Policy{}
	createPols := []objects.Policy{}

	// Fetch the running Policy list

	ePols, err := c.FetchPolicies()
	if err != nil {
		return err
	}

	DashIDMap := map[string]int{}
	GitIDMap := map[string]int{}

	// Build the dash ID map
	for i, pol := range ePols {
		// Lets get a full list of existing IDs
		if pol.ID != "" {
			DashIDMap[pol.ID] = i
		} else {
			DashIDMap[pol.MID.Hex()] = i
		}
	}

	// Build the Git ID Map
	for i, pol := range pols {
		if pol.ID != "" {
			GitIDMap[pol.ID] = i
		} else if pol.MID.Hex() != "" {
			GitIDMap[pol.MID.Hex()] = i
		} else {
			created := fmt.Sprintf("temp-pol-%v", uuid.NewV4().String())
			GitIDMap[created] = i
		}
	}

	// Updates are when we find items in git that are also in dash
	for key, index := range GitIDMap {
		dashIndex, ok := DashIDMap[key]
		if ok {
			p := pols[index]
			// Make sure we target the correct DB ID
			p.MID = ePols[dashIndex].MID
			updatePols = append(updatePols, p)
		}
	}

	// Deletes are when we find items in the dash that are not in git
	for key, i := range DashIDMap {
		_, ok := GitIDMap[key]
		if !ok {
			deletePols = append(deletePols, ePols[i].MID.Hex())
		}
	}

	// Create operations are when we find things in Git that are not in the dashboard
	for key, index := range GitIDMap {
		_, ok := DashIDMap[key]
		if !ok {
			createPols = append(createPols, pols[index])
		}
	}

	fmt.Printf("Deleting policies: %v\n", len(deletePols))
	fmt.Printf("Updating policies: %v\n", len(updatePols))
	fmt.Printf("Creating policies: %v\n", len(createPols))

	// Do the deletes
	for _, dbId := range deletePols {
		fmt.Printf("SYNC Deleting Policy: %v\n", dbId)
		if err := c.DeletePolicy(dbId); err != nil {
			return err
		}
	}

	// Do the updates
	if err := c.UpdatePolicies(&updatePols); err != nil {
		return err
	}
	for _, pol := range updatePols {
		fmt.Printf("SYNC Updated Policy: %v\n", pol.Name)
	}

	// Do the creates
	if err = c.CreatePolicies(&createPols); err != nil {
		return err
	}
	for _, pol := range createPols {
		fmt.Printf("SYNC Created Policy: %v\n", pol.Name)
	}

	return nil
}
