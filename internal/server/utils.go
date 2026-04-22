package server

import (
	"fmt"

	"github.com/nyambati/litmus/internal/engine/pipeline"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/prometheus/common/model"
)

func traceRoute(route *amconfig.Route, labels model.LabelSet) *RouteNode {
	node := &RouteNode{
		Receiver: route.Receiver,
		Continue: route.Continue,
		Matched:  pipeline.RouteMatches(route, labels),
	}

	// Capture Grouping and Timing
	for _, l := range route.GroupBy {
		node.GroupBy = append(node.GroupBy, string(l))
	}
	if route.GroupWait != nil {
		node.GroupWait = route.GroupWait.String()
	}
	if route.GroupInterval != nil {
		node.GroupInterval = route.GroupInterval.String()
	}
	if route.RepeatInterval != nil {
		node.RepeatInterval = route.RepeatInterval.String()
	}

	// Capture matchers for UI display
	for k, v := range route.Match {
		node.Match = append(node.Match, fmt.Sprintf("%s=%q", k, v))
	}
	for k, re := range route.MatchRE {
		node.Match = append(node.Match, fmt.Sprintf("%s=~%q", k, re.String()))
	}
	for _, m := range route.Matchers {
		node.Match = append(node.Match, m.String())
	}

	if !node.Matched {
		return node
	}

	for _, child := range route.Routes {
		childNode := traceRoute(child, labels)
		if childNode.Matched {
			node.Children = append(node.Children, childNode)
			if !childNode.Continue {
				break
			}
		}
	}

	return node
}

// flattenMatchedPath walks a RouteNode tree and returns each matched node as a flat list.
func flattenMatchedPath(node *RouteNode) []*RouteNode {
	if node == nil {
		return nil
	}
	var steps []*RouteNode
	if node.Matched {
		steps = append(steps, &RouteNode{Receiver: node.Receiver, Match: node.Match})
	}
	for _, child := range node.Children {
		steps = append(steps, flattenMatchedPath(child)...)
	}
	return steps
}

// findRoutesByReceiver walks the route tree and returns every route whose receiver equals the target.
func findRoutesByReceiver(route *amconfig.Route, receiver string) []*amconfig.Route {
	if route == nil {
		return nil
	}
	var found []*amconfig.Route
	if route.Receiver == receiver {
		found = append(found, route)
	}
	for _, child := range route.Routes {
		found = append(found, findRoutesByReceiver(child, receiver)...)
	}
	return found
}

// identifyMatcherFailures returns matchers on the given route that do not match the label set.
func identifyMatcherFailures(route *amconfig.Route, labels model.LabelSet) []MatcherMismatch {
	if route == nil {
		return nil
	}
	var out []MatcherMismatch

	for k, v := range route.Match {
		actual := string(labels[model.LabelName(k)])
		if actual != v {
			out = append(out, MatcherMismatch{Label: k, Required: v, Actual: actual})
		}
	}

	for k, re := range route.MatchRE {
		actual := string(labels[model.LabelName(k)])
		if !re.MatchString(actual) {
			out = append(out, MatcherMismatch{Label: k, Required: "~" + re.String(), Actual: actual})
		}
	}

	for _, m := range route.Matchers {
		actual := string(labels[model.LabelName(m.Name)])
		if !m.Matches(actual) {
			out = append(out, MatcherMismatch{Label: m.Name, Required: m.Value, Actual: actual})
		}
	}

	return out
}
