package pipeline

import (
	"testing"

	amLabels "github.com/prometheus/alertmanager/pkg/labels"

	"github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouter_RootFallback(t *testing.T) {
	root := &config.Route{
		Receiver: "default",
		Routes:   []*config.Route{},
	}
	r := NewRouter(root)
	got := r.Match(model.LabelSet{"alertname": "X"})
	assert.Equal(t, []string{"default"}, got)
}

func TestRouter_ExactMatch(t *testing.T) {
	root := &config.Route{
		Receiver: "default",
		Routes: []*config.Route{
			{
				Receiver: "team-a",
				Match:    map[string]string{"label_team": "a"},
			},
		},
	}
	r := NewRouter(root)

	got := r.Match(model.LabelSet{"label_team": "a"})
	assert.Equal(t, []string{"team-a"}, got)

	got = r.Match(model.LabelSet{"label_team": "b"})
	assert.Equal(t, []string{"default"}, got)
}

func TestRouter_RegexMatch(t *testing.T) {
	re := config.Regexp{}
	err := re.UnmarshalYAML(func(v interface{}) error {
		*(v.(*string)) = "routing|assignment"
		return nil
	})
	require.NoError(t, err)
	root := &config.Route{
		Receiver: "default",
		Routes: []*config.Route{
			{
				Receiver: "routing-team",
				MatchRE:  map[string]config.Regexp{"label_team": re},
			},
		},
	}
	r := NewRouter(root)

	got := r.Match(model.LabelSet{"label_team": "routing"})
	assert.Equal(t, []string{"routing-team"}, got)

	got = r.Match(model.LabelSet{"label_team": "unrelated"})
	assert.Equal(t, []string{"default"}, got)
}

func TestRouter_ContinueTrue_MultipleReceivers(t *testing.T) {
	root := &config.Route{
		Receiver: "default",
		Routes: []*config.Route{
			{
				Receiver: "first",
				Match:    map[string]string{"env": "prod"},
				Continue: true,
			},
			{
				Receiver: "second",
				Match:    map[string]string{"team": "ops"},
				Continue: true,
			},
		},
	}
	r := NewRouter(root)

	got := r.Match(model.LabelSet{"env": "prod", "team": "ops"})
	assert.ElementsMatch(t, []string{"first", "second"}, got)
}

func TestRouter_ContinueFalse_StopsSiblings(t *testing.T) {
	root := &config.Route{
		Receiver: "default",
		Routes: []*config.Route{
			{
				Receiver: "first",
				Match:    map[string]string{"env": "prod"},
				Continue: false, // stop here
			},
			{
				Receiver: "second",
				Match:    map[string]string{"team": "ops"},
			},
		},
	}
	r := NewRouter(root)

	got := r.Match(model.LabelSet{"env": "prod", "team": "ops"})
	assert.Equal(t, []string{"first"}, got)
}

func TestRouter_NestedRoute(t *testing.T) {
	root := &config.Route{
		Receiver: "default",
		Routes: []*config.Route{
			{
				Receiver: "parent",
				Match:    map[string]string{"team": "logistics"},
				Continue: true,
				Routes: []*config.Route{
					{
						Receiver: "child-prod",
						Match:    map[string]string{"env": "production"},
						Continue: false,
					},
				},
			},
		},
	}
	r := NewRouter(root)

	// production → child overrides parent's receiver
	got := r.Match(model.LabelSet{"team": "logistics", "env": "production"})
	assert.Equal(t, []string{"child-prod"}, got)

	// staging → no child match → parent receiver used
	got = r.Match(model.LabelSet{"team": "logistics", "env": "staging"})
	assert.Equal(t, []string{"parent"}, got)
}

func TestRouter_ModernMatchers(t *testing.T) {
	m, err := amLabels.NewMatcher(amLabels.MatchEqual, "severity", "critical")
	require.NoError(t, err)
	root := &config.Route{
		Receiver: "default",
		Routes: []*config.Route{
			{
				Receiver: "critical-team",
				Matchers: config.Matchers{m},
			},
		},
	}
	r := NewRouter(root)

	got := r.Match(model.LabelSet{"severity": "critical"})
	assert.Equal(t, []string{"critical-team"}, got)

	got = r.Match(model.LabelSet{"severity": "warning"})
	assert.Equal(t, []string{"default"}, got)
}
