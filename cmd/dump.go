package cmd

import (
	"fmt"

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
		// Get secret from flags
		secret, _ := cmd.Flags().GetString("secret")
		if secret == "" {
			// If not set, try env var
			secret = os.Getenv("TYKGIT_DB_SECRET")
			if secret == "" {
				fmt.Println("Please set TYKGIT_DB_SECRET, or set the --secret flag, to your dashboard user secret")
				return
			}
		}

		fmt.Printf("Extracting APIs and Policies from %v\n", dbString)

		c, err := dashboard.NewDashboardClient(dbString, secret, "")
		if err != nil {
			fmt.Println(err)
		}

		wantedPoliciesByID, _ := cmd.Flags().GetStringSlice("policies")
		wantedAPIsByID, _ := cmd.Flags().GetStringSlice("apis")
		wantedTags, _ := cmd.Flags().GetStringSlice("tags")
		wantedCategories, _ := cmd.Flags().GetStringSlice("categories")

		fmt.Println("> Fetching policies")
		totalPolicies, err := c.FetchPolicies()
		if err != nil {
			fmt.Println("Error fetching policies:", err)
			return
		}
		fmt.Printf("--> Identified %v policies\n", len(totalPolicies))
		fmt.Println("> Fetching APIs")
		totalApis, err := c.FetchAPIs()
		if err != nil {
			fmt.Println("Error fetching apis:", err)
			return
		}
		fmt.Printf("--> Identified %v APIs\n", len(totalApis))

		if len(totalPolicies) == 0 && len(totalApis) == 0 {
			fmt.Println("No policies or APIs found")
			return
		}

		// Let's first filter all the policies.
		var filteredPolicies []objects.Policy
		if len(wantedPoliciesByID) > 0 {
			fmt.Println("--> Filtering policies by ID")
			filteredPolicies, err = helpers.GetPoliciesByID(totalPolicies, wantedPoliciesByID)
			if err != nil {
				fmt.Println("Error filtering Policies by ID:", err)
				return
			}
		}

		// Let's filter all the APIs.
		var filteredApis []objects.DBApiDefinition
		if len(wantedAPIsByID) > 0 {
			fmt.Println("--> Filtering APIs by ID")
			filteredApis, err = helpers.GetApisByID(totalApis, wantedAPIsByID)
			if err != nil {
				fmt.Println("Error filtering APIs by ID:", err)
				return
			}
		}

		// Let's filter APIs and Policies by tags.
		if len(wantedTags) > 0 {
			fmt.Println("--> Filtering APIs and Policies by tags")
			tempFilteredPoliciesByTag, tempFilteredApisByTag, err := helpers.LookForTags(totalPolicies, totalApis, wantedTags)
			if err != nil {
				fmt.Println("Error filtering APIs and Policies by tags:", err)
				return
			}

			filteredPolicies = append(filteredPolicies, tempFilteredPoliciesByTag...)
			filteredApis = append(filteredApis, tempFilteredApisByTag...)

		}

		// Let's filter Policies and APIs by categories.
		if len(wantedCategories) > 0 {
			fmt.Println("--> Filtering APIs and Policies by categories")
			tempFilteredApisByCategory, err := helpers.GetApisByCategory(totalApis, wantedCategories)
			if err != nil {
				fmt.Println("Error filtering APIs by categories:", err)
				return
			}

			filteredApis = append(filteredApis, tempFilteredApisByCategory...)
		}

		// Let's remove duplicates from the filtered policies.
		cleanPolicies := helpers.RemoveDuplicatesFromPolicies(filteredPolicies)
		cleanApis := helpers.RemoveDuplicatesFromApis(filteredApis)

		// Generate JSON files for APIs - we also check if an imported API has a non-imported Policy
		dir, _ := cmd.Flags().GetString("target")
		apiFiles, err := helpers.GenerateApiFiles(cleanApis, cleanPolicies, dir)
		if err != nil {
			fmt.Println("Error generating API files:", err)
			return
		}

		// Generate JSON files for Policies - we also check if an imported Policy has access over a non-imported API
		policyFiles, err := helpers.GeneratePolicyFiles(cleanPolicies, cleanApis, dir)
		if err != nil {
			fmt.Println("Error generating Policy files:", err)
			return
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
	dumpCmd.Flags().StringSlice("categories", []string{}, "Specific Apis categories to dump")
	dumpCmd.Flags().StringSlice("tags", []string{}, "Specific Apis and Policies tags to dump")
}
