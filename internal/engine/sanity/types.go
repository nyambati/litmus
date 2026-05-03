package sanity

import (
	litconfig "github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
)

// CheckContext bundles all inputs a sanity check may need.
type CheckContext struct {
	Route           *amconfig.Route
	Receivers       map[string]*amconfig.Receiver
	Rules           []*amconfig.InhibitRule
	Policy          litconfig.PolicyConfig
	Fragments       []*fragment.Fragment
	RegressionState *types.RegressionState
}

// Check is the interface all sanity checks must implement.
type Check interface {
	// Name returns the config key for this check (e.g. "dead_receivers").
	Name() litconfig.SanityCheck
	// Run executes the check against ctx and returns issue strings.
	Run(CheckContext) []string
}
