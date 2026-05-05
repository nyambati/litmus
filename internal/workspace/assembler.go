package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	"github.com/nyambati/litmus/internal/utils"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/sirupsen/logrus"
)

const fieldRoutes = "routes"

// Assemble produces the complete *amconfig.Config in w.Root by reading the
// workspace and merging every child fragment under <dir>/fragments.
//
// When w.Logger is non-nil, Assemble emits structured debug entries
// describing each child as it is merged, plus a final summary. Errors
// include the offending fragment's directory for fast diagnosis.
func (w *Workspace) Assemble() error {
	log := w.logger

	// Reset accumulated state so Assemble is idempotent on the same instance.
	w.Fragments = nil
	w.config = nil

	meta, err := w.read()
	if err != nil {
		return err
	}
	log.WithFields(logrus.Fields{
		"dir":        meta.Dir,
		"base":       meta.BaseFile,
		"root_tests": len(meta.TestFiles),
	}).Debug("workspace read")

	childrenDir := filepath.Join(w.dir, "fragments")
	children, err := fragment.Load(childrenDir)
	if err != nil {
		// No fragments directory is valid — treat as empty.
		if errors.Is(err, os.ErrNotExist) {
			children = &fragment.LoadResult{}
		} else {
			return fmt.Errorf("load children dir=%q: %w", childrenDir, err)
		}
	}
	log.WithField("count", len(children.Fragments)).Debug("children loaded")

	for _, ch := range children.Fragments {
		w.Fragments = append(w.Fragments, ch.Fragment)
	}

	return assemble(w.config, children.Fragments, log)
}

// assemble is the pure merge function: given a root config and a slice of
// augmented fragments it mutates root in place. No I/O — directly testable.
// Metadata is preserved so log entries and errors identify the offending
// fragment by directory.
func assemble(root *types.AlertmanagerConfig, children []fragment.AugmentedFragment, log logrus.FieldLogger) error {
	if root.Route == nil {
		root.Route = &amconfig.Route{}
	}

	groups := newGroupSet()
	for _, ch := range children {
		frag := ch.Fragment
		dir := metaDir(ch.Metadata)
		entry := log.WithFields(logrus.Fields{
			"child":     frag.Namespace,
			"dir":       dir,
			"files":     metaFiles(ch.Metadata),
			"receivers": len(frag.Receivers),
			fieldRoutes: len(frag.Routes),
			"inhibit":   len(frag.InhibitRules),
		})
		entry.Debug("merging child")

		applyNamespace(frag)

		mergeFlat(root, frag)

		if len(frag.Routes) == 0 {
			continue
		}
		if frag.Group == nil {
			root.Route.Routes = append(root.Route.Routes, frag.Routes...)
			entry.WithField(fieldRoutes, len(frag.Routes)).Debug("attached routes directly to root")
			continue
		}
		if err := groups.add(frag); err != nil {
			return fmt.Errorf("group child %q (dir=%s): %w", frag.Namespace, dir, err)
		}
		entry.WithFields(logrus.Fields{
			fieldRoutes: len(frag.Routes),
			"match":     frag.Group.Match,
		}).Debug("grouped routes")
	}

	groups.routes(root, log)
	return nil
}

func metaDir(m *fragment.Metadata) string {
	if m == nil {
		return ""
	}
	return m.Dir
}

func metaFiles(m *fragment.Metadata) int {
	if m == nil {
		return 0
	}
	return len(m.Files)
}

// mergeFlat copies fields that always concatenate without coordination across
// fragments — receivers, inhibit rules, time intervals.
func mergeFlat(root *types.AlertmanagerConfig, frag *fragment.Fragment) {
	root.Receivers = append(root.Receivers, frag.Receivers...)
	root.InhibitRules = append(root.InhibitRules, frag.InhibitRules...)
	root.MuteTimeIntervals = append(root.MuteTimeIntervals, frag.MuteTimeIntervals...)
	root.TimeIntervals = append(root.TimeIntervals, frag.TimeIntervals...)
}

func newGroupSet() *groupSet {
	return &groupSet{entries: map[string]*groupEntry{}}
}

func (g *groupSet) add(frag *fragment.Fragment) error {
	key := utils.LabelFormat(frag.Group.Match)
	entry, ok := g.entries[key]
	if !ok {
		entry = &groupEntry{
			route:    &amconfig.Route{Match: frag.Group.Match},
			receiver: frag.Group.Receiver,
		}
		g.entries[key] = entry
		g.order = append(g.order, key)
	} else if frag.Group.Receiver != "" {
		switch {
		case entry.receiver == "":
			entry.receiver = frag.Group.Receiver
		case entry.receiver != frag.Group.Receiver:
			return fmt.Errorf("group %q: conflicting receivers %q vs %q",
				key, entry.receiver, frag.Group.Receiver)
		}
	}
	entry.route.Routes = append(entry.route.Routes, frag.Routes...)
	return nil
}

func (g *groupSet) routes(root *types.AlertmanagerConfig, log logrus.FieldLogger) {
	out := make([]*amconfig.Route, 0, len(g.order))
	for _, key := range g.order {
		entry := g.entries[key]
		entry.route.Receiver = entry.receiver
		if entry.route.Receiver == "" {
			entry.route.Receiver = root.Route.Receiver
		}
		out = append(out, entry.route)
	}
	root.Route.Routes = append(root.Route.Routes, out...)
	log.WithFields(logrus.Fields{
		"receivers":  len(root.Receivers),
		fieldRoutes:  len(root.Route.Routes),
		"group_subs": len(out),
		"inhibit":    len(root.InhibitRules),
	}).Debug("workspace assembled")
}

func applyNamespace(frag *fragment.Fragment) {
	if frag.Namespace == "" {
		return
	}
	prefix := frag.Namespace + "-"

	for _, r := range frag.Receivers {
		if r.Name != "" && !strings.HasPrefix(r.Name, prefix) {
			r.Name = prefix + r.Name
		}
	}
	for _, route := range frag.Routes {
		prefixRouteReceivers(route, prefix)
	}
	if frag.Group != nil && frag.Group.Receiver != "" && !strings.HasPrefix(frag.Group.Receiver, prefix) {
		frag.Group.Receiver = prefix + frag.Group.Receiver
	}
}

func prefixRouteReceivers(r *amconfig.Route, prefix string) {
	if r == nil {
		return
	}
	if r.Receiver != "" && !strings.HasPrefix(r.Receiver, prefix) {
		r.Receiver = prefix + r.Receiver
	}
	for _, child := range r.Routes {
		prefixRouteReceivers(child, prefix)
	}
}
