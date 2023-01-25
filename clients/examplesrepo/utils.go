package examplesrepo

func IndexHasExamples(index *RepositoryIndex) bool {
	if index == nil {
		return false
	}

	return len(index.Examples.UDG) > 0
}

func MergeExamples(index *RepositoryIndex) []ExampleMetadata {
	var examples []ExampleMetadata
	if index == nil {
		return examples
	}

	for _, udgExample := range index.Examples.UDG {
		examples = append(examples, udgExample)
	}

	return examples
}
