package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/TykTechnologies/tyk-sync/clients/examplesrepo"
)

func TestGenerateExampleDetailsString(t *testing.T) {
	example := examplesrepo.ExampleMetadata{
		Location:      "udg/example",
		Name:          "An Example",
		Description:   "An example that can be published",
		Features:      []string{"rest", "graphql", "kafka"},
		MinTykVersion: "5.0",
	}

	exampleDetails := generateExampleDetailsString(example)
	expectedExampleDetails := `LOCATION
udg/example

NAME
An Example

DESCRIPTION
An example that can be published

FEATURES
- rest
- graphql
- kafka

MIN TYK VERSION
5.0`
	assert.Equal(t, expectedExampleDetails, exampleDetails)
}
