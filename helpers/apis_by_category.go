package helpers

import (
	"fmt"
	"strings"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func GetApisByCategory(totalApis []objects.DBApiDefinition, wantedCategories []string) ([]objects.DBApiDefinition, error) {
	var filteredAPIs []objects.DBApiDefinition

	for _, wantedCategory := range wantedCategories {
		var tempAPIs []objects.DBApiDefinition
		fmt.Printf("--> Looking for APIs with category: %v", wantedCategory)

		for _, api := range totalApis {
			// Categories are stored in the API name, so we need to check if the wanted category is in the API name
			// with the following format: #category
			if strings.Contains(api.APIDefinition.Name, "#"+wantedCategory) {
				tempAPIs = append(tempAPIs, api)
			}
		}

		fmt.Printf("--> Found %v apis with category: %v", len(tempAPIs), wantedCategory)
		if len(tempAPIs) == 0 {
			return nil, fmt.Errorf("no APIs found with the provided category: %v", wantedCategory)
		}

		filteredAPIs = append(filteredAPIs, tempAPIs...)
	}

	return filteredAPIs, nil

}
