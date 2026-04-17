package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/nyambati/litmus/internal/stores"
	lithtypes "github.com/nyambati/litmus/internal/types"

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

	pipeline := NewRunner(silenceStore, alertStore, []string{"default"})

	labels := model.LabelSet{"service": "api", "severity": "critical"}
	outcome, err := pipeline.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.Equal(t, "silenced", outcome.Status)
}

func TestPipeline_Execute_Active(t *testing.T) {
	silences := []lithtypes.Silence{}
	silenceStore := stores.NewSilenceStore(silences)
	alertStore := stores.NewAlertStore()

	pipeline := NewRunner(silenceStore, alertStore, []string{"default"})

	labels := model.LabelSet{"service": "api", "severity": "critical"}
	outcome, err := pipeline.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.Equal(t, "active", outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestPipeline_Execute_Inhibited(t *testing.T) {
	silences := []lithtypes.Silence{}
	silenceStore := stores.NewSilenceStore(silences)
	alertStore := stores.NewAlertStore()

	// Add inhibiting alert with critical severity
	inhibitor := &types.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	alertStore.Put(inhibitor)

	pipeline := NewRunner(silenceStore, alertStore, []string{"default"})

	// Incoming alert with warning severity matches critical inhibitor
	labels := model.LabelSet{"severity": "critical"}
	outcome, err := pipeline.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.Equal(t, "inhibited", outcome.Status)
}
