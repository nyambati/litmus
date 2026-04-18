package stores

import (
	"sync"

	"github.com/prometheus/alertmanager/provider"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

// AlertStore holds alerts in memory for testing.
type AlertStore struct {
	mu     sync.RWMutex
	alerts map[model.Fingerprint]*types.Alert
}

// NewAlertStore creates a new alert store.
func NewAlertStore() *AlertStore {
	return &AlertStore{
		alerts: make(map[model.Fingerprint]*types.Alert),
	}
}

// Put adds alerts to the store.
func (as *AlertStore) Put(alerts ...*types.Alert) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	for _, alert := range alerts {
		as.alerts[alert.Fingerprint()] = alert
	}
	return nil
}

// GetPending returns an iterator over all stored alerts.
// Creates a snapshot of alerts to prevent races with concurrent Put/Reset calls.
func (as *AlertStore) GetPending() provider.AlertIterator {
	as.mu.RLock()
	defer as.mu.RUnlock()

	alerts := make([]*types.Alert, 0, len(as.alerts))
	for _, alert := range as.alerts {
		// Copy alert pointer to snapshot to avoid stale pointer issues
		// if the underlying alert is modified or released
		alertCopy := *alert
		alerts = append(alerts, &alertCopy)
	}
	return newAlertIterator(alerts)
}

// Reset clears all alerts from the store.
func (as *AlertStore) Reset() {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.alerts = make(map[model.Fingerprint]*types.Alert)
}

// alertIterator implements provider.AlertIterator.
type alertIterator struct {
	mu     sync.Mutex
	once   sync.Once
	alerts []*types.Alert
	index  int
	done   chan struct{}
}

func newAlertIterator(alerts []*types.Alert) provider.AlertIterator {
	return &alertIterator{
		alerts: alerts,
		done:   make(chan struct{}),
	}
}

// Next returns a channel that yields alerts.
func (ai *alertIterator) Next() <-chan *types.Alert {
	ch := make(chan *types.Alert, 32)
	go func() {
		defer close(ch)
		for {
			ai.mu.Lock()
			if ai.index >= len(ai.alerts) {
				ai.mu.Unlock()
				return
			}
			alert := ai.alerts[ai.index]
			ai.index++
			ai.mu.Unlock()

			select {
			case <-ai.done:
				return
			case ch <- alert:
			}
		}
	}()
	return ch
}

// Close stops the iterator. Safe to call multiple times.
func (ai *alertIterator) Close() {
	ai.once.Do(func() { close(ai.done) })
}

// Err returns any error encountered during iteration.
func (ai *alertIterator) Err() error {
	return nil
}
