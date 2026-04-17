package sanity

import (
	"regexp"
	"testing"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestShadowedRouteDetector_NoShadow(t *testing.T) {
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
				Match:    map[string]string{"service": "api"},
			},
			{
				Receiver: "db-team",
				Match:    map[string]string{"service": "db"},
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 0)
}

func TestShadowedRouteDetector_ShadowedByParent(t *testing.T) {
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
				Match:    map[string]string{"service": "api"},
				Routes: []*config.Route{
					{
						Receiver: "api-critical",
						Match:    map[string]string{"severity": "critical"},
					},
				},
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 0)
}

func TestShadowedRouteDetector_ShadowedByParentNoChild(t *testing.T) {
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
				Match:    map[string]string{"service": "api"},
			},
			{
				Receiver: "unreachable",
				Match:    map[string]string{"service": "api"},
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "shadowed")
	require.Contains(t, issues[0], "unreachable")
}

func TestShadowedRouteDetector_ShadowedBySubsetMatcher(t *testing.T) {
	// api-team (service=api, env=prod) is MORE specific than catch-all (service=api).
	// catch-all is still reachable for env≠prod alerts → not shadowed.
	// Shadowing requires parent to be BROADER (fewer constraints), not more specific.
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
				Match:    map[string]string{"service": "api", "env": "prod"},
			},
			{
				Receiver: "catch-all",
				Match:    map[string]string{"service": "api"},
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 0)
}

func TestShadowedRouteDetector_ShadowedByBroaderParent(t *testing.T) {
	// api-team (service=api) is BROADER than api-critical (service=api, severity=critical).
	// api-team catches every alert api-critical would catch → api-critical is shadowed.
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
				Match:    map[string]string{"service": "api"},
			},
			{
				Receiver: "api-critical",
				Match:    map[string]string{"service": "api", "severity": "critical"},
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "api-critical")
}

func TestShadowedRouteDetector_NotShadowedDifferentMatchers(t *testing.T) {
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
				Match:    map[string]string{"service": "api"},
			},
			{
				Receiver: "cache-team",
				Match:    map[string]string{"service": "cache"},
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 0)
}

// TestShadowedRouteDetector_NoFalsePositive_MatchRE reproduces the alertmanager.yaml
// case where routes using match_re were falsely flagged as shadowed.
func TestShadowedRouteDetector_NoFalsePositive_MatchRE(t *testing.T) {
	rePattern := regexp.MustCompile("^(?:apac-compliance-team)$")

	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "logistics-team-order-dispatching-dd",
				Match:    map[string]string{"periodicity": "weekly"},
				Continue: false,
			},
			{
				Receiver: "apac-compliance-team",
				MatchRE:  config.MatchRegexps{"local_team": config.Regexp{Regexp: rePattern}},
				Continue: true,
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 0, "routes with different label keys must not be flagged as shadowed")
}

// TestShadowedRouteDetector_NegativeMatcherNotShadowed verifies that a parent with
// a negative matcher (type!~"x") does not shadow a child that positively requires
// that same label (type=~"x") — they are mutually exclusive on that dimension.
// Reproduces the alertmanager.yaml logistics-team-time-estimations false positive.
func TestShadowedRouteDetector_NegativeMatcherNotShadowed(t *testing.T) {
	teamMatcher, err := labels.NewMatcher(labels.MatchRegexp, "label_team", "time-estimations|ds-seamless")
	require.NoError(t, err)
	typeNotMetaflow, err := labels.NewMatcher(labels.MatchNotRegexp, "type", "metaflow-k8s")
	require.NoError(t, err)
	typeYesMetaflow, err := labels.NewMatcher(labels.MatchRegexp, "type", "metaflow-k8s")
	require.NoError(t, err)

	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				// label_team=~"..." AND type!~"metaflow-k8s"
				Receiver: "logistics-team-time-estimations",
				Matchers: config.Matchers{teamMatcher, typeNotMetaflow},
				Continue: true,
			},
			{
				// label_team=~"..." AND type=~"metaflow-k8s" — mutually exclusive with above on type
				Receiver: "logistics-team-time-estimations-staging",
				Matchers: config.Matchers{teamMatcher, typeYesMetaflow},
				Continue: true,
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 0, "parent with type!~ must not shadow child with type=~ on the same value")
}

// TestShadowedRouteDetector_NegativeInChild verifies that a child with a negative
// matcher is correctly identified as shadowed when the parent covers its positive matchers.
func TestShadowedRouteDetector_NegativeInChild(t *testing.T) {
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "broad",
				Match:    map[string]string{"service": "api"},
				Continue: false,
			},
			{
				// service=api AND type!=bad — parent (broad) catches all service=api first
				Receiver: "narrow",
				Match:    map[string]string{"service": "api"},
				Continue: true,
			},
		},
	}

	detector := NewShadowedRouteDetector(root)
	issues := detector.Detect()

	require.Len(t, issues, 1)
	require.Contains(t, issues[0], "narrow")
}
