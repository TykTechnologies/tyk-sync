package tyk_git

import (
	"github.com/TykTechnologies/tyk/apidef"
)

type Publisher interface {
	Create(apiDef *apidef.APIDefinition) (string, error)
	Update(id string, apiDef *apidef.APIDefinition) error
}
