package cli

import (
	"github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/engine/sanity"
	"github.com/nyambati/litmus/internal/workspace"
	amconfig "github.com/prometheus/alertmanager/config"
)

// buildCheckContext constructs a sanity.CheckContext from an assembled
// alertmanager config, fragment list, and policy. Centralises the receiver-map
// and inhibit-rule-pointer construction that would otherwise be duplicated
// across every command that runs sanity checks.
func buildCheckContext(amCfg *amconfig.Config, ws *workspace.Workspace, policy config.PolicyConfig) sanity.CheckContext {
	receiversMap := make(map[string]*amconfig.Receiver, len(amCfg.Receivers))
	for i := range amCfg.Receivers {
		receiversMap[amCfg.Receivers[i].Name] = &amCfg.Receivers[i]
	}
	rules := make([]*amconfig.InhibitRule, 0, len(amCfg.InhibitRules))
	for i := range amCfg.InhibitRules {
		rules = append(rules, &amCfg.InhibitRules[i])
	}
	return sanity.CheckContext{
		Route:           amCfg.Route,
		Receivers:       receiversMap,
		Rules:           rules,
		Policy:          policy,
		Fragments:       ws.Fragments,
		RegressionState: ws.RegressionState,
	}
}
