package stores

import (
	"testing"
	"time"

	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/require"
)

func TestAlertStore_Put(t *testing.T) {
	store := NewAlertStore()

	alert := &types.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}

	err := store.Put(alert)
	require.NoError(t, err)
}

func TestAlertStore_Get(t *testing.T) {
	store := NewAlertStore()

	now := time.Now()
	alert := &types.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: now,
		},
	}

	err := store.Put(alert)
	require.NoError(t, err)

	fp := alert.Fingerprint()
	retrieved, err := store.Get(fp)
	require.NoError(t, err)
	require.NotNil(t, retrieved)
	require.Equal(t, alert.Labels, retrieved.Labels)
}

func TestAlertStore_Subscribe(t *testing.T) {
	store := NewAlertStore()

	alert := &types.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api"},
			StartsAt: time.Now(),
		},
	}

	err := store.Put(alert)
	require.NoError(t, err)

	iter := store.Subscribe()
	require.NotNil(t, iter)

	alertChan := iter.Next()
	require.NotNil(t, alertChan)

	iter.Close()
}

func TestAlertStore_Reset(t *testing.T) {
	store := NewAlertStore()

	alert := &types.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api"},
			StartsAt: time.Now(),
		},
	}

	err := store.Put(alert)
	require.NoError(t, err)

	fp := alert.Fingerprint()
	_, err = store.Get(fp)
	require.NoError(t, err)

	store.Reset()

	_, err = store.Get(fp)
	require.Error(t, err)
}
