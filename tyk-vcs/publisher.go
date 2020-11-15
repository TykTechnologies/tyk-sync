package tyk_vcs

import (
	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

type Publisher interface {
	Name() string
	Create(apiDef *objects.DBApiDefinition) (string, error)
	Update(apiDef *objects.DBApiDefinition) error
	Sync(apiDefs []objects.DBApiDefinition) error
	CreatePolicy(*objects.Policy) (string, error)
	UpdatePolicy(*objects.Policy) error
	SyncPolicies([]objects.Policy) error
	Reload() error
}
