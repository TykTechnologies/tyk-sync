package helpers

import (
	"fmt"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func LookForTags(totalPolicies []objects.Policy, totalApis []objects.DBApiDefinition, wantedTags []string) ([]objects.Policy, []objects.DBApiDefinition, error) {
	var filteredPolicies []objects.Policy
	var filteredAPIs []objects.DBApiDefinition

	for _, wantedTag := range wantedTags {
		var tempPolicies []objects.Policy
		var tempAPIs []objects.DBApiDefinition
		fmt.Printf("--> Looking for policies with tag: %v\n", wantedTag)

		for _, pol := range totalPolicies {
			// If policy does not have tags, just skip it
			if len(pol.Tags) > 0 {
				for _, currentPolicyTag := range pol.Tags {
					if wantedTag == currentPolicyTag {
						tempPolicies = append(tempPolicies, pol)
					}
				}
			}
		}

		fmt.Printf("--> Looking for apis with tag: %v\n", wantedTag)

		for _, api := range totalApis {
			// If api does not have tags, just skip it
			if len(api.Tags) > 0 {
				for _, currentApiTag := range api.Tags {
					if wantedTag == currentApiTag {
						tempAPIs = append(tempAPIs, api)
					}
				}
			}
		}

		fmt.Printf("--> Found %v policies and %v apis with tag: %v\n", len(tempPolicies), len(tempAPIs), wantedTag)
		if len(tempPolicies) == 0 && len(tempAPIs) == 0 {
			return nil, nil, fmt.Errorf("no policies or apis found with the provided tag: %v", wantedTag)
		}

		filteredPolicies = append(filteredPolicies, tempPolicies...)
		filteredAPIs = append(filteredAPIs, tempAPIs...)
	}

	return filteredPolicies, filteredAPIs, nil

}
