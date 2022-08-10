package tyk_vcs

import (
	"github.com/dmayo3/tyk-sync/clients/objects"
)

type Publisher interface {
	Name() string
	Create(apiDef *objects.DBApiDefinition) (string, error)
	Update(apiDef *objects.DBApiDefinition) error
	Delete(id string) error
	Sync(apiDefs []objects.DBApiDefinition) error
	CreatePolicy(*objects.Policy) (string, error)
	UpdatePolicy(*objects.Policy) error
	DeletePolicy(id string) error
	SyncPolicies([]objects.Policy) error
	Reload() error
}
