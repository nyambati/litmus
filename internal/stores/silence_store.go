package stores

import (
	"context"

	"github.com/nyambati/litmus/internal/types"
	"github.com/prometheus/common/model"
)

// SilenceStore holds silences and provides muting logic.
type SilenceStore struct {
	silences []types.Silence
}

// NewSilenceStore creates a new silence store with initial silences.
func NewSilenceStore(silences []types.Silence) *SilenceStore {
	return &SilenceStore{
		silences: silences,
	}
}

// Mutes returns true if any silence in the store matches all labels in the provided set.
func (s *SilenceStore) Mutes(ctx context.Context, labels model.LabelSet) bool {
	for _, silence := range s.silences {
		if s.silenceMatches(silence, labels) {
			return true
		}
	}
	return false
}

// silenceMatches checks if a silence matches all its labels in the label set.
// A silence with no labels never matches.
func (s *SilenceStore) silenceMatches(silence types.Silence, labels model.LabelSet) bool {
	if len(silence.Labels) == 0 {
		return false
	}
	for silenceKey, silenceValue := range silence.Labels {
		if labelValue, exists := labels[model.LabelName(silenceKey)]; !exists || string(labelValue) != silenceValue {
			return false
		}
	}
	return true
}
