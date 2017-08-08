package tyk_vcs

import (
	"github.com/TykTechnologies/tyk/apidef"
)

type Publisher interface {
	Name() string
	Create(apiDef *apidef.APIDefinition) (string, error)
	Update(id string, apiDef *apidef.APIDefinition) error
}
