package config

import (
	"fmt"
	"strings"

	amconfig "github.com/prometheus/alertmanager/config"
)

// Assembler handles merging fragments into a base Alertmanager configuration.
type Assembler struct {
	base *amconfig.Config
}

// NewAssembler creates a new assembler with the given base configuration.
func NewAssembler(base *amconfig.Config) *Assembler {
	return &Assembler{base: base}
}

// Assemble merges the provided fragments into the base configuration.
func (a *Assembler) Assemble(fragments []*Fragment) (*amconfig.Config, error) {
	for _, frag := range fragments {
		a.applyNamespace(frag)

		// 1. Merge Receivers
		a.base.Receivers = append(a.base.Receivers, frag.Receivers...)

		// 2. Merge Inhibit Rules
		a.base.InhibitRules = append(a.base.InhibitRules, frag.InhibitRules...)

		// 3. Hierarchical Mounting of Routes
		if err := a.mountRoutes(frag); err != nil {
			return nil, fmt.Errorf("mounting routes for fragment %s: %w", frag.Name, err)
		}
	}

	return a.base, nil
}

// applyNamespace prefixes receiver names within the fragment if a namespace is defined.
func (a *Assembler) applyNamespace(frag *Fragment) {
	if frag.Namespace == "" {
		return
	}

	prefix := frag.Namespace + "-"

	for i := range frag.Receivers {
		if !strings.HasPrefix(frag.Receivers[i].Name, prefix) {
			frag.Receivers[i].Name = prefix + frag.Receivers[i].Name
		}
	}

	for _, route := range frag.Routes {
		a.prefixRouteReceivers(route, prefix)
	}
}

func (a *Assembler) prefixRouteReceivers(route *amconfig.Route, prefix string) {
	if route == nil {
		return
	}
	if route.Receiver != "" && !strings.HasPrefix(route.Receiver, prefix) {
		route.Receiver = prefix + route.Receiver
	}
	for _, child := range route.Routes {
		a.prefixRouteReceivers(child, prefix)
	}
}

// mountRoutes finds the anchor point in the base tree and attaches fragment routes.
func (a *Assembler) mountRoutes(frag *Fragment) error {
	if len(frag.Routes) == 0 {
		return nil
	}

	anchor := a.findAnchor(a.base.Route, frag.MountPoint)
	if anchor == nil {
		return fmt.Errorf("mount point %v not found in base routing tree", frag.MountPoint)
	}

	anchor.Routes = append(anchor.Routes, frag.Routes...)
	return nil
}

// findAnchor recursively searches for a route matching all mount point labels.
func (a *Assembler) findAnchor(root *amconfig.Route, mountPoint map[string]string) *amconfig.Route {
	if root == nil {
		return nil
	}

	if len(mountPoint) == 0 {
		return root // Default to root if no mount point specified
	}

	if a.matchesMountPoint(root, mountPoint) {
		return root
	}

	for _, child := range root.Routes {
		if found := a.findAnchor(child, mountPoint); found != nil {
			return found
		}
	}

	return nil
}

func (a *Assembler) matchesMountPoint(route *amconfig.Route, mountPoint map[string]string) bool {
	if len(route.Match) == 0 && len(route.MatchRE) == 0 && len(route.Matchers) == 0 {
		return false // Root route usually has no matchers, don't match it unless mountPoint is empty
	}

	for k, v := range mountPoint {
		matched := false
		if val, ok := route.Match[k]; ok && val == v {
			matched = true
		} else if val, ok := route.MatchRE[k]; ok && val.String() == v {
			matched = true
		} else {
			for _, m := range route.Matchers {
				if m.Name == k && m.Value == v {
					matched = true
					break
				}
			}
		}
		if !matched {
			return false
		}
	}
	return true
}
