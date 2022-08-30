package cmd

import (
	"fmt"
	"sync"

	"encoding/json"
	"io/ioutil"
	"os"
	"path"

	"github.com/TykTechnologies/tyk-sync/clients/dashboard"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk-sync/helpers"
	tyk_vcs "github.com/TykTechnologies/tyk-sync/tyk-vcs"
	"github.com/spf13/cobra"
)

// dumpCmd represents the dump command
var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump will extract policies and APIs from a target (dashboard)",
	Long: `Dump will extract policies and APIs from a target (dashboard) and
    place them in a directory of your choosing. It will also generate a spec file
    that can be used for sync.`,
	Run: func(cmd *cobra.Command, args []string) {
		dbString, _ := cmd.Flags().GetString("dashboard")

		if dbString == "" {
			fmt.Println("Dump requires a dashboard URL to be set")
			return
		}

		flagVal, _ := cmd.Flags().GetString("secret")

		sec := os.Getenv("TYKGIT_DB_SECRET")
		if sec == "" && flagVal == "" {
			fmt.Println("Please set TYKGIT_DB_SECRET, or set the --secret flag, to your dashboard user secret")
			return
		}

		secret := ""
		if sec != "" {
			secret = sec
		}

		if flagVal != "" {
			secret = flagVal
		}

		fmt.Printf("Extracting APIs and Policies from %v\n", dbString)

		c, err := dashboard.NewDashboardClient(dbString, secret, "")
		if err != nil {
			fmt.Println(err)
		}

		wantedAPIsIDs, _ := cmd.Flags().GetStringSlice("apis")
		wantedPoliciesIDs, _ := cmd.Flags().GetStringSlice("policies")
		wantedTags, _ := cmd.Flags().GetStringSlice("tags")
		wantedCategories, _ := cmd.Flags().GetStringSlice("categories")

		//maps that will keep track of already stored policies and apis (avoiding repeated apis and policies)
		storedPoliciesIds := map[string]bool{}
		storedAPIsIds := map[string]bool{}

		var errPoliciesFetch error
		var errApisFetch error

		//arrays that will store the full information about the policies and apis
		var policies []objects.Policy
		var apis []objects.DBApiDefinition

		//variables to add concurrency support
		var wg sync.WaitGroup
		var l sync.Mutex

		//list of collected errors (we may change this to a channel of type error)
		var errList []error
		//looking for apis by their respective IDs
		if len(wantedAPIsIDs) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fmt.Println("--> Fetching APIs objects by IDs")
				apisByIds, err := helpers.LookForApiIDs(wantedAPIsIDs, c, storedAPIsIds)
				if err != nil {
					l.Lock()
					errList = append(errList, err)
					l.Unlock()
					return
				}
				l.Lock()
				apis = append(apis, apisByIds...)
				l.Unlock()
			}()
		}

		//looking for policies by their respective IDs
		if len(wantedPoliciesIDs) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fmt.Println("--> Fetching and cleaning Policies objects by IDs")
				policiesByIds, err := helpers.LookForPoliciesIDs(wantedPoliciesIDs, c, storedPoliciesIds)
				if err != nil {
					l.Lock()
					errList = append(errList, err)
					l.Unlock()
					return
				}
				l.Lock()
				policies = append(policies, policiesByIds...)
				l.Unlock()
			}()
		}
		//looking for APIs policies by their respective tags
		if len(wantedTags) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fmt.Println("--> Fetching and cleaning APIs and Policies by tags")
				policiesByTags, apisByTags, err := helpers.LookForTags(wantedTags, c, storedPoliciesIds, storedAPIsIds)
				if err != nil {
					l.Lock()
					errList = append(errList, err)
					l.Unlock()
					return
				}
				l.Lock()
				policies = append(policies, policiesByTags...)
				apis = append(apis, apisByTags...)
				l.Unlock()
			}()
		}
		//looking for APIs by their respective categories
		if len(wantedCategories) > 0 {
			wg.Add(1)
			go func() {
				defer wg.Done()
				fmt.Println("--> Fetching and cleaning APIs and Policies by categories")
				apisByCategories, err := helpers.LookForCategories(wantedCategories, c, storedAPIsIds)
				if err != nil {
					l.Lock()
					errList = append(errList, err)
					l.Unlock()
					return
				}
				l.Lock()
				apis = append(apis, apisByCategories...)
				l.Unlock()
			}()
		}

		wg.Wait()

		//checking if there is any value in the error list
		if len(errList) > 0 {
			for _, err := range errList {
				fmt.Println(err)
			}
			return
		}

		//if no IDs, tags or categories were specified, fetch all the apis and policies
		if len(policies) == 0 && len(apis) == 0 {
			fmt.Println("> Fetching policies ")

			policies, errPoliciesFetch = c.FetchPolicies()
			if errPoliciesFetch != nil {
				fmt.Println(errPoliciesFetch)
				return
			}
			fmt.Println("> Fetching APIs")

			apis, errApisFetch = c.FetchAPIs()
			if err != nil {
				fmt.Println(errApisFetch)
				return
			}
		}

		fmt.Printf("--> Identified  and fetched %v APIs\n", len(apis))
		fmt.Printf("--> Identified and fetched %v Policies\n", len(policies))

		dir, _ := cmd.Flags().GetString("target")
		apiFiles := make([]string, len(apis))
		for i, api := range apis {

			j, jerr := json.MarshalIndent(api, "", "  ")
			if jerr != nil {
				fmt.Printf("JSON Encoding error: %v\n", jerr.Error())
				return
			}

			fname := fmt.Sprintf("api-%v.json", api.APIID)
			p := path.Join(dir, fname)
			err := ioutil.WriteFile(p, j, 0644)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				return
			}
			apiFiles[i] = fname
		}

		// If we have selected Policies specified we're going to check if we're importing all the necessary APIs
		if len(wantedPoliciesIDs) > 0 || len(wantedTags) > 0 {
			for _, policy := range policies {
				for _, accesRights := range policy.AccessRights {
					found := false
					for _, api := range apis {
						if api.APIID == accesRights.APIID {
							found = true
						}
					}
					if !found {
						fmt.Println("--> [WARNING] Policy ", policy.ID, " has access rights over API ID ", accesRights.APIID, " and that API it's not imported. It might cause some problems in the future.")
					}
				}
			}
		}
		// If we have selected APIs specified we're going to check if we're importing all the necessary policies
		if len(wantedAPIsIDs) > 0 || len(wantedTags) > 0 || len(wantedCategories) > 0 {
			//checking selected APIs  -  Policies integrity
			for _, api := range apis {
				for _, provider := range api.OpenIDOptions.Providers {
					for _, id := range provider.ClientIDs {
						found := false
						for _, policy := range policies {
							if policy.ID == id {
								found = true
								break
							}
						}
						if !found {
							fmt.Println("--> [WARNING] Api ", api.APIID, " has the Policy ", id, " as an OIDC issuer policy and that policy is not imported. It might cause some problems in the future.")
						}
					}
				}
			}
		}

		policyFiles := make([]string, len(policies))
		for i, pol := range policies {
			if pol.ID == "" {
				pol.ID = pol.MID.Hex()
			}

			j, jerr := json.MarshalIndent(pol, "", "  ")
			if jerr != nil {
				fmt.Printf("JSON Encoding error: %v\n", jerr.Error())
				return
			}

			fname := fmt.Sprintf("policy-%v.json", pol.ID)
			p := path.Join(dir, fname)
			err := ioutil.WriteFile(p, j, 0644)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				return
			}

			policyFiles[i] = fname
		}

		// Create a spec file
		gitSpec := tyk_vcs.TykSourceSpec{
			Type:     tyk_vcs.TYPE_APIDEF,
			Files:    make([]tyk_vcs.APIInfo, len(apiFiles)),
			Policies: make([]tyk_vcs.PolicyInfo, len(policyFiles)),
		}

		for i, apiFile := range apiFiles {
			asInfo := tyk_vcs.APIInfo{
				File: apiFile,
			}
			gitSpec.Files[i] = asInfo
		}

		for i, polFile := range policyFiles {
			asInfo := tyk_vcs.PolicyInfo{
				File: polFile,
			}
			gitSpec.Policies[i] = asInfo
		}

		fname := ".tyk.json"
		p := path.Join(dir, fname)
		fmt.Printf("> Creating spec file in: %v\n", p)
		j, jerr := json.MarshalIndent(gitSpec, "", "  ")
		if jerr != nil {
			fmt.Printf("JSON Encoding error: %v\n", jerr.Error())
			return
		}
		if err := ioutil.WriteFile(p, j, 0644); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return
		}
		fmt.Println("Done.")
	},
}

func init() {
	RootCmd.AddCommand(dumpCmd)

	dumpCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	dumpCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	dumpCmd.Flags().StringP("branch", "b", "refs/heads/master", "Branch to use (defaults to refs/heads/master)")
	dumpCmd.Flags().StringP("secret", "s", "", "Your API secret")
	dumpCmd.Flags().StringP("target", "t", "", "Target directory for files")
	dumpCmd.Flags().StringSlice("policies", []string{}, "Specific Policies ids to dump")
	dumpCmd.Flags().StringSlice("apis", []string{}, "Specific Apis ids to dump")
	dumpCmd.Flags().StringSlice("tags", []string{}, "Specific Tags to dump (includes Apis and Policies)")
	dumpCmd.Flags().StringSlice("categories", []string{}, "Specific Categories to dump (includes Apis and Policies)")
}
