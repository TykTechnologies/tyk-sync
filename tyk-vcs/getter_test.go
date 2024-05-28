package tyk_vcs

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

const REPO string = "https://github.com/lonelycode/integration-test.git"
const README_REPO string = "https://github.com/TykTechnologies/tyk-sync"

func TestNewGGetter(t *testing.T) {
	_, e := NewGGetter(REPO, "refs/heads/master", []byte{}, "")
	if e != nil {
		t.Fatal(e)
	}
}

func TestGitGetter_FetchRepo(t *testing.T) {
	g, e := NewGGetter(REPO, "refs/heads/master", []byte{}, "")
	if e != nil {
		t.Fatal(e)
	}

	e = g.FetchRepo()
	if e != nil {
		t.Fatal(e)
	}
}

func TestGitGetter_FetchTykSpec(t *testing.T) {
	g, e := NewGGetter(REPO, "refs/heads/master", []byte{}, "")
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

func TestGitGetter_FetchReadme(t *testing.T) {
	g, e := NewGGetter(README_REPO, "refs/heads/master", []byte{}, "")
	if e != nil {
		t.Fatal(e)
	}

	e = g.FetchRepo()
	if e != nil {
		t.Fatal(e)
	}

	readmeContent, err := g.FetchReadme()
	if err != nil {
		t.Fatal(err)
	}

	if len(readmeContent) == 0 {
		t.Fatal("Readme content is empty")
	}
}

func TestGitGetter_FetchAPIDef(t *testing.T) {
	g, e := NewGGetter(REPO, "refs/heads/master", []byte{}, "")
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
	g, e := NewGGetter(REPO, "refs/heads/swagger-test", []byte{}, "")
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

func TestGetFilepath(t *testing.T) {
	t.Run("filepath without path segments", func(t *testing.T) {
		fullPath := getFilepath(".tyk.json", "")
		assert.Equal(t, ".tyk.json", fullPath)
	})
	t.Run("filepath with path segments", func(t *testing.T) {
		fullPath := getFilepath(".tyk-json", "examples/udg/simple")
		assert.Equal(t, filepath.Clean("examples/udg/simple/.tyk-json"), fullPath)
	})
	t.Run("filepath with multiple separators in path segments", func(t *testing.T) {
		fullPath := getFilepath(".tyk-json", "examples////udg///simple////")
		assert.Equal(t, filepath.Clean("examples/udg/simple/.tyk-json"), fullPath)
	})
	t.Run("filepath with multiple path segments", func(t *testing.T) {
		fullPath := getFilepath(".tyk-json", "examples//udg//", "/simple")
		assert.Equal(t, filepath.Clean("examples/udg/simple/.tyk-json"), fullPath)
	})
}
