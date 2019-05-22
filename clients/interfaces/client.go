package interfaces

import (
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk/apidef"
)

type APIManagementClient interface {
	CreateAPI(def *apidef.APIDefinition) (string, error)
	FetchAPIs() ([]objects.DBApiDefinition, error)
	UpdateAPI(def *apidef.APIDefinition) error
	DeleteAPI(id string) error
}

type CertificateManagementClient interface {
	CreateCertificate(cert []byte) (string, error)
}

type UniversalClient interface {
	APIManagementClient
	CertificateManagementClient
	GetActiveID(def *apidef.APIDefinition) string
	SetInsecureTLS(bool)
}
