package utils

import (
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"
)

// LabelFormat returns a stable, order-independent string key for a label map.
// Output format: "k1=v1,k2=v2,..." sorted by key. Nil or empty map returns "".
func LabelFormat(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(m[k])
	}
	return b.String()
}

func ExpandEnvVars(s string) (string, error) {
	var envPattern = regexp.MustCompile(`env\(([A-Za-z_][A-Za-z0-9_]*)\)`)
	if !strings.Contains(s, "env(") {
		return s, nil
	}
	var expandErr error
	result := envPattern.ReplaceAllStringFunc(s, func(match string) string {
		if expandErr != nil {
			return match
		}
		sub := envPattern.FindStringSubmatch(match)
		if len(sub) < 2 {
			return match
		}
		val, ok := os.LookupEnv(strings.ToUpper(sub[1]))
		if !ok {
			expandErr = fmt.Errorf("env var %q referenced in config but not set", strings.ToUpper(sub[1]))
			return match
		}
		return val
	})
	if expandErr != nil {
		return "", expandErr
	}
	return result, nil
}
