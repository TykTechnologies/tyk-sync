package tyk_vcs

import (
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk/apidef"
)

type Publisher interface {
	Name() string
	Create(apiDef *apidef.APIDefinition) (string, error)
	Update(apiDef *apidef.APIDefinition) error
	Sync(apiDefs []apidef.APIDefinition) error
	CreatePolicy(*objects.Policy) (string, error)
	UpdatePolicy(*objects.Policy) error
	SyncPolicies([]objects.Policy) error
	Reload() error
}
