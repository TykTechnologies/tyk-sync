package objects

import (
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/apidef/oas"

	"gopkg.in/mgo.v2/bson"
)

func NewDefinition() *DBApiDefinition {
	return &DBApiDefinition{}
}

type DBApiDefinition struct {
	*APIDefinition  `bson:"api_definition" json:"api_definition"`
	OAS             *oas.OAS        `json:"oas,omitempty"`
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
