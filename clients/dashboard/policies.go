package dashboard

import (
	"fmt"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/kataras/go-errors"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
	"github.com/satori/go.uuid"
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

func (c *Client) CreatePolicy(pol *objects.Policy) (string, error) {
	existingPols, err := c.FetchPolicies()
	if err != nil {
		return "", err
	}

	for _, ePol := range existingPols {
		if ePol.MID.Hex() == pol.MID.Hex() {
			return "", UsePolUpdateError
		}

		if ePol.ID == pol.ID {
			return "", UsePolUpdateError
		}
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
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("API Returned error: %v", resp.String())
	}

	dbResp := APIResponse{}
	if err := resp.JSON(&dbResp); err != nil {
		return "", err
	}

	if dbResp.Status != "OK" {
		return "", fmt.Errorf("API request completed, but with error: %v", dbResp.Message)
	}

	return dbResp.Meta, nil
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

func (c *Client) UpdatePolicy(pol *objects.Policy) error {
	existingPols, err := c.FetchPolicies()
	if err != nil {
		return err
	}

	if pol.MID.Hex() == "" && pol.ID == "" {
		return errors.New("--> Can't update policy without an ID or explicit (legacy) ID")
	}

	found := false
	for _, ePol := range existingPols {
		if ePol.ID == pol.ID {
			fmt.Println("--> Found policy using explicit ID, substituting remote ID for update")
			pol.MID = ePol.MID
			found = true
			break
		}

		if ePol.MID.Hex() == pol.MID.Hex() {
			found = true
			break
		}
	}

	if !found {
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
	for _, pol := range updatePols {
		fmt.Printf("SYNC Updating Policy: %v\n", pol.Name)
		if err := c.UpdatePolicy(&pol); err != nil {
			return err
		}
	}

	// Do the creates
	for _, pol := range createPols {
		fmt.Printf("SYNC Creating Policy: %v\n", pol.Name)
		var err error
		var id string
		if id, err = c.CreatePolicy(&pol); err != nil {
			return err
		}
		intID := "new"
		if pol.ID != "" {
			intID = pol.ID
		}
		fmt.Printf("--> ID: %v (%v)\n", id, intID)
	}

	return nil
}
