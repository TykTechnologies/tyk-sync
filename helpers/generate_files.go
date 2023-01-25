package helpers

import (
	"encoding/json"
	"fmt"
	"os"
	"path"

	"github.com/TykTechnologies/tyk-sync/clients/objects"
)

func GenerateApiFiles(cleanApis []objects.DBApiDefinition, cleanPolicies []objects.Policy, dir string) ([]string, error) {
	apiFiles := make([]string, len(cleanApis))
	for i, api := range cleanApis {
		j, jerr := json.MarshalIndent(api, "", "  ")
		if jerr != nil {
			return nil, jerr
		}

		fname := fmt.Sprintf("api-%v.json", api.APIID)
		p := path.Join(dir, fname)
		err := os.WriteFile(p, j, 0644)
		if err != nil {
			return nil, err
		}
		apiFiles[i] = fname

		// Check if the OIDC issuer policies are imported
		for _, provider := range api.OpenIDOptions.Providers {
			for _, id := range provider.ClientIDs {
				found := false
				for _, policy := range cleanPolicies {
					if policy.MID.Hex() == id {
						found = true
						break
					}
				}
				if !found {
					fmt.Println("--> [WARNING] Api ", api.APIID, " has the Policy ", id, " as an OIDC issuer policy and it isn't imported. It might cause some problems in the future.")
				}
			}
		}
	}
	return apiFiles, nil
}

func GeneratePolicyFiles(cleanPolicies []objects.Policy, cleanApis []objects.DBApiDefinition, dir string) ([]string, error) {
	policyFiles := make([]string, len(cleanPolicies))
	for i, pol := range cleanPolicies {
		if pol.ID == "" {
			pol.ID = pol.MID.Hex()
		}

		j, jerr := json.MarshalIndent(pol, "", "  ")
		if jerr != nil {
			return policyFiles, jerr
		}

		fname := fmt.Sprintf("policy-%v.json", pol.ID)
		p := path.Join(dir, fname)
		err := os.WriteFile(p, j, 0644)
		if err != nil {
			return policyFiles, err
		}
		policyFiles[i] = fname

		for _, accesRights := range pol.AccessRights {
			found := false
			for _, api := range cleanApis {
				if api.APIID == accesRights.APIID {
					found = true
				}
			}
			if !found {
				fmt.Println("--> [WARNING] policy ", pol.ID, " has access rights over API ID ", accesRights.APIID, " and that API is not imported. It might cause some problems in the future.")
			}
		}
	}
	return policyFiles, nil
}
