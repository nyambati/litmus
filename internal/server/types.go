package server

// RouteNode represents a node in the route tree with evaluation details.
type RouteNode struct {
	Receiver       string       `json:"receiver,omitempty"`
	Match          []string     `json:"match,omitempty"`
	Matched        bool         `json:"matched"`
	Continue       bool         `json:"continue"`
	GroupBy        []string     `json:"group_by,omitempty"`
	GroupWait      string       `json:"group_wait,omitempty"`
	GroupInterval  string       `json:"group_interval,omitempty"`
	RepeatInterval string       `json:"repeat_interval,omitempty"`
	Children       []*RouteNode `json:"children,omitempty"`
}

// EvalResponse is the verbose result of an alert evaluation.
type EvalResponse struct {
	Labels      map[string]string `json:"labels"`
	Receivers   []string          `json:"receivers"`
	Path        *RouteNode        `json:"path"`
	Suppression *SuppressionInfo  `json:"suppression,omitempty"`
}

// SuppressionInfo holds details about why an alert might be silenced or inhibited.
type SuppressionInfo struct {
	Inhibited bool   `json:"inhibited"`
	Silenced  bool   `json:"silenced"`
	Reason    string `json:"reason,omitempty"`
}

// ConfigResponse returns metadata about the current workspace.
type ConfigResponse struct {
	ConfigPath string `json:"config_path"`
	Ready      bool   `json:"ready"`
}

// MatcherMismatch describes a single matcher in a route that no longer matches the label set.
type MatcherMismatch struct {
	Label    string `json:"label"`
	Required string `json:"required"` // value/pattern the current config expects
	Actual   string `json:"actual"`   // value present in the baseline label set
}

// RouteDrift describes why a specific expected receiver no longer matches.
type RouteDrift struct {
	Receiver   string            `json:"receiver"`
	Found      bool              `json:"found"`
	Mismatches []MatcherMismatch `json:"mismatches,omitempty"`
}

type deltaResult struct {
	Name       string            `json:"name"`
	Pass       bool              `json:"pass"`
	Kind       string            `json:"kind,omitempty"`
	Labels     map[string]string `json:"labels,omitempty"`
	Expected   []string          `json:"expected,omitempty"`
	Actual     []string          `json:"actual,omitempty"`
	RoutePath  []*RouteNode      `json:"route_path,omitempty"`
	WhyDrifted []RouteDrift      `json:"why_drifted,omitempty"`
}

type diffResponse struct {
	HasSnapshot bool           `json:"has_snapshot"`
	Total       int            `json:"total"`
	Passed      int            `json:"passed"`
	Drifted     int            `json:"drifted"`
	Results     []*deltaResult `json:"results"`
}
