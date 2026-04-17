package sanity

import (
	"testing"

	"github.com/prometheus/alertmanager/config"
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
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
				Match:    map[string]string{"service": "api", "env": "prod"},
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
