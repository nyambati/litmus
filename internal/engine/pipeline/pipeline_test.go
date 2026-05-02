package pipeline

import (
	"context"
	"testing"
	"time"

	"github.com/nyambati/litmus/internal/stores"
	lithtypes "github.com/nyambati/litmus/internal/types"

	amconfig "github.com/prometheus/alertmanager/config"
	alertmgrtypes "github.com/prometheus/alertmanager/types"
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
	require.Equal(t, StatusSilenced, outcome.Status)
}

func TestPipeline_Execute_Active(t *testing.T) {
	silenceStore := stores.NewSilenceStore([]lithtypes.Silence{})
	alertStore := stores.NewAlertStore()
	router := NewRouter(&amconfig.Route{Receiver: "default"})

	pipeline := NewRunner(silenceStore, alertStore, router, nil)

	labels := model.LabelSet{"service": "api", "severity": "critical"}
	outcome, err := pipeline.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.Equal(t, StatusActive, outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestPipeline_Execute_Inhibited(t *testing.T) {
	silenceStore := stores.NewSilenceStore([]lithtypes.Silence{})
	alertStore := stores.NewAlertStore()

	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := alertStore.Put(inhibitor)
	require.NoError(t, err)

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
	require.Equal(t, StatusInhibited, outcome.Status)
}

func TestExecute_NilRunner(t *testing.T) {
	var r *Runner
	labels := model.LabelSet{"service": "api"}

	outcome, err := r.Execute(context.Background(), labels)

	require.Error(t, err)
	require.Nil(t, outcome)
	require.Contains(t, err.Error(), "runner is nil")
}

func TestExecute_NilSilenceStore(t *testing.T) {
	r := NewRunner(nil, stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), nil)
	labels := model.LabelSet{"service": "api"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusActive, outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestExecute_NilAlertStore(t *testing.T) {
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), nil, NewRouter(&amconfig.Route{Receiver: "default"}), nil)
	labels := model.LabelSet{"service": "api"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusActive, outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestExecute_NilRouter(t *testing.T) {
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), nil, nil)
	labels := model.LabelSet{"service": "api"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusActive, outcome.Status)
	require.Empty(t, outcome.Receivers)
}

func TestExecute_InhibitRulesNil(t *testing.T) {
	r := NewRunner(stores.NewSilenceStore(
		[]lithtypes.Silence{}),
		stores.NewAlertStore(),
		NewRouter(&amconfig.Route{Receiver: "default"}),
		nil,
	)
	labels := model.LabelSet{"service": "api"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusActive, outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestExecute_InhibitRulesEmpty(t *testing.T) {
	r := NewRunner(
		stores.NewSilenceStore([]lithtypes.Silence{}),
		stores.NewAlertStore(),
		NewRouter(&amconfig.Route{Receiver: "default"}),
		[]amconfig.InhibitRule{},
	)
	labels := model.LabelSet{"service": "api"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusActive, outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestExecute_InhibitRules_NilSourceMatchMap(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: nil,
			TargetMatch: map[string]string{"severity": "warning"},
			Equal:       model.LabelNames{"service"},
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add an active alert
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	labels := model.LabelSet{"service": "api", "severity": "warning"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusInhibited, outcome.Status)
}

func TestExecute_InhibitRules_NilTargetMatchMap(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"severity": "critical"},
			TargetMatch: nil,
			Equal:       model.LabelNames{"service"},
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add an active alert
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	labels := model.LabelSet{"service": "api", "severity": "warning"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusInhibited, outcome.Status)
}

func TestExecute_InhibitRules_EmptySourceMatchMap(t *testing.T) {
	// Empty SourceMatch matches everything, but TargetMatch must also match target
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{}, // Empty matches everything
			TargetMatch: map[string]string{"severity": "warning"},
			Equal:       model.LabelNames{"service"},
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add an active alert - source alert doesn't have severity=warning, so it won't match target
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "db", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	// Target has severity=warning but source alert's target match is severity=critical not warning
	labels := model.LabelSet{"service": "api", "severity": "warning"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	// Should be active because the source alert doesn't match the target match condition
	require.Equal(t, StatusActive, outcome.Status)
}

func TestExecute_InhibitRules_NilEqual(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"service": "api"},
			TargetMatch: map[string]string{"service": "api"},
			Equal:       nil, // Nil means no equality requirement
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add an active alert
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	// Same service but different severity should still inhibit
	labels := model.LabelSet{"service": "api", "severity": "warning"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusInhibited, outcome.Status)
}

func TestExecute_InhibitRules_EmptyEqual(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"service": "api"},
			TargetMatch: map[string]string{"service": "api"},
			Equal:       model.LabelNames{}, // Empty - no equality requirement
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add an active alert
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	// Same service but different severity should inhibit
	labels := model.LabelSet{"service": "api", "severity": "warning"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusInhibited, outcome.Status)
}

func TestExecute_InhibitRules_NoInhibition(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"service": "api"},
			TargetMatch: map[string]string{"service": "api"},
			Equal:       model.LabelNames{"service"},
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add an active alert but it doesn't match source
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "db", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	// Target should not be inhibited because source doesn't match
	labels := model.LabelSet{"service": "api", "severity": "warning"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusActive, outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestExecute_MultipleInhibitionRules(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"severity": "critical"},
			TargetMatch: map[string]string{"severity": "warning"},
			Equal:       model.LabelNames{"service"},
		},
		{
			SourceMatch: map[string]string{"severity": "critical"},
			TargetMatch: map[string]string{"severity": "info"},
			Equal:       model.LabelNames{"service"},
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add a critical alert
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	// Both warning and info should be inhibited
	tests := []struct {
		name       string
		labels     model.LabelSet
		wantStatus string
	}{
		{"warning inhibited", model.LabelSet{"service": "api", "severity": "warning"}, StatusInhibited},
		{"info inhibited", model.LabelSet{"service": "api", "severity": "info"}, StatusInhibited},
		{"critical not inhibited", model.LabelSet{"service": "api", "severity": "critical"}, StatusActive},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fresh alert store for each test
			r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

			inhibitor := &alertmgrtypes.Alert{
				Alert: model.Alert{
					Labels:   model.LabelSet{"service": "api", "severity": "critical"},
					StartsAt: time.Now(),
				},
			}
			err := r.alertStore.Put(inhibitor)
			require.NoError(t, err)

			outcome, err := r.Execute(context.Background(), tt.labels)

			require.NoError(t, err)
			require.Equal(t, tt.wantStatus, outcome.Status)
		})
	}
}

func TestExecute_SilencedBeforeInhibited(t *testing.T) {
	silences := []lithtypes.Silence{
		{
			Labels:  map[string]string{"service": "api"},
			Comment: "maintenance",
		},
	}

	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"severity": "critical"},
			TargetMatch: map[string]string{"severity": "warning"},
			Equal:       model.LabelNames{"service"},
		},
	}

	alertStore := stores.NewAlertStore()
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}
	err := alertStore.Put(inhibitor)
	require.NoError(t, err)

	r := NewRunner(stores.NewSilenceStore(silences), alertStore, NewRouter(&amconfig.Route{Receiver: "default"}), rules)
	labels := model.LabelSet{"service": "api", "severity": "warning"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.Equal(t, StatusSilenced, outcome.Status) // Silence takes precedence over inhibition
}

func TestExecute_EmptyLabels(t *testing.T) {
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), nil)
	labels := model.LabelSet{}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusActive, outcome.Status)
	require.Contains(t, outcome.Receivers, "default")
}

func TestExecute_SpecialCharactersInLabels(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"service": "api-gateway"},
			TargetMatch: map[string]string{"service": "api-gateway"},
			Equal:       model.LabelNames{"env"},
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add an active alert with special characters
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api-gateway", "env": "prod"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	// Target with same service and env should be inhibited
	labels := model.LabelSet{"service": "api-gateway", "env": "prod"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusInhibited, outcome.Status)
}

func TestExecute_LabelsWithDots(t *testing.T) {
	rules := []amconfig.InhibitRule{
		{
			SourceMatch: map[string]string{"service": "my.app"},
			TargetMatch: map[string]string{"service": "my.app"},
			Equal:       model.LabelNames{"region"},
		},
	}
	r := NewRunner(stores.NewSilenceStore([]lithtypes.Silence{}), stores.NewAlertStore(), NewRouter(&amconfig.Route{Receiver: "default"}), rules)

	// Add an active alert with dot in label value
	inhibitor := &alertmgrtypes.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "my.app", "region": "us-east"},
			StartsAt: time.Now(),
		},
	}
	err := r.alertStore.Put(inhibitor)
	require.NoError(t, err)

	labels := model.LabelSet{"service": "my.app", "region": "us-east"}

	outcome, err := r.Execute(context.Background(), labels)

	require.NoError(t, err)
	require.NotNil(t, outcome)
	require.Equal(t, StatusInhibited, outcome.Status)
}
