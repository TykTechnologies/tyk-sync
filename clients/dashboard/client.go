package dashboard

import (
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
}

const (
	endpointAPIs     string = "/api/apis"
	endpointPolicies string = "/api/portal/policies"
	endpointCerts    string = "/api/certs"
	endpointUsers    string = "/api/users"
)

var (
	UseUpdateError    error = errors.New("Object seems to exist (same ID, API ID, Listen Path or Slug), use update()")
	UsePolUpdateError error = errors.New("Object seems to exist (same ID, Explicit ID), use update()")
	UseCreateError    error = errors.New("Object does not exist, use create()")
)

func NewDashboardClient(url, secret, orgID string) (*Client, error) {
	client := &Client{
		url:     url,
		secret:  secret,
		isCloud: strings.Contains(url, "tyk.io"),
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
