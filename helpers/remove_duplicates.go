package helpers

import "github.com/TykTechnologies/tyk-sync/clients/objects"

func RemoveDuplicatesFromPolicies(policies []objects.Policy) []objects.Policy {
	keys := make(map[string]bool)
	list := []objects.Policy{}
	for _, entry := range policies {
		if _, value := keys[entry.MID.Hex()]; !value {
			keys[entry.MID.Hex()] = true
			list = append(list, entry)
		}
	}
	return list
}

func RemoveDuplicatesFromApis(apis []objects.DBApiDefinition) []objects.DBApiDefinition {
	keys := make(map[string]bool)
	list := []objects.DBApiDefinition{}
	for _, entry := range apis {
		if _, value := keys[entry.APIID]; !value {
			keys[entry.APIID] = true
			list = append(list, entry)
		}
	}
	return list
}
