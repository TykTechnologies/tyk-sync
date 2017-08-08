package cli_publisher

import (
	"github.com/TykTechnologies/tyk/apidef"
	"fmt"
)

type MockPublisher struct{}

func (mp MockPublisher) Create(apiDef *apidef.APIDefinition) (string, error) {
	newID := "654321"
	fmt.Printf("Creating API ID: %v (on: %v to: %v)\n",
		newID,
		apiDef.Proxy.ListenPath,
		apiDef.Proxy.TargetURL)
	return newID, nil
}

func (mp MockPublisher) Update(id string, apiDef *apidef.APIDefinition) error {
	fmt.Printf("Updating API ID: %v (on: %v to: %v)\n",
		apiDef.APIID,
		apiDef.Proxy.ListenPath,
		apiDef.Proxy.TargetURL)

	return nil
}

func (mp MockPublisher) Name() string {
	return "Mock Publisher"
}