package matching

// ExactMatch returns true if actual and expected contain exactly the same receivers.
// Order-independent. Both slices must have the same elements.
//
// Used for regression testing where we need to detect any changes to routing behavior.
func ExactMatch(actual, expected []string) bool {
	if len(actual) != len(expected) {
		return false
	}
	actualMap := make(map[string]bool)
	for _, r := range actual {
		actualMap[r] = true
	}
	for _, r := range expected {
		if !actualMap[r] {
			return false
		}
	}
	return true
}

// SubsetMatch returns true if actual contains all expected receivers (and possibly more).
// Order-independent. Allows actual to have additional receivers beyond expected.
//
// Used for behavioral testing where we verify that critical receivers are included,
// but don't fail if routing logic adds additional receivers.
func SubsetMatch(actual, expected []string) bool {
	actualMap := make(map[string]bool)
	for _, r := range actual {
		actualMap[r] = true
	}
	for _, r := range expected {
		if !actualMap[r] {
			return false
		}
	}
	return true
}
