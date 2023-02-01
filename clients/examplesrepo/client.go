package examplesrepo

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	RepoRootUrl   = "https://raw.githubusercontent.com/TykTechnologies/tyk-examples/main"
	RepoGitUrl    = "https://github.com/TykTechnologies/tyk-examples.git"
	RepoIndexFile = "repository.json"
)

type ExamplesClient struct {
	RepositoryRootUrl *url.URL
	httpClient        *http.Client
}

func NewExamplesClient(repoRootUrl string) (*ExamplesClient, error) {
	parsedUrl, err := url.Parse(repoRootUrl)
	if err != nil {
		return nil, err
	}

	return &ExamplesClient{
		RepositoryRootUrl: parsedUrl,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

func (e *ExamplesClient) GetRepositoryIndex() (*RepositoryIndex, error) {
	targetUrl := fmt.Sprintf("%s/%s", e.RepositoryRootUrl.String(), RepoIndexFile)
	req, err := http.NewRequest(http.MethodGet, targetUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	// We cannot rely on status code 200 for GitHub, so we need a bit more flexibility here
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("example repository responded with unexpected status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	repositoryIndex := &RepositoryIndex{}
	err = json.Unmarshal(body, repositoryIndex)
	if err != nil {
		return nil, err
	}

	return repositoryIndex, nil
}

func (e *ExamplesClient) GetAllExamples() ([]ExampleMetadata, error) {
	index, err := e.GetRepositoryIndex()
	if err != nil {
		return nil, err
	}

	if !IndexHasExamples(index) {
		return nil, err
	}

	return MergeExamples(index), nil
}

func (e *ExamplesClient) GetAllExamplesAsLocationIndexedMap() (map[string]ExampleMetadata, error) {
	repositoryIndex, err := e.GetRepositoryIndex()
	if err != nil {
		return nil, err
	}

	return ExamplesAsLocationIndexedMap(repositoryIndex), nil
}
