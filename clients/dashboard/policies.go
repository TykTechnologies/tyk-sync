package dashboard

import (
	"fmt"
	"github.com/TykTechnologies/tyk-git/clients/objects"
	"github.com/kataras/go-errors"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
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
		if ePol.MID.Hex() == pol.MID.Hex() {
			found = true
			break
		}

		if ePol.ID == pol.ID {
			fmt.Println("--> Found policy using explicit ID, substituting remote ID for update")
			pol.MID = ePol.MID
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
	return nil
}
