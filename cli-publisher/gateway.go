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

func (p *GatewayPublisher) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.CreateAPIs(apiDefs)
}

func (p *GatewayPublisher) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.UpdateAPIs(apiDefs)
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

func (p *GatewayPublisher) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.SyncAPIs(apiDefs)
}

func (p *GatewayPublisher) CreatePolicies(pols *[]objects.Policy) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}

func (p *GatewayPublisher) UpdatePolicies(pols *[]objects.Policy) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}

func (p *GatewayPublisher) SyncPolicies(pols []objects.Policy) error {
	return errors.New("Policy handling not supported by Gateway publisher")
}

func (p *GatewayPublisher) CreateAssets(apiDefs *[]objects.DBAssets) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.CreateAssets(apiDefs)
}

func (p *GatewayPublisher) SyncAssets(assetsDefs []objects.DBAssets) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.SyncAssets(&assetsDefs)
}

func (p *GatewayPublisher) UpdateAssets(assetsDefs *[]objects.DBAssets) error {
	c, err := gateway.NewGatewayClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.UpdateAssets(assetsDefs)
}
