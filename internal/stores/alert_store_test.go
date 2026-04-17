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

func TestAlertStore_GetPending(t *testing.T) {
	store := NewAlertStore()

	alert := &types.Alert{
		Alert: model.Alert{
			Labels:   model.LabelSet{"service": "api", "severity": "critical"},
			StartsAt: time.Now(),
		},
	}

	err := store.Put(alert)
	require.NoError(t, err)

	iter := store.GetPending()
	require.NotNil(t, iter)
	defer iter.Close()

	received := <-iter.Next()
	require.NotNil(t, received)
	require.Equal(t, alert.Labels, received.Labels)
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

	store.Reset()

	iter := store.GetPending()
	defer iter.Close()
	ch := iter.Next()
	_, ok := <-ch
	require.False(t, ok)
}

func TestAlertIterator_CloseSafe(t *testing.T) {
	store := NewAlertStore()
	iter := store.GetPending()

	// Double-close must not panic
	require.NotPanics(t, func() {
		iter.Close()
		iter.Close()
	})
}
