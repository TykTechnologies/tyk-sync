package cli_publisher

import (
	"fmt"

	"github.com/TykTechnologies/tyk-sync/clients/dashboard"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

type DashboardPublisher struct {
	Secret         string
	Hostname       string
	OrgOverride    string
	AllowUnsafeOAS bool
}

func (p *DashboardPublisher) enforceOrgID(apiDefs *[]objects.DBApiDefinition) {
	if p.OrgOverride != "" {
		fmt.Println("org override detected, setting.")

		for i := range *apiDefs {
			(*apiDefs)[i].OrgID = p.OrgOverride
		}
	}
}

func (p *DashboardPublisher) enforceOrgIDForPolicies(pols *[]objects.Policy) {
	if p.OrgOverride != "" {
		fmt.Println("org override detected, setting.")

		for i := range *pols {
			(*pols)[i].OrgID = p.OrgOverride
		}
	}
}

func (p *DashboardPublisher) enforceOrgIDForAssets(assetsDefs *[]objects.DBAssets) {
	if p.OrgOverride != "" {
		fmt.Println("org override detected, setting.")

		for i := range *assetsDefs {
			(*assetsDefs)[i].OrgID = p.OrgOverride
		}
	}
}

func (p *DashboardPublisher) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, p.AllowUnsafeOAS)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgID(apiDefs)
	return c.CreateAPIs(apiDefs)
}

func (p *DashboardPublisher) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, p.AllowUnsafeOAS)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgID(apiDefs)
	return c.UpdateAPIs(apiDefs)
}

func (p *DashboardPublisher) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, p.AllowUnsafeOAS)
	if err != nil {
		return err
	}

	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	if p.OrgOverride != "" {
		fixedDefs := make([]objects.DBApiDefinition, len(apiDefs))
		for i, a := range apiDefs {
			if a.APIDefinition == nil {
				continue
			}

			newDef := a
			newDef.SetOrgID(p.OrgOverride)
			fixedDefs[i] = newDef
		}

		return c.SyncAPIs(fixedDefs)
	}

	return c.SyncAPIs(apiDefs)
}

func (p *DashboardPublisher) Reload() error {
	fmt.Println("Dashboard does not require explicit reload. Skipping Reload.")
	return nil
}

func (p *DashboardPublisher) Name() string {
	return "Dashboard Publisher"
}

func (p *DashboardPublisher) CreatePolicies(pols *[]objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, p.AllowUnsafeOAS)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgIDForPolicies(pols)
	return c.CreatePolicies(pols)
}

func (p *DashboardPublisher) UpdatePolicies(pols *[]objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, p.AllowUnsafeOAS)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgIDForPolicies(pols)
	return c.UpdatePolicies(pols)
}

func (p *DashboardPublisher) SyncPolicies(pols []objects.Policy) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, p.AllowUnsafeOAS)
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

func (p *DashboardPublisher) CreateAssets(assets *[]objects.DBAssets) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, false)
	if err != nil {
		return err
	}

	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgIDForAssets(assets)

	return c.CreateAssets(assets)
}

func (p *DashboardPublisher) UpdateAssets(assetDefs *[]objects.DBAssets) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, false)
	if err != nil {
		return err
	}
	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	p.enforceOrgIDForAssets(assetDefs)
	return c.UpdateAssets(assetDefs)
}

func (p *DashboardPublisher) SyncAssets(assets []objects.DBAssets) error {
	c, err := dashboard.NewDashboardClient(p.Hostname, p.Secret, p.OrgOverride, false)
	if err != nil {
		return err
	}

	if p.OrgOverride == "" {
		p.OrgOverride = c.OrgID
	}

	if p.OrgOverride != "" {
		fixedDefs := make([]objects.DBAssets, len(assets))
		for i, a := range assets {
			newDef := a
			newDef.OrgID = p.OrgOverride
			fixedDefs[i] = newDef
		}

		return c.SyncAssets(fixedDefs)
	}

	return c.SyncAssets(assets)
}
