package helpers

import (
	"fmt"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func GetApisByID(totalApis []objects.DBApiDefinition, wantedApisByID []string) ([]objects.DBApiDefinition, error) {
	var filteredAPIs []objects.DBApiDefinition

	for _, wantedApiID := range wantedApisByID {
		var tempAPIs []objects.DBApiDefinition
		fmt.Printf("--> Looking for apis with ID: %v\n", wantedApiID)

		for _, api := range totalApis {
			if wantedApiID == api.APIID {
				tempAPIs = append(tempAPIs, api)
			}
		}

		if len(tempAPIs) == 0 {
			return nil, fmt.Errorf("no APIs found with the provided ID: %v", wantedApiID)
		}
		fmt.Printf("--> Found %v APIs with ID: %v\n", len(tempAPIs), wantedApiID)

		filteredAPIs = append(filteredAPIs, tempAPIs...)
	}

	return filteredAPIs, nil
}
