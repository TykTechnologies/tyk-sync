package cli_publisher

import (
	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

type MockPublisher struct{}

func (mp MockPublisher) CreateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	//newID := "654321"
	//fmt.Printf("Creating API ID: %v (on: %v to: %v)\n",
	//	newID,
	//	apiDef.Proxy.ListenPath,
	//	apiDef.Proxy.TargetURL)
	return nil
}

func (mp MockPublisher) UpdateAPIs(apiDefs *[]objects.DBApiDefinition) error {
	//fmt.Printf("Updating API ID: %v (on: %v to: %v)\n",
	//	apiDef.APIID,
	//	apiDef.Proxy.ListenPath,
	//	apiDef.Proxy.TargetURL)

	return nil
}

func (mp MockPublisher) SyncAPIs(apiDefs []objects.DBApiDefinition) error {
	return nil
}

func (mp MockPublisher) CreatePolicies(pols *[]objects.Policy) error {
	return nil
}

func (mp MockPublisher) UpdatePolicies(pols *[]objects.Policy) error {
	return nil
}

func (mp MockPublisher) SyncPolicies(pols []objects.Policy) error {
	return nil
}

func (mp MockPublisher) Name() string {
	return "Mock Publisher"
}

func (mp MockPublisher) Reload() error {
	return nil
}
