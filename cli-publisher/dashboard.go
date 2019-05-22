package cli_publisher

import (
	"fmt"
	"github.com/TykTechnologies/tyk-sync/clients/dashboard"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk/apidef"
)

type DashboardPublisher struct {
	Secret      string
	Hostname    string
	OrgOverride string
}

func (p *DashboardPublisher) enforceOrgID(apiDef *apidef.APIDefinition) *apidef.APIDefinition {
	if p.OrgOverride != "" {
		fmt.Println("org override detected, setting.")
		apiDef.OrgID = p.OrgOverride
	}

	return apiDef
}

func (p *DashboardPublisher) enforceOrgIDForPolicy(pol *objects.Policy) *objects.Policy {
	if p.OrgOverride != "" {
		fmt.Println("org override detected, setting.")
		pol.OrgID = p.OrgOverride
	}

	return pol
}

func (p *DashboardPublisher) Create(apiDef *apidef.APIDefinition) (string, error) {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret)
	if err != nil {
		return "", err
	}

	return c.CreateAPI(p.enforceOrgID(apiDef))
}

func (p *DashboardPublisher) Update(apiDef *apidef.APIDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.UpdateAPI(p.enforceOrgID(apiDef))
}

func (p *DashboardPublisher) Sync(apiDefs []apidef.APIDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	if p.OrgOverride != "" {
		fixedDefs := make([]apidef.APIDefinition, len(apiDefs))
		for i, a := range apiDefs {
			newDef := a
			newDef.OrgID = p.OrgOverride
			fixedDefs[i] = newDef
		}

		return c.Sync(fixedDefs)
	}

	return c.Sync(apiDefs)
}

func (p *DashboardPublisher) Reload() error {
	fmt.Println("Dashboard does not require explicit reload. Skipping Reload.")
	return nil
}

func (p *DashboardPublisher) Name() string {
	return "Dashboard Publisher"
}

func (p *DashboardPublisher) CreatePolicy(pol *objects.Policy) (string, error) {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret)
	if err != nil {
		return "", err
	}

	return c.CreatePolicy(pol)
}

func (p *DashboardPublisher) UpdatePolicy(pol *objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	return c.UpdatePolicy(pol)
}

func (p *DashboardPublisher) SyncPolicies(pols []objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret)
	if err != nil {
		return err
	}

	if p.OrgOverride != "" {
		fixedPols := make([]objects.Policy, len(pols))
		for i, pol := range pols {
			newPol := pol
			newPol.OrgID = p.OrgOverride
			fixedPols[i] = newPol
		}

		return c.SyncPolicies(fixedPols)
	}

	return c.SyncPolicies(pols)
}
