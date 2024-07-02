package dashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/levigross/grequests"
	"github.com/ongoingio/urljoin"
)

type Client struct {
	url                string
	secret             string
	isCloud            bool
	InsecureSkipVerify bool
	OrgID              string
	allowUnsafeOAS     bool
}

// CategoriesPayload is a struct that holds a list of categories.
type CategoriesPayload struct {
	Categories []string `json:"categories"`
}

func (c *Client) UpdateOASCategory(oasApi *objects.DBApiDefinition) (*grequests.Response, error) {
	if oasApi == nil {
		return nil, nil
	}

	if !oasApi.IsOASAPI() {
		return nil, fmt.Errorf("malformed input to update OAS API category")
	}

	if oasApi.OAS == nil || oasApi.OAS.GetTykExtension() == nil {
		return nil, fmt.Errorf("malformed input to update OAS API category, invalid OAS spec or x-tyk-api-gateway field")
	}

	data, err := json.Marshal(CategoriesPayload{Categories: oasApi.Categories})
	if err != nil {
		return nil, err
	}

	fullPath := urljoin.Join(c.url, endpointOASAPIs, oasApi.OAS.GetTykExtension().Info.ID, endpointCategories)

	putResp, err := grequests.Put(fullPath, &grequests.RequestOptions{
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
		return nil, err
	}

	if putResp.StatusCode != 200 {
		return nil, fmt.Errorf("OAS Categories API Returned error: %v (code: %v)", putResp.String(), putResp.StatusCode)
	}

	return putResp, nil
}

const (
	endpointAPIs       string = "/api/apis"
	endpointCategories string = "/categories"
	endpointOASAPIs    string = "/api/apis/oas"
	endpointPolicies   string = "/api/portal/policies"
	endpointCerts      string = "/api/certs"
	endpointUsers      string = "/api/users"
	endpointAssets     string = "/api/assets"
)

var (
	UseUpdateError      error = errors.New("Object seems to exist (same ID, API ID, Listen Path or Slug), use update()")
	UsePolUpdateError   error = errors.New("Object seems to exist (same ID, Explicit ID), use update()")
	UseCreateError      error = errors.New("Object does not exist, use create()")
	UseAssetUpdateError error = errors.New("Object seems to exist (same ID) use update()")
)

func NewDashboardClient(url, secret, orgID string, allowUnsafeOAS bool) (*Client, error) {
	client := &Client{
		url:            url,
		secret:         secret,
		isCloud:        strings.Contains(url, "tyk.io"),
		allowUnsafeOAS: allowUnsafeOAS,
	}

	if orgID == "" {
		fullPath := urljoin.Join(url, endpointUsers)

		ro := &grequests.RequestOptions{
			Params: map[string]string{"p": "-2"},
			Headers: map[string]string{
				"Authorization": secret,
			},
		}

		resp, err := grequests.Get(fullPath, ro)
		if err != nil {
			return client, err
		}

		if resp.StatusCode != 200 {
			return client, fmt.Errorf("Error getting users from dashboard: %v for %v", resp.String(), fullPath)
		}

		users := objects.UsersResponse{}
		if err := resp.JSON(&users); err != nil {
			return client, err
		}

		if len(users.Users) > 0 {
			client.OrgID = users.Users[0].OrgID
		}
	}

	return client, nil
}
