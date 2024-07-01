package objects

import (
	"github.com/TykTechnologies/storage/persistent/model"
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/apidef/oas"

	"gopkg.in/mgo.v2/bson"
)

func NewDefinition() *DBApiDefinition {
	return &DBApiDefinition{}
}

type DBApiDefinition struct {
	*APIDefinition  `bson:"api_definition" json:"api_definition,omitempty"`
	OAS             *oas.OAS        `json:"oas,omitempty"`
	Categories      []string        `bson:"categories" json:"categories,omitempty"`
	HookReferences  []interface{}   `bson:"hook_references" json:"hook_references"`
	IsSite          bool            `bson:"is_site" json:"is_site"`
	SortBy          int             `bson:"sort_by" json:"sort_by"`
	UserGroupOwners []bson.ObjectId `bson:"user_group_owners" json:"user_group_owners"`
	UserOwners      []bson.ObjectId `bson:"user_owners" json:"user_owners"`
}

type APIDefinition struct {
	apidef.APIDefinition
	Scopes                *apidef.Scopes                `json:"scopes,omitempty"`
	AnalyticsPluginConfig *apidef.AnalyticsPluginConfig `json:"analytics_plugin,omitempty"`
	ExternalOAuth         *apidef.ExternalOAuth         `json:"external_oauth,omitempty"`
}

func (d *DBApiDefinition) IsOASAPI() bool {
	if d.APIDefinition != nil && d.APIDefinition.IsOAS {
		return true
	}

	return false
}

func (d *DBApiDefinition) GetAPIName() string {
	if d.IsOASAPI() {
		if tykExt := d.OAS.GetTykExtension(); tykExt != nil {
			return tykExt.Info.Name
		}

		return ""
	}

	return d.Name
}

func (d *DBApiDefinition) GetAPIID() string {
	if d.IsOASAPI() {
		if d.OAS.GetTykExtension() != nil {
			return d.OAS.GetTykExtension().Info.ID
		}

		return ""
	}

	return d.APIID
}

func (d *DBApiDefinition) SetAPIID(apiID string) {
	if d.IsOASAPI() {
		if d.OAS.GetTykExtension() != nil {
			d.OAS.GetTykExtension().Info.ID = apiID
		}

		return
	}

	d.APIID = apiID
}

func (d *DBApiDefinition) GetListenPath() string {
	if d.IsOASAPI() {
		if d.OAS.GetTykExtension() != nil {
			return d.OAS.GetTykExtension().Server.ListenPath.Value
		}

		return ""
	}

	return d.Proxy.ListenPath
}

func (d *DBApiDefinition) GetDomain() string {
	if d.IsOASAPI() {
		if d.OAS.GetTykExtension() != nil && d.OAS.GetTykExtension().Server.CustomDomain != nil {
			return d.OAS.GetTykExtension().Server.CustomDomain.Name
		}
	}

	return d.Domain
}

func (d *DBApiDefinition) GetDBID() model.ObjectID {
	if d.IsOASAPI() {
		if d.OAS.GetTykExtension() != nil {
			return d.OAS.GetTykExtension().Info.DBID
		}
	}

	return d.Id
}

func (d *DBApiDefinition) SetDBID(id model.ObjectID) {
	if d.IsOASAPI() {
		if d.OAS.GetTykExtension() != nil {
			d.OAS.GetTykExtension().Info.DBID = id
		}

		return
	}

	d.Id = id
}

func (d *DBApiDefinition) SetOrgID(orgID string) {
	if d.IsOASAPI() {
		if d.OAS.GetTykExtension() != nil {
			d.OAS.GetTykExtension().Info.OrgID = orgID
		}

		return
	}

	d.OrgID = orgID
}
