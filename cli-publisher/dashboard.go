package cli_publisher

import (
	"fmt"

	"github.com/dmayo3/tyk-sync/clients/dashboard"
	"github.com/dmayo3/tyk-sync/clients/objects"
)

type DashboardPublisher struct {
	Secret      string
	Hostname    string
	OrgOverride string
}

func (p *DashboardPublisher) enforceOrgID(apiDef *objects.DBApiDefinition) *objects.DBApiDefinition {
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

func (p *DashboardPublisher) Create(apiDef *objects.DBApiDefinition) (string, error) {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	if err != nil {
		return "", err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	return c.CreateAPI(p.enforceOrgID(apiDef))
}

func (p *DashboardPublisher) Update(apiDef *objects.DBApiDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	return c.UpdateAPI(p.enforceOrgID(apiDef))
}

func (p *DashboardPublisher) Delete(id string) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	return c.DeleteAPI(id)
}

func (p *DashboardPublisher) Sync(apiDefs []objects.DBApiDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	if err != nil {
		return err
	}

	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	if p.OrgOverride != "" {
		fixedDefs := make([]objects.DBApiDefinition, len(apiDefs))
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
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	if err != nil {
		return "", err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}
	return c.CreatePolicy(p.enforceOrgIDForPolicy(pol))
}

func (p *DashboardPublisher) UpdatePolicy(pol *objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}
	return c.UpdatePolicy(p.enforceOrgIDForPolicy(pol))
}

func (p *DashboardPublisher) DeletePolicy(id string) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}
	return c.DeletePolicy(id)
}

func (p *DashboardPublisher) SyncPolicies(pols []objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
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
