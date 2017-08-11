package cmd

import (
	"errors"
	"fmt"
	"github.com/TykTechnologies/tyk-git/cli-publisher"
	"github.com/TykTechnologies/tyk-git/tyk-vcs"
	"github.com/TykTechnologies/tyk/apidef"
	"github.com/spf13/cobra"
	"os"
)

var isGateway bool

func doGitFetchCycle(getter *tyk_vcs.GitGetter) ([]apidef.APIDefinition, error) {
	err := getter.FetchRepo()
	if err != nil {
		return nil, err
	}

	ts, err := getter.FetchTykSpec()
	if err != nil {
		return nil, err
	}

	ads, err := getter.FetchAPIDef(ts)
	if err != nil {
		return nil, err
	}

	if len(ads) == 0 {
		return nil, err
	}

	return ads, nil
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

		newDashPublisher := &cli_publisher.DashboardPublisher{
			Secret:   secret,
			Hostname: dbString,
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

func doGetData(cmd *cobra.Command, args []string) (*tyk_vcs.GitGetter, []apidef.APIDefinition, error) {
	auth, branch := getAuthAndBranch(cmd, args)

	if len(args) == 0 {
		return nil, nil, errors.New("Must specify repo address to pull from as first argument")
	}

	publisher, err := getPublisher(cmd, args)
	if err != nil {
		return nil, nil, err
	}

	fmt.Printf("Using publisher: %v\n", publisher.Name())

	getter, err := tyk_vcs.NewGGetter(args[0], branch, auth, publisher)
	if err != nil {
		return nil, nil, err
	}

	defs, err := doGitFetchCycle(getter)
	if err != nil {
		return nil, nil, err
	}

	if err != nil {
		return nil, nil, err
	}

	return getter, defs, nil
}

func processSync(cmd *cobra.Command, args []string) error {
	getter, defs, err := doGetData(cmd, args)
	if err != nil {
		return err
	}

	if err := getter.Sync(defs); err != nil {
		return err
	}

	return nil
}

func processPublish(cmd *cobra.Command, args []string) error {
	getter, defs, err := doGetData(cmd, args)

	if err != nil {
		return err
	}

	for i, d := range defs {
		if cmd.Use == "publish" {
			fmt.Printf("Creating API %v: %v\n", i, d.Name)
			id, err := getter.Create(&d)
			if err != nil {
				fmt.Printf("--> Status: FAIL, Error:%v\n", err)
			} else {
				fmt.Printf("--> Status: OK, ID:%v\n", id)
			}
		}

		if cmd.Use == "update" {
			fmt.Printf("Updating API %v: %v\n", i, d.Name)
			err := getter.Update(d.APIID, &d)
			if err != nil {
				fmt.Printf("--> Status: FAIL, Error:%v\n", err)
			} else {
				fmt.Printf("--> Status: OK, ID:%v\n", d.APIID)
			}
		}
	}

	if isGateway {
		if err := getter.Reload(); err != nil {
			return err
		}
	}

	fmt.Println("Done")
	return nil
}
