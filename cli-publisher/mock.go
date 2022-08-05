package cli_publisher

import (
	"fmt"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

type MockPublisher struct{}

func (mp MockPublisher) Create(apiDef *objects.DBApiDefinition) (string, error) {
	newID := "654321"
	fmt.Printf("Creating API ID: %v (on: %v to: %v)\n",
		newID,
		apiDef.Proxy.ListenPath,
		apiDef.Proxy.TargetURL)
	return newID, nil
}

func (mp MockPublisher) Update(apiDef *objects.DBApiDefinition) error {
	fmt.Printf("Updating API ID: %v (on: %v to: %v)\n",
		apiDef.APIID,
		apiDef.Proxy.ListenPath,
		apiDef.Proxy.TargetURL)

	return nil
}

func (mp MockPublisher) Delete(id string) error {
	fmt.Printf("Deleting API ID: %v\n", id)
	return nil
}

func (mp MockPublisher) Sync(apiDef []objects.DBApiDefinition) error {
	return nil
}

func (mp MockPublisher) CreatePolicy(pol *objects.Policy) (string, error) {
	return "", nil
}

func (mp MockPublisher) UpdatePolicy(pol *objects.Policy) error {
	return nil
}

func (mp MockPublisher) DeletePolicy(id string) error {
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
