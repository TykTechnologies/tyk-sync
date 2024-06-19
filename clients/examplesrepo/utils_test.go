package examplesrepo

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndexHasExamples(t *testing.T) {
	t.Run("should return false if index is nil", func(t *testing.T) {
		assert.False(t, IndexHasExamples(nil))
	})

	t.Run("should return false if no examples are present", func(t *testing.T) {
		index := RepositoryIndex{}
		assert.False(t, IndexHasExamples(&index))
	})

	t.Run("should return true if index has at least one example", func(t *testing.T) {
		index := RepositoryIndex{
			Examples: ExamplesCategories{
				"udg": []ExampleMetadata{
					{
						Location: "location",
					},
				},
			},
		}

		assert.True(t, IndexHasExamples(&index))
	})
}

func TestMergeExamples(t *testing.T) {
	t.Run("should return empty slice when index is nil", func(t *testing.T) {
		examples := MergeExamples(nil)
		assert.Len(t, examples, 0)
	})

	t.Run("should merge examples successfully", func(t *testing.T) {
		udgExample := ExampleMetadata{
			Location: "udg",
		}

		graphqlExample := ExampleMetadata{
			Location: "graphql",
		}

		index := RepositoryIndex{
			Examples: ExamplesCategories{
				"udg": []ExampleMetadata{
					udgExample,
				},
				"graphql": []ExampleMetadata{
					graphqlExample,
				},
			},
		}

		expectedExamples := []ExampleMetadata{
			udgExample,
			graphqlExample,
		}

		examples := MergeExamples(&index)
		assert.Equal(t, expectedExamples, examples)
	})
}

func TestExamplesAsLocationIndexedMap(t *testing.T) {
	t.Run("should return nil when no examples are available", func(t *testing.T) {
		examplesMap := ExamplesAsLocationIndexedMap(nil)
		assert.Nil(t, examplesMap)
	})

	t.Run("should successfully create an examples map", func(t *testing.T) {
		udgExample1 := ExampleMetadata{
			Location: "udg/example-1",
		}

		udgExample2 := ExampleMetadata{
			Location: "udg/example-2",
		}

		graphqlExample1 := ExampleMetadata{
			Location: "graphql/example-1",
		}

		index := RepositoryIndex{
			Examples: ExamplesCategories{
				"udg": []ExampleMetadata{
					udgExample1,
					udgExample2,
				},
				"graphql": []ExampleMetadata{
					graphqlExample1,
				},
			},
		}

		examplesMap := ExamplesAsLocationIndexedMap(&index)
		assert.Len(t, examplesMap, 3)
		assert.Equal(t, udgExample1, examplesMap[udgExample1.Location])
		assert.Equal(t, udgExample2, examplesMap[udgExample2.Location])
		assert.Equal(t, graphqlExample1, examplesMap[graphqlExample1.Location])
	})
}
