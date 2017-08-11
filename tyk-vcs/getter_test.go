package tyk_vcs

import (
	"fmt"
	"github.com/TykTechnologies/tyk/apidef"
	"testing"
)

var REPO string = "https://github.com/lonelycode/integration-test.git"

type MockPublisher struct{}

var mockPublish MockPublisher = MockPublisher{}

func (mp MockPublisher) Create(apiDef *apidef.APIDefinition) (string, error) {
	newID := "654321"
	fmt.Printf("Creating API ID: %v (on: %v to: %v)\n",
		newID,
		apiDef.Proxy.ListenPath,
		apiDef.Proxy.TargetURL)
	return newID, nil
}

func (mp MockPublisher) Update(apiDef *apidef.APIDefinition) error {
	fmt.Printf("Updating API ID: %v (on: %v to: %v)\n",
		apiDef.APIID,
		apiDef.Proxy.ListenPath,
		apiDef.Proxy.TargetURL)

	return nil
}

func (mp MockPublisher) Name() string {
	return "Mock Publisher"
}

func (mp MockPublisher) Reload() error {
	return nil
}

func (mp MockPublisher) Sync(defs []apidef.APIDefinition) error {
	return nil
}

func TestNewGGetter(t *testing.T) {
	_, e := NewGGetter(REPO, "refs/heads/master", []byte{}, mockPublish)
	if e != nil {
		t.Fatal(e)
	}
}

func TestGitGetter_FetchRepo(t *testing.T) {
	g, e := NewGGetter(REPO, "refs/heads/master", []byte{}, mockPublish)
	if e != nil {
		t.Fatal(e)
	}

	e = g.FetchRepo()
	if e != nil {
		t.Fatal(e)
	}
}

func TestGitGetter_FetchTykSpec(t *testing.T) {
	g, e := NewGGetter(REPO, "refs/heads/master", []byte{}, mockPublish)
	if e != nil {
		t.Fatal(e)
	}

	e = g.FetchRepo()
	if e != nil {
		t.Fatal(e)
	}

	ts, err := g.FetchTykSpec()
	if err != nil {
		t.Fatal(err)
	}

	if ts.Type != TYPE_APIDEF {
		t.Fatalf("Spec Type is invalid: %v expected: '%v'", ts.Type, TYPE_APIDEF)
	}
}

func TestGitGetter_FetchAPIDef(t *testing.T) {
	g, e := NewGGetter(REPO, "refs/heads/master", []byte{}, mockPublish)
	if e != nil {
		t.Fatal(e)
	}

	e = g.FetchRepo()
	if e != nil {
		t.Fatal(e)
	}

	ts, err := g.FetchTykSpec()
	if err != nil {
		t.Fatal(err)
	}

	ads, err := g.FetchAPIDef(ts)
	if err != nil {
		t.Fatal(err)
	}

	if len(ads) == 0 {
		t.Fatal("Should have returned more than 0 API Defs")
	}

	ad := ads[0]

	if ad.APIID != ts.Files[0].APIID {
		t.Fatalf("APIID Was not properly set, expected: %v, got %v", ts.Files[0].APIID, ad.APIID)
	}
}

func TestGitGetter_FetchAPIDef_Swagger(t *testing.T) {
	g, e := NewGGetter(REPO, "refs/heads/swagger-test", []byte{}, mockPublish)
	if e != nil {
		t.Fatal(e)
	}

	e = g.FetchRepo()
	if e != nil {
		t.Fatal(e)
	}

	ts, err := g.FetchTykSpec()
	if err != nil {
		t.Fatal(err)
	}

	if ts.Type != TYPE_OAI {
		t.Fatalf("Spec type setting is unexpected expected: 'oas', got %v", ts.Type)
	}

	ads, err := g.FetchAPIDef(ts)
	if err != nil {
		t.Fatal(err)
	}

	if len(ads) == 0 {
		t.Fatal("Should have returned more than 0 API Defs")
	}

	ad := ads[0]

	if ad.Name != "Swagger Petstore" {
		t.Fatalf("Name Was not properly set, expected: 'Swagger Petstore', got %v", ad.Name)
	}

	if ad.APIID != ts.Files[0].APIID {
		t.Fatalf("APIID Was not properly set, expected: %v, got %v", ts.Files[0].APIID, ad.APIID)
	}

	if ad.Proxy.TargetURL != ts.Files[0].OAS.OverrideTarget {
		t.Fatalf("Target Was not properly set, got: %v, expected %v", ad.Proxy.TargetURL, ts.Files[0].OAS.OverrideTarget)
	}

	if ad.Proxy.ListenPath != ts.Files[0].OAS.OverrideListenPath {
		t.Fatalf("Target Was not properly set, expected: %v, got %v", ad.Proxy.ListenPath, ts.Files[0].OAS.OverrideListenPath)
	}
}
