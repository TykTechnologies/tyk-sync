package examplesrepo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExamplesClient_GetRepositoryIndex(t *testing.T) {
	successTestServer := createRepositoryTestServer(t, http.StatusOK)
	notFoundTestServer := createRepositoryTestServer(t, http.StatusNotFound)
	t.Cleanup(func() {
		successTestServer.Close()
		notFoundTestServer.Close()
	})

	t.Run("should return error if no success status code was captured", func(t *testing.T) {
		client, err := NewExamplesClient(notFoundTestServer.URL)
		require.NoError(t, err)

		_, err = client.GetRepositoryIndex()
		assert.Error(t, err)
	})

	t.Run("should successfully get repository index", func(t *testing.T) {
		runSuccessfulTest := func(testServerURL string) func(t *testing.T) {
			return func(t *testing.T) {
				client, err := NewExamplesClient(testServerURL)
				require.NoError(t, err)

				index, err := client.GetRepositoryIndex()
				assert.NoError(t, err)
				assert.Equal(t, repositoryIndexModel, index)
			}
		}

		t.Run("for status code 200", runSuccessfulTest(successTestServer.URL))
	})
}

func TestExamplesClient_GetAllExamples(t *testing.T) {
	successTestServer := createRepositoryTestServer(t, http.StatusOK)
	t.Cleanup(func() {
		successTestServer.Close()
	})

	t.Run("should successfully return all examples", func(t *testing.T) {
		client, err := NewExamplesClient(successTestServer.URL)
		require.NoError(t, err)

		examples, err := client.GetAllExamples()
		assert.Len(t, examples, len(repositoryIndexModel.Examples.UDG))
		assert.Equal(t, repositoryIndexModel.Examples.UDG[0], examples[0])
	})
}

func TestExamplesClient_GetAllExamplesAsLocationIndexedMap(t *testing.T) {
	successTestServer := createRepositoryTestServer(t, http.StatusOK)
	t.Cleanup(func() {
		successTestServer.Close()
	})

	t.Run("should successfully return examples map", func(t *testing.T) {
		client, err := NewExamplesClient(successTestServer.URL)
		require.NoError(t, err)

		examplesMap, err := client.GetAllExamplesAsLocationIndexedMap()
		assert.Len(t, examplesMap, len(repositoryIndexModel.Examples.UDG))
		assert.Equal(t, repositoryIndexModel.Examples.UDG[0], examplesMap[repositoryIndexModel.Examples.UDG[0].Location])
	})
}

func createRepositoryTestServer(t *testing.T, statusCode int) *httptest.Server {
	t.Helper()
	repositoryIndexHandler := func(w http.ResponseWriter, r *http.Request) {
		require.True(t, strings.HasSuffix(r.URL.Path, RepoIndexFile))
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		if statusCode >= 200 && statusCode <= 399 {
			_, err := w.Write([]byte(repositoryIndexJson))
			fmt.Println(err)
		}
	}
	return httptest.NewServer(http.HandlerFunc(repositoryIndexHandler))
}

var repositoryIndexModel = &RepositoryIndex{
	Examples: ExamplesCategories{
		UDG: []ExampleMetadata{
			{
				Location:    "udg/first-demo",
				Name:        "First UDG Demo",
				Description: "This UDG demo is very simple",
				Features: []string{
					"GraphQL DataSource",
					"REST DataSource",
				},
				MinTykVersion: "5.0",
			},
			{
				Location:    "udg/complex-demo",
				Name:        "Complex UDG Demo",
				Description: "This UDG demo is very complex",
				Features: []string{
					"GraphQL DataSource",
					"REST DataSource",
					"Kafka DataSource",
					"Subscriptions",
				},
				MinTykVersion: "5.0",
			},
		},
	},
}

const repositoryIndexJson = `{
	"examples": {
		"udg": [
			{
				"location": "udg/first-demo",
				"name": "First UDG Demo",
				"description": "This UDG demo is very simple",
				"features": [
					"GraphQL DataSource",
					"REST DataSource"
				],
				"minTykVersion": "5.0"
			},
			{
				"location": "udg/complex-demo",
				"name": "Complex UDG Demo",
				"description": "This UDG demo is very complex",
				"features": [
					"GraphQL DataSource",
					"REST DataSource",
					"Kafka DataSource",
					"Subscriptions"
				],
				"minTykVersion": "5.0"
			}
		]
	}
}`
