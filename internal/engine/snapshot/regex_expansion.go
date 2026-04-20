package snapshot

import (
	"regexp"
	"strings"
)

var (
	ncgRegex    = regexp.MustCompile(`^\(\?:([^)]+)\)\$?$`)
	capRegex    = regexp.MustCompile(`^\(([^)]+)\)$`)
	prefixRegex = regexp.MustCompile(`^\^([^.]+)`)
	charRegex   = regexp.MustCompile(`^\[([a-z])-([a-z])\]$`)
)

// RegexExpander extracts values from regex patterns.
type RegexExpander struct{}

// NewRegexExpander creates regex expander.
func NewRegexExpander() *RegexExpander {
	return &RegexExpander{}
}

// ExpandAlternations extracts all concrete options from a regex pattern.
func (re *RegexExpander) ExpandAlternations(pattern string) []string {
	// Non-capturing group: (?:opt1|opt2)$ — Alertmanager's compiled regex format
	if matches := ncgRegex.FindStringSubmatch(pattern); matches != nil {
		return sanitizeParts(strings.Split(matches[1], "|"))
	}

	// Simple capturing group: (opt1|opt2)
	if matches := capRegex.FindStringSubmatch(pattern); matches != nil {
		return sanitizeParts(strings.Split(matches[1], "|"))
	}

	// Bare alternation without parens: a|b|c.*
	if strings.Contains(pattern, "|") {
		return sanitizeParts(strings.Split(pattern, "|"))
	}

	// Anchored prefix: ^api-.* -> "api-"
	if matches := prefixRegex.FindStringSubmatch(pattern); matches != nil {
		return []string{matches[1]}
	}

	// Suffix wildcard: prd.* -> "prd"
	if strings.HasSuffix(pattern, ".*") || strings.HasSuffix(pattern, ".+") {
		base := strings.TrimSuffix(strings.TrimSuffix(pattern, ".*"), ".+")
		if base == "" {
			return []string{"litmus_match"}
		}
		return []string{base}
	}

	// Pure wildcard
	if pattern == ".*" || pattern == ".+" {
		return []string{"litmus_match"}
	}

	// Character class: [a-z] -> "a"
	if matches := charRegex.FindStringSubmatch(pattern); matches != nil {
		if len(matches[1]) > 0 {
			return []string{string(matches[1][0])}
		}
	}

	// No expansion: return literal as-is
	return []string{pattern}
}

// sanitizeParts strips regex metacharacters from alternation parts and deduplicates.
func sanitizeParts(parts []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSuffix(strings.TrimSuffix(p, ".*"), ".+")
		p = strings.TrimPrefix(p, "^")
		if p != "" && !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result
}

// LabelCombinationGenerator creates balanced label combinations.
type LabelCombinationGenerator struct {
	maxCombinations int
}

// NewLabelCombinationGenerator creates combination generator.
// Values < 1 are clamped to 1.
func NewLabelCombinationGenerator(maxCombinations int) *LabelCombinationGenerator {
	if maxCombinations < 1 {
		maxCombinations = 1
	}
	return &LabelCombinationGenerator{maxCombinations: maxCombinations}
}

// GenerateCovering generates balanced covering set of label combinations.
// If product <= max, generates full Cartesian product.
// Otherwise, generates minimum set to exercise each option at least once.
func (lcg *LabelCombinationGenerator) GenerateCovering(matchers map[string][]string) []map[string]string {
	totalCombos := 1
	for _, vals := range matchers {
		totalCombos *= len(vals)
	}

	if totalCombos <= lcg.maxCombinations {
		return lcg.cartesianProduct(matchers)
	}

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

// minimalCoveringSet selects up to maxCombinations combos that maximise coverage.
func (lcg *LabelCombinationGenerator) minimalCoveringSet(matchers map[string][]string) []map[string]string {
	full := lcg.cartesianProduct(matchers)

	covered := make(map[string]map[string]bool)
	for k := range matchers {
		covered[k] = make(map[string]bool)
	}

	remaining := make([]map[string]string, len(full))
	copy(remaining, full)

	var result []map[string]string
	for len(result) < lcg.maxCombinations && len(remaining) > 0 {
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

		chosen := remaining[bestIdx]
		result = append(result, chosen)
		for k, v := range chosen {
			covered[k][v] = true
		}

		remaining = append(remaining[:bestIdx], remaining[bestIdx+1:]...)
	}

	return result
}
