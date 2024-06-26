package examplesrepo

type RepositoryIndex struct {
	Examples ExamplesCategories `json:"examples"`
}

type ExamplesCategories map[string][]ExampleMetadata

type ExampleMetadata struct {
	Location      string   `json:"location"`
	Name          string   `json:"name"`
	Description   string   `json:"description"`
	Features      []string `json:"features"`
	MinTykVersion string   `json:"minTykVersion"`
}
