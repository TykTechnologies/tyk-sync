package helpers

import (
	"fmt"

	"github.com/TykTechnologies/tyk-sync/clients/dashboard"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"gopkg.in/mgo.v2/bson"
)

//LookForCategories looks for apis that matches with the given categories.
func LookForCategories(wantedCategories []string, c *dashboard.Client, storedApisIds map[string]bool) ([]objects.DBApiDefinition, error) {
	var filteredAPIs []objects.DBApiDefinition
	for _, category := range wantedCategories {
		apis, err := (c.FetchAPIsByCategory(category))
		if err != nil {
			return nil, err
		}
		for _, api := range apis {
			if !storedApisIds[api.Id.String()] {
				filteredAPIs = append(filteredAPIs, api)
				storedApisIds[api.Id.String()] = true
			}
		}
	}
	if len(filteredAPIs) == 0 {
		return nil, fmt.Errorf("no apis found with the given categories: %v", wantedCategories)
	}
	return filteredAPIs, nil
}

//LookForApiIDs looks for all APIs that matches with the given API ID.
func LookForApiIDs(wantedAPIs []string, c *dashboard.Client, storedAPIsIds map[string]bool) ([]objects.DBApiDefinition, error) {
	var filteredAPIs []objects.DBApiDefinition
	for _, APIID := range wantedAPIs {
		//if we already saved this API ID, skip it
		if storedAPIsIds[APIID] {
			continue
		}

		fullAPI, err := c.FetchAPI(APIID)
		if err != nil {
			return nil, err
		}
		filteredAPIs = append(filteredAPIs, fullAPI)
		storedAPIsIds[APIID] = true
	}
	if len(filteredAPIs) == 0 {
		return nil, fmt.Errorf("no apis found with the given IDs: %v", wantedAPIs)
	}
	return filteredAPIs, nil
}

//LookForPoliciesIDs looks for all the policies that matches with the given Policy ID.
func LookForPoliciesIDs(wantedPolicies []string, c *dashboard.Client, storedPoliciesIds map[string]bool) ([]objects.Policy, error) {
	var policies []objects.Policy
	for _, wantedPolicy := range wantedPolicies {
		if !bson.IsObjectIdHex(wantedPolicy) {
			return nil, fmt.Errorf("invalid selected policy ID: %s", wantedPolicy)
		}
		//if we already saved this policy ID, skip it (we don't want duplicates)
		if storedPoliciesIds[wantedPolicy] {
			continue
		}

		// A bug exists which causes decoding of the access rights to break,
		// so we should fetch individually
		cp, err := c.FetchPolicy(wantedPolicy)
		if err != nil {
			return nil, err
		}

		// Make sure we retain IDs
		if cp.ID == "" {
			cp.ID = cp.MID.Hex()
		}

		policies = append(policies, *cp)
		storedPoliciesIds[wantedPolicy] = true
	}
	if len(policies) == 0 {
		return nil, fmt.Errorf("no policies found with the given IDs: %v", wantedPolicies)
	}
	return policies, nil
}

// LookForTags Looks for all apis and policies that matches with the given tags
func LookForTags(wantedTags []string, c *dashboard.Client, storedPoliciesIds map[string]bool, storedApisIds map[string]bool) ([]objects.Policy, []objects.DBApiDefinition, error) {
	var filteredPolicies []objects.Policy
	var filteredAPIs []objects.DBApiDefinition

	policies, err := c.FetchPolicies()
	if err != nil {
		return nil, nil, err
	}

	apis, err := c.FetchAPIs()
	if err != nil {
		fmt.Println(err)
		return nil, nil, err
	}

	for _, wantedTag := range wantedTags {

		fmt.Printf("--> Looking for policies with tag: %v\n", wantedTag)

		for _, pol := range policies {
			//omitting already stored policies and policies without tags
			if len(pol.Tags) > 0 && !storedPoliciesIds[pol.MID.String()] {
				for _, currentPolicyTag := range pol.Tags {
					if wantedTag == currentPolicyTag {
						filteredPolicies = append(filteredPolicies, pol)
						storedPoliciesIds[pol.MID.Hex()] = true
					}
				}
			}
		}

		fmt.Printf("--> Looking for apis with tag: %v\n", wantedTag)

		for _, api := range apis {
			//omitting already stored apis and apis without tags
			if len(api.Tags) > 0 && !storedApisIds[api.Id.String()] {
				for _, currentApiTag := range api.Tags {
					if wantedTag == currentApiTag {
						filteredAPIs = append(filteredAPIs, api)
						storedApisIds[api.Id.String()] = true
					}
				}
			}
		}
	}

	if len(filteredPolicies) == 0 && len(filteredAPIs) == 0 {
		return nil, nil, fmt.Errorf("no policies or apis found with the provided tags: %v", wantedTags)
	}

	return filteredPolicies, filteredAPIs, nil

}
