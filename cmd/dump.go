package cmd

import (
	"fmt"

	"encoding/json"
	"github.com/TykTechnologies/tyk-sync/clients/dashboard"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk-sync/tyk-vcs"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"path"
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

		c, err := dashboard.NewDashboardClient(dbString, secret)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("> Fetching policies")
		policies, err := c.FetchPolicies()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("--> Identified %v policies\n", len(policies))
		fmt.Println("--> Fetching and cleaning policy objects")
		// A bug exists which causes decoding of the access rights to break,
		// so we should fetch individually
		cleanPolicyObjects := make([]*objects.Policy, len(policies))
		for i, p := range policies {
			cp, err := c.FetchPolicy(p.MID.Hex())
			if err != nil {
				fmt.Println(err)
				return
			}

			// Make sure we retain IDs
			if cp.ID == "" {
				cp.ID = cp.MID.Hex()
			}

			cleanPolicyObjects[i] = cp
		}

		fmt.Println("> Fetching APIs")
		apis, err := c.FetchAPIs()
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Printf("--> Fetched %v APIs\n", len(apis))

		dir, _ := cmd.Flags().GetString("target")
		apiFiles := make([]string, len(apis))
		for i, api := range apis {
			j, jerr := json.MarshalIndent(api.APIDefinition, "", "  ")
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

		policyFiles := make([]string, len(cleanPolicyObjects))
		for i, pol := range cleanPolicyObjects {
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
}
