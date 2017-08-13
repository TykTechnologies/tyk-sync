package cli_publisher

import (
	"github.com/TykTechnologies/tyk-git/clients/gateway"
	"github.com/TykTechnologies/tyk-git/clients/objects"
	"github.com/TykTechnologies/tyk/apidef"
)

type GatewayPublisher struct {
	Secret   string
	Hostname string
}

func (p *GatewayPublisher) Create(apiDef *apidef.APIDefinition) (string, error) {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return "", err
	}

	return c.CreateAPI(apiDef)
}

func (p *GatewayPublisher) Update(apiDef *apidef.APIDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.UpdateAPI(apiDef)
}

func (p *GatewayPublisher) Name() string {
	return "Gateway Publisher"
}

func (p *GatewayPublisher) Reload() error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.Reload()
}

func (p *GatewayPublisher) Sync(apiDefs []apidef.APIDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.Sync(apiDefs)
}

func (p *GatewayPublisher) CreatePolicy(pol *objects.Policy) (string, error) {
	return "", nil
}

func (p *GatewayPublisher) UpdatePolicy(pol *objects.Policy) error {
	return nil
}

func (p *GatewayPublisher) SyncPolicies(pols []objects.Policy) error {
	return nil
}
