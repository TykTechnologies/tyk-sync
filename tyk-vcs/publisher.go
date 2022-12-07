package tyk_vcs

import (
	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

type Publisher interface {
	Name() string
	CreateAPIs(apiDefs *[]objects.DBApiDefinition) error
	UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error
	SyncAPIs(apiDefs []objects.DBApiDefinition) error
	CreatePolicies(pols *[]objects.Policy) error
	UpdatePolicies(pols *[]objects.Policy) error
	SyncPolicies(pols []objects.Policy) error
	Reload() error
}
