package objects

import (
	"github.com/TykTechnologies/tyk/apidef"
	"gopkg.in/mgo.v2/bson"
)

func NewDefinition() *DBApiDefinition {
	return &DBApiDefinition{}
}

type DBApiDefinition struct {
	apidef.APIDefinition `bson:"api_definition,inline" json:"api_definition,inline"`
	HookReferences       []interface{} `bson:"hook_references" json:"hook_references"`
	IsSite               bool          `bson:"is_site" json:"is_site"`
	SortBy               int           `bson:"sort_by" json:"sort_by"`
	UserGroupOwners      []bson.ObjectId `bson:"user_group_owners" json:"user_group_owners"`
	UserOwners           []bson.ObjectId `bson:"user_owners" json:"user_owners"`
}
