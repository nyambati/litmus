package snapshot

import (
	"regexp"
	"strings"
)

const wildcardPlaceholder = ""

var (
	ncgRegex    = regexp.MustCompile(`^\^?\(\?:([^)]+)\)\$?$`)
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

// replaceWildcards replaces .* and .+ with the placeholder constant.
func replaceWildcards(s string) string {
	s = strings.ReplaceAll(s, ".*", wildcardPlaceholder)
	s = strings.ReplaceAll(s, ".+", wildcardPlaceholder)
	return s
}

// isPureWildcard reports whether s is solely a wildcard token.
func isPureWildcard(s string) bool {
	return s == ".*" || s == ".+"
}

// ExpandAlternations extracts all concrete options from a regex pattern.
func (re *RegexExpander) ExpandAlternations(pattern string) []string {
	// Pure wildcard — no useful value, drop.
	if isPureWildcard(pattern) {
		return nil
	}

	// Non-capturing group: (?:opt1|opt2)$ — Alertmanager's compiled regex format
	if matches := ncgRegex.FindStringSubmatch(pattern); matches != nil {
		return re.sanitizeParts(strings.Split(matches[1], "|"))
	}

	// Simple capturing group: (opt1|opt2)
	if matches := capRegex.FindStringSubmatch(pattern); matches != nil {
		return re.sanitizeParts(strings.Split(matches[1], "|"))
	}

	// Bare alternation without parens: a|b|c.*
	if strings.Contains(pattern, "|") {
		return re.sanitizeParts(strings.Split(pattern, "|"))
	}

	// Anchored prefix: ^api-.* — strip anchor, replace wildcards inline
	if matches := prefixRegex.FindStringSubmatch(pattern); matches != nil {
		rest := pattern[1:] // drop leading ^
		return []string{replaceWildcards(rest)}
	}

	// Character class: [a-z] -> "a"
	if matches := charRegex.FindStringSubmatch(pattern); matches != nil {
		if len(matches[1]) > 0 {
			return []string{string(matches[1][0])}
		}
	}

	// No expansion: replace any wildcards inline and return.
	return []string{replaceWildcards(pattern)}
}

// sanitizeParts strips regex metacharacters from alternation parts.
func (re *RegexExpander) sanitizeParts(parts []string) []string {
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimPrefix(p, "^")
		if isPureWildcard(p) {
			continue
		}
		p = replaceWildcards(p)
		if p != "" {
			result = append(result, p)
		}
	}
	if len(result) == 0 {
		return nil
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

// minimalCoveringSet selects up to maxCombinations combos that maximise coverage
// without generating the full Cartesian product to avoid OOM.
func (lcg *LabelCombinationGenerator) minimalCoveringSet(matchers map[string][]string) []map[string]string {
	if len(matchers) == 0 {
		return nil
	}

	keys := make([]string, 0, len(matchers))
	for k := range matchers {
		keys = append(keys, k)
	}

	covered := make(map[string]map[string]bool)
	uncoveredCount := 0
	for k, vals := range matchers {
		covered[k] = make(map[string]bool)
		uncoveredCount += len(vals)
	}

	var result []map[string]string

	// 1. Greedily generate combinations to cover all values at least once
	for uncoveredCount > 0 && len(result) < lcg.maxCombinations {
		combo := make(map[string]string)
		for _, k := range keys {
			vals := matchers[k]
			// Find an uncovered value if possible
			var chosen string
			found := false
			for _, v := range vals {
				if !covered[k][v] {
					chosen = v
					found = true
					break
				}
			}
			// Fallback to first value if all are covered for this key
			if !found {
				chosen = vals[0]
			}
			combo[k] = chosen
		}

		result = append(result, combo)
		// Mark as covered and update counter
		for k, v := range combo {
			if !covered[k][v] {
				covered[k][v] = true
				uncoveredCount--
			}
		}
	}

	// 2. If we still have room, add a few more "interesting" combinations
	// (e.g., using different values for the first few keys)
	if len(result) < lcg.maxCombinations {
		// This is a simple fallback to fill up to maxCombinations if needed.
		// In practice, the first loop often covers most scenarios or hits the limit.
		for i := 1; len(result) < lcg.maxCombinations; i++ {
			combo := make(map[string]string)
			changed := false
			for j, k := range keys {
				vals := matchers[k]
				idx := (i + j) % len(vals)
				combo[k] = vals[idx]
				if idx > 0 {
					changed = true
				}
			}
			if !changed && i > 0 {
				break // We've looped through all simple variations
			}
			result = append(result, combo)
		}
	}

	return result
}
