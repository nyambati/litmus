package stores

import (
	"fmt"
	"sync"

	"github.com/prometheus/alertmanager/provider"
	"github.com/prometheus/alertmanager/types"
	"github.com/prometheus/common/model"
)

// AlertStore holds alerts in memory for testing.
type AlertStore struct {
	mu      sync.RWMutex
	alerts  map[model.Fingerprint]*types.Alert
	pending map[model.Fingerprint]*types.Alert
}

// NewAlertStore creates a new alert store.
func NewAlertStore() *AlertStore {
	return &AlertStore{
		alerts:  make(map[model.Fingerprint]*types.Alert),
		pending: make(map[model.Fingerprint]*types.Alert),
	}
}

// Put adds alerts to the store.
func (as *AlertStore) Put(alerts ...*types.Alert) error {
	as.mu.Lock()
	defer as.mu.Unlock()

	for _, alert := range alerts {
		fp := alert.Fingerprint()
		as.alerts[fp] = alert
		as.pending[fp] = alert
	}
	return nil
}

// Get retrieves an alert by fingerprint.
func (as *AlertStore) Get(fp model.Fingerprint) (*types.Alert, error) {
	as.mu.RLock()
	defer as.mu.RUnlock()

	alert, exists := as.alerts[fp]
	if !exists {
		return nil, fmt.Errorf("alert not found: %v", fp)
	}
	return alert, nil
}

// Subscribe returns an iterator over alerts.
func (as *AlertStore) Subscribe() provider.AlertIterator {
	as.mu.RLock()
	defer as.mu.RUnlock()

	alerts := make([]*types.Alert, 0, len(as.alerts))
	for _, alert := range as.alerts {
		alerts = append(alerts, alert)
	}
	return newAlertIterator(alerts)
}

// GetPending returns an iterator over pending alerts.
func (as *AlertStore) GetPending() provider.AlertIterator {
	as.mu.RLock()
	defer as.mu.RUnlock()

	alerts := make([]*types.Alert, 0, len(as.pending))
	for _, alert := range as.pending {
		alerts = append(alerts, alert)
	}
	return newAlertIterator(alerts)
}

// Reset clears all alerts from the store.
func (as *AlertStore) Reset() {
	as.mu.Lock()
	defer as.mu.Unlock()

	as.alerts = make(map[model.Fingerprint]*types.Alert)
	as.pending = make(map[model.Fingerprint]*types.Alert)
}

// alertIterator implements provider.AlertIterator.
type alertIterator struct {
	mu     sync.Mutex
	alerts []*types.Alert
	index  int
	done   chan struct{}
}

// newAlertIterator creates a new alert iterator.
func newAlertIterator(alerts []*types.Alert) provider.AlertIterator {
	return &alertIterator{
		alerts: alerts,
		index:  0,
		done:   make(chan struct{}),
	}
}

// Next returns a channel that yields alerts.
func (ai *alertIterator) Next() <-chan *types.Alert {
	ch := make(chan *types.Alert)
	go func() {
		defer close(ch)
		ai.mu.Lock()
		defer ai.mu.Unlock()

		for ai.index < len(ai.alerts) {
			select {
			case <-ai.done:
				return
			case ch <- ai.alerts[ai.index]:
				ai.index++
			}
		}
	}()
	return ch
}

// Close stops the iterator.
func (ai *alertIterator) Close() {
	close(ai.done)
}

// Err returns any error encountered during iteration.
func (ai *alertIterator) Err() error {
	return nil
}
