package sanity

import (
	"testing"

	"github.com/prometheus/alertmanager/config"
	labels "github.com/prometheus/alertmanager/pkg/labels"
	"github.com/stretchr/testify/require"
)

func TestInhibitionCycleDetector_NoCycle(t *testing.T) {
	rules := []*config.InhibitRule{
		{
			SourceMatch: map[string]string{"severity": "critical"},
			TargetMatch: map[string]string{"severity": "warning"},
		},
		{
			SourceMatch: map[string]string{"severity": "warning"},
			TargetMatch: map[string]string{"severity": "info"},
		},
	}

	detector := NewInhibitionCycleDetector(rules)
	cycles := detector.DetectCycles()

	require.Len(t, cycles, 0)
}

func TestInhibitionCycleDetector_DirectCycle(t *testing.T) {
	rules := []*config.InhibitRule{
		{
			SourceMatch: map[string]string{"alert": "A"},
			TargetMatch: map[string]string{"alert": "B"},
		},
		{
			SourceMatch: map[string]string{"alert": "B"},
			TargetMatch: map[string]string{"alert": "A"},
		},
	}

	detector := NewInhibitionCycleDetector(rules)
	cycles := detector.DetectCycles()

	require.Len(t, cycles, 1)
	require.Contains(t, cycles[0], "cycle")
}

func TestInhibitionCycleDetector_MatcherBasedCycle(t *testing.T) {
	sourceA, err := labels.NewMatcher(labels.MatchEqual, "alert", "A")
	require.NoError(t, err)
	targetA, err := labels.NewMatcher(labels.MatchEqual, "alert", "A")
	require.NoError(t, err)
	sourceB, err := labels.NewMatcher(labels.MatchEqual, "alert", "B")
	require.NoError(t, err)
	targetB, err := labels.NewMatcher(labels.MatchEqual, "alert", "B")
	require.NoError(t, err)

	rules := []*config.InhibitRule{
		{
			SourceMatchers: config.Matchers{sourceA},
			TargetMatchers: config.Matchers{targetB},
		},
		{
			SourceMatchers: config.Matchers{sourceB},
			TargetMatchers: config.Matchers{targetA},
		},
	}

	detector := NewInhibitionCycleDetector(rules)
	cycles := detector.DetectCycles()

	require.Len(t, cycles, 1)
	require.Contains(t, cycles[0], "cycle")
}

func TestOrphanReceiverDetector_NoOrphans(t *testing.T) {
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
			},
			{
				Receiver: "db-team",
			},
		},
	}

	receivers := map[string]*config.Receiver{
		"root":     {},
		"api-team": {},
		"db-team":  {},
	}

	detector := NewOrphanReceiverDetector(root, receivers)
	orphans := detector.DetectOrphans()

	require.Len(t, orphans, 0)
}

func TestOrphanReceiverDetector_HasOrphans(t *testing.T) {
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
			},
		},
	}

	receivers := map[string]*config.Receiver{
		"root":       {},
		"api-team":   {},
		"unused-dev": {},
	}

	detector := NewOrphanReceiverDetector(root, receivers)
	orphans := detector.DetectOrphans()

	require.Len(t, orphans, 1)
	require.Contains(t, orphans[0], "unused-dev")
}

func TestOrphanReceiverDetector_NestedRoutes(t *testing.T) {
	root := &config.Route{
		Receiver: "root",
		Routes: []*config.Route{
			{
				Receiver: "api-team",
				Routes: []*config.Route{
					{
						Receiver: "api-critical",
					},
				},
			},
		},
	}

	receivers := map[string]*config.Receiver{
		"root":          {},
		"api-team":      {},
		"api-critical":  {},
		"unused-backup": {},
	}

	detector := NewOrphanReceiverDetector(root, receivers)
	orphans := detector.DetectOrphans()

	require.Len(t, orphans, 1)
	require.Contains(t, orphans[0], "unused-backup")
}
