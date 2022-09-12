package cmd

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"

	cli_publisher "github.com/TykTechnologies/tyk-sync/cli-publisher"
	"github.com/TykTechnologies/tyk-sync/clients/objects"
	tyk_vcs "github.com/TykTechnologies/tyk-sync/tyk-vcs"
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
	wantedTags, _ := cmd.Flags().GetStringSlice("tags")
	wantedCategories, _ := cmd.Flags().GetStringSlice("categories")

	//if no flags are set, we want to publish everything
	if len(wantedAPIs) == 0 && len(wantedPolicies) == 0 && len(wantedTags) == 0 && len(wantedCategories) == 0 {
		return defs, pols, nil
	}

	storedPoliciesIds := map[string]bool{}
	storedApisIds := map[string]bool{}

	//variables to keep track of the number of filtered items
	filteredAPIS := []objects.DBApiDefinition{}
	filteredPolicies := []objects.Policy{}

	//variables to support concurrency
	var wg sync.WaitGroup
	var l sync.Mutex

	//filter APIs by ID
	if len(wantedAPIs) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, apiID := range wantedAPIs {
				l.Lock()
				if storedApisIds[apiID] {
					l.Unlock()
					continue
				}
				l.Unlock()
				found := false
				for _, api := range defs {
					if apiID != api.APIID {
						continue
					}
					l.Lock()
					storedApisIds[apiID] = true
					filteredAPIS = append(filteredAPIS, api)
					l.Unlock()
					fmt.Println("--> Found API with ID: ", api.Id)
					found = true
				}
				if !found {
					fmt.Println("--> No API found with ID: ", apiID)
				}
			}

		}()
	}

	//filter policies by ID
	if len(wantedPolicies) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, polID := range wantedPolicies {
				l.Lock()
				if storedPoliciesIds[polID] {
					l.Unlock()
					continue
				}
				l.Unlock()
				found := false
				for _, pol := range pols {
					if !((polID == pol.ID) || (polID == pol.MID.Hex())) {
						continue
					}
					l.Lock()
					storedPoliciesIds[polID] = true
					filteredPolicies = append(filteredPolicies, pol)
					l.Unlock()
					fmt.Println("--> Found Policy with ID: ", pol.Name)
					found = true
					break
				}

				if !found {
					fmt.Println("--> No policiy found with ID:", polID)
				}
			}

		}()
	}

	//filter APIs and Policies by tags
	if len(wantedTags) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, tag := range wantedTags {
				found := false
				for _, api := range defs {
					for _, apiTag := range api.Tags {
						if tag == apiTag {
							l.Lock()
							//we can check this before creating the for loop, but users would never know
							//if the policies they sent were found
							if storedApisIds[api.APIID] {
								l.Unlock()
								found = true
								break
							}
							storedApisIds[api.APIID] = true
							filteredAPIS = append(filteredAPIS, api)
							l.Unlock()
							fmt.Println("--> Found API with ID: ", api.APIID)
							found = true
							break
						}
					}
				}

				if !found {
					fmt.Println("--> No API found with tag:", tag)
				}

				found = false
				for _, pol := range pols {
					for _, polTag := range pol.Tags {
						if tag == polTag {
							l.Lock()
							//we can check this before creating the for loop, but users would never know
							//if the policies they sent were found
							if storedPoliciesIds[pol.ID] {
								l.Unlock()
								found = true
								break
							}
							storedPoliciesIds[pol.ID] = true
							filteredPolicies = append(filteredPolicies, pol)
							l.Unlock()
							fmt.Println("--> Found Policy with ID: ", pol.ID)
							found = true
						}
					}
				}
				if !found {
					fmt.Println("--> No policy found with tag:", tag)
				}
			}
		}()

	}

	//filter APIs by categories
	if len(wantedCategories) > 0 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for _, category := range wantedCategories {
				found := false
				for _, api := range defs {
					//categories are stored in the api Name, after a '#' character
					if strings.Contains(api.Name, "#"+category) {
						l.Lock()
						if storedApisIds[api.APIID] {
							l.Unlock()
							found = true
							continue
						}
						storedApisIds[api.APIID] = true
						filteredAPIS = append(filteredAPIS, api)
						l.Unlock()
						fmt.Println("--> Found API with ID: ", api.APIID)
						found = true
					}
				}

				if !found {
					fmt.Println("--> No API found with category:", category)
				}

			}
		}()
	}

	wg.Wait()

	//if no apis or policies were found, return an error
	if len(filteredAPIS) == 0 && len(filteredPolicies) == 0 {
		return nil, nil, errors.New("no APIs or Policies found with the given filters")
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
	if err := publisher.Sync(defs); err != nil {
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

	if !isGateway {
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
	}

	if isGateway {
		if err := publisher.Reload(); err != nil {
			return err
		}
	}

	fmt.Println("Done")
	return nil
}
