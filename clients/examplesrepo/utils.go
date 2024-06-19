package examplesrepo

import (
	"sort"
)

func IndexHasExamples(index *RepositoryIndex) bool {
	if index == nil {
		return false
	}

	for category := range index.Examples {
		if len(index.Examples[category]) > 0 {
			return true
		}
	}

	return false
}

func MergeExamples(index *RepositoryIndex) []ExampleMetadata {
	var examples []ExampleMetadata
	if index == nil {
		return examples
	}

	for category := range index.Examples {
		examples = append(examples, index.Examples[category]...)
	}

	sort.Slice(examples, func(i, j int) bool {
		return examples[i].Location < examples[j].Location
	})

	return examples
}

func ExamplesAsLocationIndexedMap(index *RepositoryIndex) map[string]ExampleMetadata {
	examples := MergeExamples(index)
	if len(examples) == 0 {
		return nil
	}

	examplesMap := make(map[string]ExampleMetadata, len(examples))
	for _, example := range examples {
		examplesMap[example.Location] = example
	}

	return examplesMap
}
