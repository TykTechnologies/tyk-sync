package cli_publisher

import (
	"errors"

	"github.com/TykTechnologies/tyk-sync/clients/gateway"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

type GatewayPublisher struct {
	Secret   string
	Hostname string
}

func (p *GatewayPublisher) Create(apiDef *objects.DBApiDefinition) (string, error) {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return "", err
	}

	return c.CreateAPI(apiDef)
}

func (p *GatewayPublisher) Update(apiDef *objects.DBApiDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.UpdateAPI(apiDef)
}

func (p *GatewayPublisher) Delete(id string) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.DeleteAPI(id)
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

func (p *GatewayPublisher) Sync(apiDefs []objects.DBApiDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.Sync(apiDefs)
}

func (p *GatewayPublisher) CreatePolicy(pol *objects.Policy) (string, error) {
	return "", errors.New("Policy handling not supported by Gateway publisher")
}

func (p *GatewayPublisher) UpdatePolicy(pol *objects.Policy) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}

func (p *GatewayPublisher) DeletePolicy(id string) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}

func (p *GatewayPublisher) SyncPolicies(pols []objects.Policy) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}
