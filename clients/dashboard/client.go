package dashboard

import (
	"errors"
	"strings"
)

type Client struct {
	url                string
	secret             string
	isCloud            bool
	InsecureSkipVerify bool
}

const (
	endpointAPIs     string = "/api/apis"
	endpointPolicies string = "/api/portal/policies"
	endpointCerts    string = "/api/certs"
)

var (
	UseUpdateError    error = errors.New("Object seems to exist (same ID, API ID, Listen Path or Slug), use update()")
	UsePolUpdateError error = errors.New("Object seems to exist (same ID, Explicit ID), use update()")
	UseCreateError    error = errors.New("Object does not exist, use create()")
)

func NewDashboardClient(url, secret string) (*Client, error) {
	return &Client{
		url:     url,
		secret:  secret,
		isCloud: strings.Contains(url, "tyk.io"),
	}, nil
}
