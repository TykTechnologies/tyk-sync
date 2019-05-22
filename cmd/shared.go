package cmd

import (
	"errors"
	"fmt"
	"github.com/TykTechnologies/tyk-sync/cli-publisher"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	"github.com/TykTechnologies/tyk-sync/tyk-vcs"
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/spf13/cobra"
	"os"
)

var isGateway bool

func doGitFetchCycle(getter tyk_vcs.Getter) ([]apidef.APIDefinition, []objects.Policy, error) {
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
		//TODO: Set up auth
	}

	branch, _ := cmd.Flags().GetString("branch")
	return auth, branch
}

func NewGetter(cmd *cobra.Command, args []string) (tyk_vcs.Getter, error) {
	filePath, _ :=  cmd.Flags().GetString("path")
	if filePath != "" {
		return tyk_vcs.NewFSGetter(filePath)
	}

	if len(args) == 0 {
		return nil, errors.New("must specify repo address to pull from as first argument")
	}
	auth, branch := getAuthAndBranch(cmd, args)
	return tyk_vcs.NewGGetter(args[0], branch, auth)
}

func doGetData(cmd *cobra.Command, args []string) ([]apidef.APIDefinition, []objects.Policy, error) {
	getter, err := NewGetter(cmd, args)
	if err != nil {
		return nil, nil, err
	}

	defs, pols, err := doGitFetchCycle(getter)
	if err != nil {
		return nil, nil, err
	}

	return defs, pols, nil
}

func processSync(cmd *cobra.Command, args []string) error {
	defs, pols, err := doGetData(cmd, args)
	if err != nil {
		return err
	}

	fmt.Println("Processing APIs...")
	publisher, err := getPublisher(cmd, args)
	if err != nil {
		return err
	}
	fmt.Printf("Using publisher: %v\n", publisher.Name())

	if err := publisher.Sync(defs); err != nil {
		return err
	}

	if len(pols) > 0 {
		fmt.Println("Processing Policies...")
		if err := publisher.SyncPolicies(pols); err != nil {
			return err
		}
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

	for i, d := range defs {
		if cmd.Use == "publish" {
			fmt.Printf("Creating API %v: %v\n", i, d.Name)
			id, err := publisher.Create(&d)
			if err != nil {
				fmt.Printf("--> Status: FAIL, Error:%v\n", err)
			} else {
				fmt.Printf("--> Status: OK, ID:%v\n", id)
			}
		}

		if cmd.Use == "update" {
			fmt.Printf("Updating API %v: %v\n", i, d.Name)
			err := publisher.Update(&d)
			if err != nil {
				fmt.Printf("--> Status: FAIL, Error:%v\n", err)
			} else {
				fmt.Printf("--> Status: OK, ID:%v\n", d.APIID)
			}
		}
	}

	for i, d := range pols {
		if cmd.Use == "publish" {
			fmt.Printf("Creating Policy %v: %v\n", i, d.Name)
			id, err := publisher.CreatePolicy(&d)
			if err != nil {
				fmt.Printf("--> Status: FAIL, Error:%v\n", err)
			} else {
				fmt.Printf("--> Status: OK, ID:%v\n", id)
			}
		}

		if cmd.Use == "update" {
			fmt.Printf("Updating Policy %v: %v\n", i, d.Name)
			err := publisher.UpdatePolicy(&d)
			if err != nil {
				fmt.Printf("--> Status: FAIL, Error:%v\n", err)
			} else {
				fmt.Printf("--> Status: OK, ID:%v\n", d.Name)
			}
		}
	}

	if isGateway {
		if err := publisher.Reload(); err != nil {
			return err
		}
	}

	fmt.Println("Done")
	return nil
}
