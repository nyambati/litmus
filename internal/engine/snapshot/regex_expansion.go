package snapshot

import (
	"regexp"
	"strings"
)

// RegexExpander extracts values from regex patterns.
type RegexExpander struct{}

// NewRegexExpander creates regex expander.
func NewRegexExpander() *RegexExpander {
	return &RegexExpander{}
}

// ExpandAlternations extracts all options from alternation pattern.
func (re *RegexExpander) ExpandAlternations(pattern string) []string {
	// Simple alternation: (opt1|opt2|opt3)
	altRegex := regexp.MustCompile(`^\(([^)]+)\)$`)
	if matches := altRegex.FindStringSubmatch(pattern); matches != nil {
		opts := strings.Split(matches[1], "|")
		return opts
	}

	// Anchored prefix: ^api-.* -> "api-"
	prefixRegex := regexp.MustCompile(`^\^([^.]+)`)
	if matches := prefixRegex.FindStringSubmatch(pattern); matches != nil {
		return []string{matches[1]}
	}

	// Wildcard: .* -> "litmus_match"
	if pattern == ".*" || pattern == ".+" {
		return []string{"litmus_match"}
	}

	// Character class: [a-z] -> "a"
	charRegex := regexp.MustCompile(`^\[([a-z])-([a-z])\]$`)
	if matches := charRegex.FindStringSubmatch(pattern); matches != nil {
		return []string{string(matches[1][0])}
	}

	// No expansion: return pattern as-is
	return []string{pattern}
}

// LabelCombinationGenerator creates balanced label combinations.
type LabelCombinationGenerator struct {
	maxCombinations int
}

// NewLabelCombinationGenerator creates combination generator.
func NewLabelCombinationGenerator(max int) *LabelCombinationGenerator {
	return &LabelCombinationGenerator{maxCombinations: max}
}

// GenerateCovering generates balanced covering set of label combinations.
// If product <= max, generates full Cartesian product.
// Otherwise, generates minimum set to exercise each option at least once.
func (lcg *LabelCombinationGenerator) GenerateCovering(matchers map[string][]string) []map[string]string {
	// Count total combinations
	totalCombos := 1
	for _, vals := range matchers {
		totalCombos *= len(vals)
	}

	// Full Cartesian if small enough
	if totalCombos <= lcg.maxCombinations {
		return lcg.cartesianProduct(matchers)
	}

	// Covering set: minimum to exercise each option
	return lcg.minimalCoveringSet(matchers)
}

// cartesianProduct generates all combinations.
func (lcg *LabelCombinationGenerator) cartesianProduct(matchers map[string][]string) []map[string]string {
	keys := make([]string, 0, len(matchers))
	for k := range matchers {
		keys = append(keys, k)
	}

	var result []map[string]string
	var build func(int, map[string]string)
	build = func(depth int, current map[string]string) {
		if depth == len(keys) {
			m := make(map[string]string)
			for k, v := range current {
				m[k] = v
			}
			result = append(result, m)
			return
		}

		key := keys[depth]
		for _, val := range matchers[key] {
			current[key] = val
			build(depth+1, current)
		}
	}

	build(0, make(map[string]string))
	return result
}

// minimalCoveringSet generates combinations via partial Cartesian product up to max.
func (lcg *LabelCombinationGenerator) minimalCoveringSet(matchers map[string][]string) []map[string]string {
	// Start with full Cartesian and trim to maxCombinations
	full := lcg.cartesianProduct(matchers)
	if len(full) <= lcg.maxCombinations {
		return full
	}

	// Greedy selection: pick combos that maximize uncovered options
	var result []map[string]string
	covered := make(map[string]map[string]bool)
	for k := range matchers {
		covered[k] = make(map[string]bool)
	}

	remaining := make([]map[string]string, len(full))
	copy(remaining, full)

	for len(result) < lcg.maxCombinations && len(remaining) > 0 {
		// Pick combo covering most new options
		bestIdx := 0
		bestScore := -1

		for i, combo := range remaining {
			score := 0
			for k, v := range combo {
				if !covered[k][v] {
					score++
				}
			}
			if score > bestScore {
				bestScore = score
				bestIdx = i
			}
		}

		// Add best combo
		chosen := remaining[bestIdx]
		result = append(result, chosen)
		for k, v := range chosen {
			covered[k][v] = true
		}

		// Remove chosen from remaining
		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	return result
}
