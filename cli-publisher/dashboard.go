package cli_publisher

import (
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk-git/clients/dashboard"
)

type DashboardPublisher struct{
	Secret string
	Hostname string
}

func (p *DashboardPublisher) Create(apiDef *apidef.APIDefinition) (string, error) {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret)
	if err != nil {
		return "", err
	}

	return c.CreateAPI(apiDef)
}

func (p *DashboardPublisher) Update(id string, apiDef *apidef.APIDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.UpdateAPI(apiDef)
}

func (p *DashboardPublisher) Name() string {
	return "Dashboard Publisher"
}
