package objects

import "github.com/TykTechnologies/tyk/apidef"

func NewDefinition() *apidef.APIDefinition {
	return &apidef.APIDefinition{}
}

type DBApiDefinition struct {
	apidef.APIDefinition `bson:"api_definition,inline" json:"api_definition,inline"`
	HookReferences       []interface{} `bson:"hook_references" json:"hook_references"`
	IsSite               bool          `bson:"is_site" json:"is_site"`
	SortBy               int           `bson:"sort_by" json:"sort_by"`
}
