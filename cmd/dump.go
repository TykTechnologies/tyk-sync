package cmd

import (
	"encoding/json"
	"fmt"
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/TykTechnologies/tyk/apidef/oas"
	"os"
	"path"

	"github.com/spf13/cobra"
	"gopkg.in/mgo.v2/bson"

	"github.com/TykTechnologies/tyk-sync/clients/dashboard"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	tyk_vcs "github.com/TykTechnologies/tyk-sync/tyk-vcs"
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

		fmt.Printf("Extracting APIs, Policies, and Assets from %v\n", dbString)

		c, err := dashboard.NewDashboardClient(dbString, secret, "", false)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println("> Fetching policies")
		wantedPolicies, _ := cmd.Flags().GetStringSlice("policies")
		wantedAPIs, _ := cmd.Flags().GetStringSlice("apis")
		wantedOASAPIs, _ := cmd.Flags().GetStringSlice("oas-apis")
		wantedAssets, _ := cmd.Flags().GetStringSlice("templates")

		policies := []objects.Policy{}
		apis := []objects.DBApiDefinition{}
		assets := []objects.DBAssets{}

		var errPoliciesFetch error
		var errApisFetch error
		var errAssetsFetch error

		// building the api def objs from wantedAPIs
		for _, APIID := range wantedAPIs {
			api := objects.DBApiDefinition{APIDefinition: &objects.APIDefinition{}}
			api.APIID = APIID
			apis = append(apis, api)
		}

		// building the oas api def objs from wantedOASAPIs
		for _, APIID := range wantedOASAPIs {
			myOas := &oas.OAS{}
			myOas.SetTykExtension(&oas.XTykAPIGateway{Info: oas.Info{ID: APIID}})

			api := objects.DBApiDefinition{
				APIDefinition: &objects.APIDefinition{
					APIDefinition: apidef.APIDefinition{
						APIID: APIID,
						IsOAS: true,
					},
				},
				OAS: myOas,
			}

			apis = append(apis, api)
		}

		// building the policies obj from wantedAPIs
		for _, wantedPolicy := range wantedPolicies {
			if !bson.IsObjectIdHex(wantedPolicy) {
				fmt.Printf("Invalid selected policy ID: %s.\n", wantedPolicy)
				return
			}
			pol := objects.Policy{
				ID:  wantedPolicy,
				MID: bson.ObjectIdHex(wantedPolicy),
			}
			policies = append(policies, pol)
		}

		//building the assets object from wantedAssets
		for _, aID := range wantedAssets {
			asset := objects.DBAssets{}
			asset.ID = aID
			assets = append(assets, asset)
		}

		if len(apis) == 0 && len(wantedPolicies) == 0 {
			fmt.Println("> Fetching policies ")

			policies, errPoliciesFetch = c.FetchPolicies()
			if errPoliciesFetch != nil {
				fmt.Println(errPoliciesFetch)
				return
			}

			fmt.Println("> Fetching APIs")

			var resp *dashboard.APISResponse
			resp, errApisFetch = c.FetchAPIs()
			if err != nil {
				fmt.Println(errApisFetch)
				return
			}

			apis = resp.Apis
		}

		var oasApisDB []objects.DBApiDefinition
		apis, oasApisDB = extractOASApis(apis)
		if len(wantedAssets) == 0 {
			fmt.Println("> Fetching Asset(s)")

			assets, errAssetsFetch = c.FetchAssets()
			if err != nil {
				fmt.Println(errAssetsFetch)
				return
			}
		}

		fmt.Printf("--> Identified %v policies\n", len(policies))
		if len(wantedPolicies) > 0 {
			fmt.Println("--> Fetching and cleaning policy objects")
		} else {
			fmt.Println("--> Cleaning policy objects")
		}

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
		fmt.Printf("--> Fetched %v Policies\n", len(cleanPolicyObjects))

		if len(wantedAPIs) > 0 {
			fmt.Printf("--> Identified %v APIs\n", len(apis))
			fmt.Println("--> Fetching and cleaning APIs objects")

			for i, api := range apis {
				fullAPI, err := c.FetchAPI(api.APIID)
				if err != nil {
					fmt.Println(err)
					return
				}
				apis[i] = fullAPI
			}
		}

		if len(wantedOASAPIs) > 0 {
			fmt.Printf("--> Identified %v OAS APIs\n", len(wantedOASAPIs))

			for i, api := range oasApisDB {
				if api.IsOASAPI() {
					fullAPI, err := c.FetchOASAPI(api.GetAPIID())
					if err != nil {
						fmt.Println(err)
						return
					}

					if oasApisDB[i].OAS == nil {
						oasApisDB[i].OAS = new(oas.OAS)
					}

					oasApisDB[i].OAS = fullAPI
				}
			}
		}

		fmt.Printf("--> Fetched %v Classic APIs\n", len(apis))
		fmt.Printf("--> Fetched %v OAS APIs\n", len(oasApisDB))
		fmt.Printf("--> Fetched %v Assets\n", len(assets))

		if len(wantedAssets) > 0 {
			fmt.Println("> Fetching Asset(s)")
			fmt.Println("--> Fetching and cleaning Assets objects")
			fmt.Printf("--> Identified %v Assets\n", len(assets))

			for i, asset := range assets {
				fullAsset, err := c.FetchAsset(asset.ID)
				if err != nil {
					fmt.Println(err)
					return
				}

				assets[i] = fullAsset
			}
		}

		fmt.Printf("--> Fetched %v Apis\n", len(apis))

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
			err := os.WriteFile(p, j, 0o644)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				return
			}
			apiFiles[i] = fname
		}

		assetFiles := make([]string, len(assets))
		for i, asset := range assets {
			j, jerr := json.MarshalIndent(asset, "", "  ")
			if jerr != nil {
				fmt.Printf("JSON Encoding error: %v\n", jerr.Error())
				return
			}

			fname := fmt.Sprintf("asset-%v.json", asset.ID)

			p := path.Join(dir, fname)
			err := os.WriteFile(p, j, 0o644)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				return
			}

			assetFiles[i] = fname
		}

		oasApiFiles := make([]string, len(oasApisDB))
		for i, oasApi := range oasApisDB {
			oasApi.APIDefinition = nil

			j, jerr := json.MarshalIndent(oasApi, "", "  ")
			if jerr != nil {
				fmt.Printf("OASAPI JSON Encoding error: %v\n", jerr.Error())
				return
			}

			name := ""
			if oasApi.OAS.GetTykExtension() != nil {
				name = oasApi.OAS.GetTykExtension().Info.ID
			}

			fname := fmt.Sprintf("oas-%v.json", name)
			p := path.Join(dir, fname)
			err := os.WriteFile(p, j, 0o644)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				return
			}

			oasApiFiles[i] = fname
		}

		// If we have selected Policies specified we're going to check if we're importing all the necessary APIs
		if len(wantedPolicies) > 0 {
			for _, policy := range cleanPolicyObjects {
				for _, accessRights := range policy.AccessRights {
					found := false
					for _, api := range apis {
						if api.APIID == accessRights.APIID {
							found = true
						}
					}

					for _, api := range oasApisDB {
						if api.APIID == accessRights.APIID {
							found = true
						}
					}

					if !found {
						fmt.Println("--> [WARNING] Policy ", policy.ID,
							" has access rights over API ID ", accessRights.APIID,
							" and that API it's not imported. It might cause some problems in the future.")
					}
				}
			}
		}

		// If we have selected APIs specified we're going to check if we're importing all the necessary policies
		if len(wantedAPIs) > 0 {
			// checking selected APIs  -  Policies integrity
			for _, api := range apis {
				for _, provider := range api.OpenIDOptions.Providers {
					for _, id := range provider.ClientIDs {
						found := false
						for _, policy := range cleanPolicyObjects {
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
			err := os.WriteFile(p, j, 0o644)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				return
			}

			policyFiles[i] = fname
		}

		// Create a spec file
		gitSpec := tyk_vcs.TykSourceSpec{
			Type:     tyk_vcs.TYPE_APIDEF,
			Policies: make([]tyk_vcs.PolicyInfo, len(policyFiles)),
			Assets:   make([]tyk_vcs.AssetsInfo, len(assetFiles)),
		}

		for _, apiFile := range apiFiles {
			asInfo := tyk_vcs.APIInfo{
				File: apiFile,
			}

			gitSpec.Files = append(gitSpec.Files, asInfo)
		}

		for i, polFile := range policyFiles {
			asInfo := tyk_vcs.PolicyInfo{
				File: polFile,
			}

			gitSpec.Policies[i] = asInfo
		}

		for _, oasApiFile := range oasApiFiles {
			asInfo := tyk_vcs.APIInfo{
				File: oasApiFile,
			}

			gitSpec.Files = append(gitSpec.Files, asInfo)
		}

		for i, assetFile := range assetFiles {
			asInfo := tyk_vcs.AssetsInfo{
				File: assetFile,
			}

			gitSpec.Assets[i] = asInfo
		}

		fname := ".tyk.json"
		p := path.Join(dir, fname)

		fmt.Printf("> Creating spec file in: %v\n", p)

		j, jerr := json.MarshalIndent(gitSpec, "", "  ")
		if jerr != nil {
			fmt.Printf("JSON Encoding error: %v\n", jerr.Error())
			return
		}

		if err := os.WriteFile(p, j, 0644); err != nil {
			fmt.Printf("Error writing file: %v\n", err)
			return
		}

		fmt.Println("Done.")
	},
}

// extractOASApis extracts OAS APIs from the array of API Definition objects.
// Each object in the array corresponds to the API Definition representation stored in database.
func extractOASApis(apis []objects.DBApiDefinition) (classic, oas []objects.DBApiDefinition) {
	for i := 0; i < len(apis); i++ {
		if apis[i].IsOAS {
			oas = append(oas, apis[i])
		} else {
			classic = append(classic, apis[i])
		}
	}

	return
}

func init() {
	RootCmd.AddCommand(dumpCmd)

	dumpCmd.Flags().StringP("dashboard", "d", "", "Fully qualified dashboard target URL")
	dumpCmd.Flags().StringP("key", "k", "", "Key file location for auth (optional)")
	dumpCmd.Flags().StringP("branch", "b", "refs/heads/master", "Branch to use (defaults to refs/heads/master)")
	dumpCmd.Flags().StringP("secret", "s", "", "Your API secret")
	dumpCmd.Flags().StringP("target", "t", "", "Target directory for files")
	dumpCmd.Flags().StringSlice("templates", []string{}, "List of template IDs to be dumped")
	dumpCmd.Flags().StringSlice("policies", []string{}, "Specific Policies IDs to dump")
	dumpCmd.Flags().StringSlice("apis", []string{}, "Specific Classic API Definition IDs to dump")
	dumpCmd.Flags().StringSlice("oas-apis", []string{}, "Specific OAS API Definition IDs to dump")
}
