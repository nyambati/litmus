package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/nyambati/litmus/internal/stores"
	lithtypes "github.com/nyambati/litmus/internal/types"

	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestPipeline_Execute_Silenced(t *testing.T) {
	silences := []lithtypes.Silence{
		{
			Labels:  map[string]string{"service": "api"},
			Comment: "api maintenance",
		},
	}
	silenceStore := stores.NewSilenceStore(silences)
	alertStore := stores.NewAlertStore()
	router := NewRouter(&amconfig.Route{Receiver: "default"})

	pipeline := NewRunner(silenceStore, alertStore, router, nil)

	labels := model.LabelSet{"service": "api", "severity": "critical"}
	outcome, err := pipeline.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.Equal(t, "silenced", outcome.Status)
}

func TestPipeline_Execute_Active(t *testing.T) {
	silenceStore := stores.NewSilenceStore([]lithtypes.Silence{})
	alertStore := stores.NewAlertStore()
	router := NewRouter(&amconfig.Route{Receiver: "default"})

	pipeline := NewRunner(silenceStore, alertStore, router, nil)

	labels := model.LabelSet{"service": "api", "severity": "critical"}
	outcome, err := pipeline.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.Equal(t, "active", outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestPipeline_Execute_Inhibited(t *testing.T) {
	silenceStore := stores.NewSilenceStore([]lithtypes.Silence{})
	alertStore := stores.NewAlertStore()

	inhibitor := &types.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	alertStore.Put(inhibitor)

	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"severity": "critical"},
			TargetMatch: map[string]string{"severity": "warning"},
			Equal:       model.LabelNames{"service"},
		},
	}

	router := NewRouter(&amconfig.Route{Receiver: "default"})
	pipeline := NewRunner(silenceStore, alertStore, router, rules)

	labels := model.LabelSet{"service": "api", "severity": "warning"}
	outcome, err := pipeline.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.Equal(t, "inhibited", outcome.Status)
}
