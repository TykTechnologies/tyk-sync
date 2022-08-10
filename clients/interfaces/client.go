package interfaces

import (
	"github.com/dmayo3/tyk-sync/clients/objects"
)

type APIManagementClient interface {
	CreateAPI(def *objects.DBApiDefinition) (string, error)
	FetchAPIs() ([]objects.DBApiDefinition, error)
	UpdateAPI(def *objects.DBApiDefinition) error
	DeleteAPI(id string) error
}

type CertificateManagementClient interface {
	CreateCertificate(cert []byte) (string, error)
}

type UniversalClient interface {
	APIManagementClient
	CertificateManagementClient
	GetActiveID(def *objects.DBApiDefinition) string
	SetInsecureTLS(bool)
}
