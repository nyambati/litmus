package config

import (
	"fmt"
	"sort"
	"strings"

	amconfig "github.com/prometheus/alertmanager/config"
)

type Assembler struct {
	base *AlertmanagerConfig
}

func NewAssembler(base *AlertmanagerConfig) *Assembler {
	return &Assembler{base: base}
}

func (a *Assembler) Assemble(fragments []*Fragment) (*AlertmanagerConfig, error) {
	if a.base.Route == nil {
		a.base.Route = &amconfig.Route{}
	}

	type groupEntry struct {
		route    *amconfig.Route
		receiver string
	}
	grouped := map[string]*groupEntry{}
	groupOrder := []string{}

	for _, frag := range fragments {
		a.applyNamespace(frag)

		a.base.Receivers = append(a.base.Receivers, frag.Receivers...)
		a.base.InhibitRules = append(a.base.InhibitRules, frag.InhibitRules...)

		if len(frag.Routes) == 0 {
			continue
		}

		if frag.Group == nil {
			a.base.Route.Routes = append(a.base.Route.Routes, frag.Routes...)
			continue
		}

		key := groupKey(frag.Group.Match)
		entry, exists := grouped[key]
		if !exists {
			entry = &groupEntry{
				route: &amconfig.Route{
					Match: frag.Group.Match,
				},
				receiver: frag.Group.Receiver,
			}
			grouped[key] = entry
			groupOrder = append(groupOrder, key)
		}

		if frag.Group.Receiver != "" {
			if entry.receiver == "" {
				entry.receiver = frag.Group.Receiver
			} else if entry.receiver != frag.Group.Receiver {
				return nil, fmt.Errorf(
					"group %q: conflicting receivers %q and %q",
					key, entry.receiver, frag.Group.Receiver,
				)
			}
		}

		entry.route.Routes = append(entry.route.Routes, frag.Routes...)
	}

	for _, key := range groupOrder {
		entry := grouped[key]
		if entry.receiver != "" {
			entry.route.Receiver = entry.receiver
		} else {
			entry.route.Receiver = a.base.Route.Receiver
		}
		a.base.Route.Routes = append(a.base.Route.Routes, entry.route)
	}

	return a.base, nil
}

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

	if frag.Group != nil && frag.Group.Receiver != "" &&
		!strings.HasPrefix(frag.Group.Receiver, prefix) {
		frag.Group.Receiver = prefix + frag.Group.Receiver
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

func groupKey(labels map[string]string) string {
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+labels[k])
	}
	return strings.Join(parts, ",")
}
