package helpers

import (
	"fmt"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func GetPoliciesByID(totalPolicies []objects.Policy, wantedPoliciesByID []string) ([]objects.Policy, error) {
	var filteredPolicies []objects.Policy

	for _, wantedPolicyID := range wantedPoliciesByID {
		var tempPolicies []objects.Policy
		fmt.Printf("--> Looking for policies with ID: %v\n", wantedPolicyID)

		for _, pol := range totalPolicies {
			if wantedPolicyID == pol.MID.Hex() {
				tempPolicies = append(tempPolicies, pol)
			}
		}

		if len(tempPolicies) == 0 {
			return nil, fmt.Errorf("no policies found with the provided ID: %v", wantedPolicyID)
		}
		fmt.Printf("--> Found %v policies with ID: %v\n", len(tempPolicies), wantedPolicyID)

		filteredPolicies = append(filteredPolicies, tempPolicies...)
	}

	return filteredPolicies, nil
}
