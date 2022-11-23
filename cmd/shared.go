package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/TykTechnologies/tyk-sync/cli-publisher"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk-sync/tyk-vcs"
	"github.com/spf13/cobra"
)

var isGateway bool

func doGitFetchCycle(getter tyk_vcs.Getter) ([]objects.DBApiDefinition, []objects.Policy, error) {
	err := getter.FetchRepo()
	if err != nil {
		return nil, nil, err
	}

	ts, err := getter.FetchTykSpec()
	if err != nil {
		return nil, nil, err
	}

	ads, err := getter.FetchAPIDef(ts)
	if err != nil {
		return nil, nil, err
	}

	pols, err := getter.FetchPolicies(ts)
	if err != nil {
		return nil, nil, err
	}

	return ads, pols, nil
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

		newDashPublisher := &cli_publisher.DashboardPublisher{
			Secret:      secret,
			Hostname:    dbString,
			OrgOverride: orgOverride,
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
		sshKey, errSsh := ioutil.ReadFile(keyFile)
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
	if filePath != "" {
		return tyk_vcs.NewFSGetter(filePath)
	}

	if len(args) == 0 {
		return nil, errors.New("must specify repo address to pull from as first argument")
	}
	auth, branch := getAuthAndBranch(cmd, args)
	return tyk_vcs.NewGGetter(args[0], branch, auth)
}

func doGetData(cmd *cobra.Command, args []string) ([]objects.DBApiDefinition, []objects.Policy, error) {

	getter, err := NewGetter(cmd, args)
	if err != nil {
		return nil, nil, err
	}

	defs, pols, err := doGitFetchCycle(getter)
	if err != nil {
		return nil, nil, err
	}

	wantedPolicies, _ := cmd.Flags().GetStringSlice("policies")
	wantedAPIs, _ := cmd.Flags().GetStringSlice("apis")

	if len(wantedAPIs) == 0 && len(wantedPolicies) == 0 {
		return defs, pols, nil
	}
	filteredAPIS := []objects.DBApiDefinition{}
	filteredPolicies := []objects.Policy{}

	if len(wantedAPIs) > 0 {
		filteredAPIS = defs[:]
		newL := 0
		for _, apiID := range wantedAPIs {
			for _, api := range filteredAPIS {
				if apiID != api.APIID {
					continue
				}
				filteredAPIS[newL] = api
				newL++
			}
		}
		filteredAPIS = filteredAPIS[:newL]
	}

	if len(wantedPolicies) > 0 {
		filteredPolicies = pols[:]
		newL := 0
		for _, polID := range wantedPolicies {
			for _, pol := range filteredPolicies {
				if !((polID == pol.ID) || (polID == pol.MID.Hex())) {
					continue
				}
				filteredPolicies[newL] = pol
				newL++
			}
		}
		filteredPolicies = filteredPolicies[:newL]
	}

	return filteredAPIS, filteredPolicies, nil
}

func processSync(cmd *cobra.Command, args []string) error {
	defs, pols, err := doGetData(cmd, args)
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

	if isGateway {
		if err := publisher.Reload(); err != nil {
			return err
		}
	}

	return nil
}

func processPublish(cmd *cobra.Command, args []string) error {
	defs, pols, err := doGetData(cmd, args)
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
