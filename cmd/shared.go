package cmd

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/spf13/cobra"

	cli_publisher "github.com/TykTechnologies/tyk-sync/cli-publisher"
	"github.com/TykTechnologies/tyk-sync/clients/examplesrepo"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	tyk_vcs "github.com/TykTechnologies/tyk-sync/tyk-vcs"
)

var isGateway bool

func doGitFetchCycle(getter tyk_vcs.Getter) ([]objects.DBApiDefinition, []objects.Policy, []objects.DBAssets, error) {
	err := getter.FetchRepo()
	if err != nil {
		return nil, nil, nil, err
	}

	ts, err := getter.FetchTykSpec()
	if err != nil {
		return nil, nil, nil, err
	}

	ads, err := getter.FetchAPIDef(ts)
	if err != nil {
		return nil, nil, nil, err
	}

	pols, err := getter.FetchPolicies(ts)
	if err != nil {
		return nil, nil, nil, err
	}

	assets, err := getter.FetchAssets(ts)
	if err != nil {
		return nil, nil, nil, err
	}

	return ads, pols, assets, nil
}

func getPublisher(cmd *cobra.Command, args []string) (tyk_vcs.Publisher, error) {
	mock, _ := cmd.Flags().GetBool("test")
	if mock {
		return cli_publisher.MockPublisher{}, nil
	}

	dbString, _ := cmd.Flags().GetString("dashboard")

	flagVal, _ := cmd.Flags().GetString("secret")
	if dbString != "" {
		sec := os.Getenv("TYKGIT_DB_SECRET")
		if sec == "" && flagVal == "" {
			return nil, errors.New("Please set TYKGIT_DB_SECRET, or set the --secret flag, to your dashboard user secret")
		}

		secret := ""
		if sec != "" {
			secret = sec
		}

		if flagVal != "" {
			secret = flagVal
		}

		orgOverride, _ := cmd.Flags().GetString("org")

		allowUnsafeOAS, _ := cmd.Flags().GetBool("allow-unsafe-oas")

		newDashPublisher := &cli_publisher.DashboardPublisher{
			Secret:         secret,
			Hostname:       dbString,
			OrgOverride:    orgOverride,
			AllowUnsafeOAS: allowUnsafeOAS,
		}

		return newDashPublisher, nil
	}

	gwString, _ := cmd.Flags().GetString("gateway")
	if gwString != "" {
		sec := os.Getenv("TYKGIT_GW_SECRET")
		if sec == "" && flagVal == "" {
			return nil, errors.New("Please set TYKGIT_GW_SECRET, or set the --secret flag, to your dashboard user secret")
		}

		secret := ""
		if sec != "" {
			secret = sec
		}

		if flagVal != "" {
			secret = flagVal
		}

		newGWPublisher := &cli_publisher.GatewayPublisher{
			Secret:   secret,
			Hostname: gwString,
		}

		isGateway = true
		return newGWPublisher, nil
	}

	return nil, errors.New("Publisher target not defined!")
}

func getAuthAndBranch(cmd *cobra.Command, args []string) ([]byte, string) {
	keyFile, _ := cmd.Flags().GetString("key")
	var auth []byte
	if keyFile != "" {
		sshKey, errSsh := os.ReadFile(keyFile)
		if errSsh != nil {
			fmt.Println("Error reading ", keyFile, " for github key:", errSsh)
		}
		auth = []byte(sshKey)
	}

	branch, _ := cmd.Flags().GetString("branch")
	return auth, branch
}

func NewGetter(cmd *cobra.Command, args []string) (tyk_vcs.Getter, error) {
	filePath, _ := cmd.Flags().GetString("path")
	subdirectoryPath, _ := cmd.Flags().GetString("location")
	if filePath != "" {
		return tyk_vcs.NewFSGetter(filePath, subdirectoryPath)
	}

	if len(args) == 0 {
		return nil, errors.New("must specify repo address to pull from as first argument")
	}
	auth, branch := getAuthAndBranch(cmd, args)
	return tyk_vcs.NewGGetter(args[0], branch, auth, subdirectoryPath)
}

func doGetData(cmd *cobra.Command, args []string) ([]objects.DBApiDefinition, []objects.Policy, []objects.DBAssets, error) {
	getter, err := NewGetter(cmd, args)
	if err != nil {
		return nil, nil, nil, err
	}

	defs, pols, assets, err := doGitFetchCycle(getter)
	if err != nil {
		return nil, nil, nil, err
	}

	wantedPolicies, _ := cmd.Flags().GetStringSlice("policies")
	wantedAPIs, _ := cmd.Flags().GetStringSlice("apis")
	wantedAssets, _ := cmd.Flags().GetStringSlice("templates")
	wantedOASAPIs, _ := cmd.Flags().GetStringSlice("oas-apis")

	if len(wantedAPIs) == 0 && len(wantedPolicies) == 0 && len(wantedAssets) == 0 && len(wantedOASAPIs) == 0 {
		return defs, pols, assets, nil
	}

	filteredAPIs := []objects.DBApiDefinition{}
	filteredPolicies := []objects.Policy{}
	filteredAssets := []objects.DBAssets{}

	if len(wantedAPIs) > 0 {
		for _, apiID := range wantedAPIs {
			for _, api := range defs {
				if apiID == api.GetAPIID() {
					filteredAPIs = append(filteredAPIs, api)
				}
			}
		}
	}

	if len(wantedOASAPIs) > 0 {
		fmt.Printf("--> Identified %v OAS APIs\n", len(wantedOASAPIs))

		idToDef := make(map[string]objects.DBApiDefinition)
		for _, def := range defs {
			idToDef[def.GetAPIID()] = def
		}

		// building the oas api def objs from wantedOASAPIs
		for _, APIID := range wantedOASAPIs {
			if def, exists := idToDef[APIID]; exists {
				filteredAPIs = append(filteredAPIs, def)
			}
		}
	}

	if len(wantedPolicies) > 0 {
		mapPolicyIDs := map[string]bool{}

		for _, id := range wantedPolicies {
			mapPolicyIDs[id] = true
		}

		for _, pol := range pols {
			if _, ok := mapPolicyIDs[pol.ID]; ok {
				filteredPolicies = append(filteredPolicies, pol)
			} else if _, ok := mapPolicyIDs[pol.MID.Hex()]; ok {
				filteredPolicies = append(filteredPolicies, pol)
			}
		}
	}

	if len(wantedAssets) > 0 {
		mapAssetIDs := map[string]bool{}

		for _, id := range wantedAssets {
			mapAssetIDs[id] = true
		}

		for _, asset := range assets {
			if _, ok := mapAssetIDs[asset.ID]; ok {
				filteredAssets = append(filteredAssets, asset)
			}
		}
	}

	return filteredAPIs, filteredPolicies, filteredAssets, nil
}

func processSync(cmd *cobra.Command, args []string) error {
	defs, pols, assets, err := doGetData(cmd, args)
	if err != nil {
		return err
	}

	publisher, err := getPublisher(cmd, args)
	if err != nil {
		return err
	}
	fmt.Printf("Using publisher: %v\n", publisher.Name())

	if len(pols) > 0 && !isGateway {
		fmt.Println("Processing Policies...")
		if err := publisher.SyncPolicies(pols); err != nil {
			return err
		}
	}

	fmt.Println("Processing APIs...")
	if err := publisher.SyncAPIs(defs); err != nil {
		return err
	}

	fmt.Println("Processing Assets...")
	if err := publisher.SyncAssets(assets); err != nil {
		return err
	}

	if isGateway {
		if err := publisher.Reload(); err != nil {
			return err
		}
	}

	return nil
}

func processPublish(cmd *cobra.Command, args []string) error {
	defs, pols, assets, err := doGetData(cmd, args)
	if err != nil {
		return err
	}

	publisher, err := getPublisher(cmd, args)
	if err != nil {
		return err
	}
	fmt.Printf("Using publisher: %v\n", publisher.Name())

	if "publish" == cmd.Use {
		err = publisher.CreateAPIs(&defs)
	} else if "update" == cmd.Use {
		err = publisher.UpdateAPIs(&defs)
	}

	if err != nil {
		fmt.Printf("--> Status: FAIL, Error:%v\n", err)
		fmt.Println("Failed to publish APIs")
		return err
	}

	if "publish" == cmd.Use {
		err = publisher.CreateAssets(&assets)
	} else if "update" == cmd.Use {
		err = publisher.UpdateAssets(&assets)
	}

	if err != nil {
		fmt.Printf("--> Status: FAIL, Error:%v\n", err)
		fmt.Println("Failed to publish Assets")
		return err
	}

	if !isGateway {
		if "publish" == cmd.Use {
			err = publisher.CreatePolicies(&pols)
		} else if "update" == cmd.Use {
			err = publisher.UpdatePolicies(&pols)
		}
	} else {
		err = publisher.Reload()
	}

	if err != nil {
		fmt.Printf("--> Status: FAIL, Error:%v\n", err)
		fmt.Println("Failed to publish Policies")
		return err
	}

	fmt.Println("Done")
	return nil
}

func processExamplesList() error {
	client, err := examplesrepo.NewExamplesClient(examplesrepo.RepoRootUrl)
	if err != nil {
		return err
	}

	examples, err := client.GetAllExamples()
	if err != nil {
		return err
	}

	if len(examples) == 0 {
		fmt.Println("no examples found")
		return nil
	}

	tabbedResultWriter := tabwriter.NewWriter(os.Stdout, 1, 1, 4, ' ', 0)
	_, err = fmt.Fprintln(tabbedResultWriter, "LOCATION\tNAME\tDESCRIPTION")
	if err != nil {
		return err
	}

	for _, example := range examples {
		_, err = fmt.Fprintf(tabbedResultWriter, "%s\t%s\t%s\n", example.Location, example.Name, example.Description)
		if err != nil {
			return err
		}
	}

	return tabbedResultWriter.Flush()
}

func processExamplePublish(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		args = append(args, examplesrepo.RepoGitUrl)
	}
	return processPublish(cmd, args)
}

func processExampleDetails(cmd *cobra.Command) error {
	location, err := cmd.Flags().GetString("location")
	if err != nil {
		return err
	}

	client, err := examplesrepo.NewExamplesClient(examplesrepo.RepoRootUrl)
	if err != nil {
		return err
	}

	examplesMap, err := client.GetAllExamplesAsLocationIndexedMap()
	if err != nil {
		return err
	}

	if len(examplesMap) == 0 {
		fmt.Println("no examples found")
		return nil
	}

	example, ok := examplesMap[location]
	if !ok {
		fmt.Printf("example with location '%s' could not be found", location)
		return nil
	}

	fmt.Println(generateExampleDetailsString(example))
	return nil
}

func generateExampleDetailsString(example examplesrepo.ExampleMetadata) string {
	featuresString := strings.Builder{}
	for i, feature := range example.Features {
		isLastItem := i == len(example.Features)-1
		var bulletPointFeature string
		if isLastItem {
			bulletPointFeature = fmt.Sprintf("- %s", feature)
		} else {
			bulletPointFeature = fmt.Sprintf("- %s\n", feature)
		}

		// string builder's Write always returns nil as err
		_, _ = featuresString.Write([]byte(bulletPointFeature))
	}
	return fmt.Sprintf(
		exampleDetailsTemplate,
		example.Location,
		example.Name,
		example.Description,
		featuresString.String(),
		example.MinTykVersion,
	)
}

var exampleDetailsTemplate = `LOCATION
%s

NAME
%s

DESCRIPTION
%s

FEATURES
%s

MIN TYK VERSION
%s`
