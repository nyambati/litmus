package workspace

import (
	"bytes"
	"io"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/sirupsen/logrus"
)

const (
	recvDBCritical = "db-critical"
	recvDBFallback = "db-fallback"
)

func nopLogger() logrus.FieldLogger {
	l := logrus.New()
	l.Out = io.Discard
	return l
}

func bufLogger(buf *bytes.Buffer) logrus.FieldLogger {
	l := logrus.New()
	l.Out = buf
	l.Level = logrus.DebugLevel
	l.Formatter = &logrus.TextFormatter{DisableTimestamp: true, DisableColors: true}
	return l
}

func augment(frags []*fragment.Fragment) []fragment.AugmentedFragment {
	out := make([]fragment.AugmentedFragment, 0, len(frags))
	for _, f := range frags {
		out = append(out, fragment.AugmentedFragment{
			Fragment: f,
			Metadata: &fragment.Metadata{Dir: filepath.Join("/test", f.Namespace)},
		})
	}
	return out
}

// --- assemble ---

func TestAssemble_FlatMergeToRoot(t *testing.T) {
	root := &types.AlertmanagerConfig{Route: &amconfig.Route{Receiver: "default"}}
	frags := []*fragment.Fragment{
		{Namespace: "db", Routes: []*amconfig.Route{{Receiver: recvDBCritical}}},
	}
	if err := assemble(root, augment(frags), nopLogger()); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if len(root.Route.Routes) != 1 {
		t.Fatalf("Route.Routes length = %d, want 1", len(root.Route.Routes))
	}
	if root.Route.Routes[0].Receiver != recvDBCritical {
		t.Errorf("Routes[0].Receiver = %q, want %q", root.Route.Routes[0].Receiver, recvDBCritical)
	}
}

func TestAssemble_GroupedSingleFragment(t *testing.T) {
	root := &types.AlertmanagerConfig{Route: &amconfig.Route{Receiver: "default"}}
	frags := []*fragment.Fragment{
		{
			Namespace: "db",
			Group:     &fragment.FragmentGroup{Match: map[string]string{"scope": "teams"}},
			Routes:    []*amconfig.Route{{Receiver: recvDBCritical}},
		},
	}
	if err := assemble(root, augment(frags), nopLogger()); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if len(root.Route.Routes) != 1 {
		t.Fatalf("Route.Routes length = %d, want 1 (synthetic parent)", len(root.Route.Routes))
	}
	parent := root.Route.Routes[0]
	wantMatch := map[string]string{"scope": "teams"}
	if !reflect.DeepEqual(parent.Match, wantMatch) {
		t.Errorf("parent.Match = %v, want %v", parent.Match, wantMatch)
	}
	// No explicit group receiver → inherits root receiver.
	if parent.Receiver != "default" {
		t.Errorf("parent.Receiver = %q, want %q (inherits root)", parent.Receiver, "default")
	}
	if len(parent.Routes) != 1 || parent.Routes[0].Receiver != recvDBCritical {
		t.Errorf("parent.Routes[0].Receiver = %q, want %q", parent.Routes[0].Receiver, recvDBCritical)
	}
}

func TestAssemble_GroupReceiverExplicit(t *testing.T) {
	root := &types.AlertmanagerConfig{Route: &amconfig.Route{Receiver: "default"}}
	frags := []*fragment.Fragment{
		{
			Namespace: "db",
			Group:     &fragment.FragmentGroup{Match: map[string]string{"scope": "teams"}, Receiver: "fallback"},
			Routes:    []*amconfig.Route{{Receiver: "critical"}},
		},
	}
	if err := assemble(root, augment(frags), nopLogger()); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	// Namespace "db" is applied to group receiver: "fallback" → recvDBFallback
	if got := root.Route.Routes[0].Receiver; got != recvDBFallback {
		t.Errorf("parent.Receiver = %q, want %q", got, recvDBFallback)
	}
}

func TestAssemble_TwoFragmentsSameGroup(t *testing.T) {
	root := &types.AlertmanagerConfig{Route: &amconfig.Route{Receiver: "default"}}
	frags := []*fragment.Fragment{
		{
			Namespace: "db",
			Group:     &fragment.FragmentGroup{Match: map[string]string{"scope": "teams"}},
			Routes:    []*amconfig.Route{{Receiver: recvDBCritical}},
		},
		{
			Namespace: "net",
			Group:     &fragment.FragmentGroup{Match: map[string]string{"scope": "teams"}},
			Routes:    []*amconfig.Route{{Receiver: "net-critical"}},
		},
	}
	if err := assemble(root, augment(frags), nopLogger()); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if len(root.Route.Routes) != 1 {
		t.Fatalf("Route.Routes length = %d, want 1 (single synthetic parent)", len(root.Route.Routes))
	}
	if len(root.Route.Routes[0].Routes) != 2 {
		t.Errorf("parent.Routes length = %d, want 2", len(root.Route.Routes[0].Routes))
	}
}

func TestAssemble_GroupReceiverConflictErrors(t *testing.T) {
	root := &types.AlertmanagerConfig{Route: &amconfig.Route{Receiver: "default"}}
	frags := []*fragment.Fragment{
		{
			Namespace: "db",
			Group:     &fragment.FragmentGroup{Match: map[string]string{"scope": "teams"}, Receiver: "fallback-a"},
			Routes:    []*amconfig.Route{{Receiver: recvDBCritical}},
		},
		{
			Namespace: "net",
			Group:     &fragment.FragmentGroup{Match: map[string]string{"scope": "teams"}, Receiver: "fallback-b"},
			Routes:    []*amconfig.Route{{Receiver: "net-critical"}},
		},
	}
	err := assemble(root, augment(frags), nopLogger())
	if err == nil {
		t.Fatal("assemble = nil, want conflicting receivers error")
	}
	if !strings.Contains(err.Error(), "conflicting receivers") {
		t.Errorf("err = %q, want 'conflicting receivers' substring", err)
	}
}

func TestAssemble_ErrorIncludesFragmentDir(t *testing.T) {
	root := &types.AlertmanagerConfig{Route: &amconfig.Route{Receiver: "default"}}
	in := []fragment.AugmentedFragment{
		{
			Fragment: &fragment.Fragment{
				Namespace: "conflictA",
				Group:     &fragment.FragmentGroup{Match: map[string]string{"team": "x"}, Receiver: "alpha"},
				Routes:    []*amconfig.Route{{Receiver: "r1"}},
			},
			Metadata: &fragment.Metadata{Dir: "/ws/conflictA"},
		},
		{
			Fragment: &fragment.Fragment{
				Namespace: "conflictB",
				Group:     &fragment.FragmentGroup{Match: map[string]string{"team": "x"}, Receiver: "beta"},
				Routes:    []*amconfig.Route{{Receiver: "r2"}},
			},
			Metadata: &fragment.Metadata{Dir: "/ws/conflictB"},
		},
	}
	err := assemble(root, in, nopLogger())
	if err == nil {
		t.Fatal("assemble = nil, want conflict error")
	}
	if !strings.Contains(err.Error(), "/ws/conflictB") {
		t.Errorf("err = %q, want offending dir '/ws/conflictB'", err)
	}
}

func TestAssemble_InhibitRulesMerged(t *testing.T) {
	root := &types.AlertmanagerConfig{
		Route:        &amconfig.Route{Receiver: "default"},
		InhibitRules: []amconfig.InhibitRule{{SourceMatch: map[string]string{"global": "rule"}}},
	}
	frags := []*fragment.Fragment{
		{
			Namespace:    "db",
			InhibitRules: []amconfig.InhibitRule{{SourceMatch: map[string]string{"team": "db"}}},
		},
	}
	if err := assemble(root, augment(frags), nopLogger()); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if len(root.InhibitRules) != 2 {
		t.Errorf("InhibitRules length = %d, want 2", len(root.InhibitRules))
	}
}

func TestAssemble_ReceiversMerged(t *testing.T) {
	root := &types.AlertmanagerConfig{
		Route:     &amconfig.Route{Receiver: "default"},
		Receivers: []*types.Receiver{{Name: "default"}},
	}
	frags := []*fragment.Fragment{
		{
			Namespace: "db",
			Receivers: []*types.Receiver{{Name: recvDBCritical}},
		},
	}
	if err := assemble(root, augment(frags), nopLogger()); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if len(root.Receivers) != 2 {
		t.Errorf("Receivers length = %d, want 2", len(root.Receivers))
	}
}

func TestAssemble_NilRootRouteInitialised(t *testing.T) {
	root := &types.AlertmanagerConfig{}
	if err := assemble(root, nil, nopLogger()); err != nil {
		t.Fatalf("assemble with nil Route = %v, want nil", err)
	}
	if root.Route == nil {
		t.Error("root.Route still nil after assemble, want initialised")
	}
}

func TestAssemble_LogsDebugFields(t *testing.T) {
	root := &types.AlertmanagerConfig{Route: &amconfig.Route{Receiver: "default"}}
	frag := &fragment.Fragment{Namespace: "db", Routes: []*amconfig.Route{{Receiver: recvDBCritical}}}
	in := []fragment.AugmentedFragment{
		{Fragment: frag, Metadata: &fragment.Metadata{Dir: "/ws/db", Files: []string{"a.yaml", "b.yaml"}}},
	}
	var buf bytes.Buffer
	if err := assemble(root, in, bufLogger(&buf)); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	out := buf.String()
	for _, want := range []string{`msg="merging child"`, `child=db`, `dir=/ws/db`, `files=2`, `msg="workspace assembled"`} {
		if !strings.Contains(out, want) {
			t.Errorf("log output missing %q\nfull output:\n%s", want, out)
		}
	}
}

// --- applyNamespace ---

func TestApplyNamespace_PrefixesReceivers(t *testing.T) {
	frag := &fragment.Fragment{
		Namespace: "db",
		Receivers: []*types.Receiver{{Name: "critical"}, {Name: "warning"}},
	}
	applyNamespace(frag)
	for _, r := range frag.Receivers {
		if !strings.HasPrefix(r.Name, "db-") {
			t.Errorf("Receiver.Name = %q, want prefix 'db-'", r.Name)
		}
	}
}

func TestApplyNamespace_PrefixesRouteReceivers(t *testing.T) {
	frag := &fragment.Fragment{
		Namespace: "db",
		Routes:    []*amconfig.Route{{Receiver: "critical"}},
	}
	applyNamespace(frag)
	if got := frag.Routes[0].Receiver; got != recvDBCritical {
		t.Errorf("Route.Receiver = %q, want %q", got, recvDBCritical)
	}
}

func TestApplyNamespace_PrefixesGroupReceiver(t *testing.T) {
	frag := &fragment.Fragment{
		Namespace: "db",
		Group:     &fragment.FragmentGroup{Receiver: "fallback"},
	}
	applyNamespace(frag)
	if got := frag.Group.Receiver; got != recvDBFallback {
		t.Errorf("Group.Receiver = %q, want %q", got, recvDBFallback)
	}
}

func TestApplyNamespace_NoDoublePrefixOnReceivers(t *testing.T) {
	frag := &fragment.Fragment{
		Namespace: "db",
		Receivers: []*types.Receiver{{Name: recvDBCritical}},
		Routes:    []*amconfig.Route{{Receiver: recvDBCritical}},
		Group:     &fragment.FragmentGroup{Receiver: recvDBFallback},
	}
	applyNamespace(frag)
	if got := frag.Receivers[0].Name; got != recvDBCritical {
		t.Errorf("Receiver already prefixed should not be doubled: got %q", got)
	}
	if got := frag.Routes[0].Receiver; got != recvDBCritical {
		t.Errorf("Route already prefixed should not be doubled: got %q", got)
	}
	if got := frag.Group.Receiver; got != recvDBFallback {
		t.Errorf("Group.Receiver already prefixed should not be doubled: got %q", got)
	}
}

func TestApplyNamespace_EmptyNamespaceIsNoop(t *testing.T) {
	frag := &fragment.Fragment{
		Namespace: "",
		Receivers: []*types.Receiver{{Name: "critical"}},
		Routes:    []*amconfig.Route{{Receiver: "critical"}},
	}
	applyNamespace(frag)
	if got := frag.Receivers[0].Name; got != "critical" {
		t.Errorf("Receiver.Name = %q, want unchanged %q", got, "critical")
	}
	if got := frag.Routes[0].Receiver; got != "critical" {
		t.Errorf("Route.Receiver = %q, want unchanged %q", got, "critical")
	}
}

func TestApplyNamespace_NestedRouteReceivers(t *testing.T) {
	child := &amconfig.Route{Receiver: "warning"}
	parent := &amconfig.Route{Receiver: "critical", Routes: []*amconfig.Route{child}}
	frag := &fragment.Fragment{
		Namespace: "db",
		Routes:    []*amconfig.Route{parent},
	}
	applyNamespace(frag)
	if got := frag.Routes[0].Receiver; got != recvDBCritical {
		t.Errorf("parent.Receiver = %q, want db-critical", got)
	}
	if got := frag.Routes[0].Routes[0].Receiver; got != "db-warning" {
		t.Errorf("child.Receiver = %q, want db-warning", got)
	}
}

// --- groupKey ---

func TestGroupKey_Deterministic(t *testing.T) {
	m := map[string]string{"team": "db", "scope": "infra"}
	first, second := groupKey(m), groupKey(m)
	if first != second {
		t.Errorf("groupKey = %q then %q, must be deterministic", first, second)
	}
}

func TestGroupKey_OrderIndependent(t *testing.T) {
	a := map[string]string{"team": "db", "scope": "infra"}
	b := map[string]string{"scope": "infra", "team": "db"}
	if groupKey(a) != groupKey(b) {
		t.Errorf("groupKey(%v) = %q != groupKey(%v) = %q; must be order-independent",
			a, groupKey(a), b, groupKey(b))
	}
}

func TestGroupKey_DifferentLabelsDifferentKeys(t *testing.T) {
	a := map[string]string{"team": "db"}
	b := map[string]string{"team": "payments"}
	if groupKey(a) == groupKey(b) {
		t.Errorf("different labels produced same key %q", groupKey(a))
	}
}

func TestGroupKey_EmptyMap(t *testing.T) {
	if k := groupKey(nil); k != "" {
		t.Errorf("groupKey(nil) = %q, want empty string", k)
	}
}

// --- mergeFlat ---

func TestMergeFlat_AppendsAllFields(t *testing.T) {
	root := &types.AlertmanagerConfig{
		Receivers:    []*types.Receiver{{Name: "default"}},
		InhibitRules: []amconfig.InhibitRule{{SourceMatch: map[string]string{"k": "v"}}},
	}
	frag := &fragment.Fragment{
		Receivers:    []*types.Receiver{{Name: "critical"}},
		InhibitRules: []amconfig.InhibitRule{{SourceMatch: map[string]string{"team": "db"}}},
	}
	mergeFlat(root, frag)
	if len(root.Receivers) != 2 {
		t.Errorf("Receivers length = %d, want 2", len(root.Receivers))
	}
	if len(root.InhibitRules) != 2 {
		t.Errorf("InhibitRules length = %d, want 2", len(root.InhibitRules))
	}
}

// --- namespace + assemble integration ---

func TestAssemble_NamespacePrefixesReceiversAndRoutes(t *testing.T) {
	root := &types.AlertmanagerConfig{Route: &amconfig.Route{Receiver: "default"}}
	frags := []*fragment.Fragment{
		{
			Namespace: "payments",
			Group: &fragment.FragmentGroup{
				Match:    map[string]string{"label_team": "payments"},
				Receiver: "fallback",
			},
			Routes: []*amconfig.Route{
				{Receiver: "critical"},
				{Receiver: "warning"},
			},
			Receivers: []*types.Receiver{
				{Name: "fallback"},
				{Name: "critical"},
				{Name: "warning"},
			},
		},
	}
	if err := assemble(root, augment(frags), nopLogger()); err != nil {
		t.Fatalf("assemble: %v", err)
	}
	if len(root.Route.Routes) != 1 {
		t.Fatalf("Route.Routes length = %d, want 1", len(root.Route.Routes))
	}
	parent := root.Route.Routes[0]
	if parent.Receiver != "payments-fallback" {
		t.Errorf("parent.Receiver = %q, want payments-fallback", parent.Receiver)
	}
	if len(parent.Routes) != 2 {
		t.Fatalf("parent.Routes length = %d, want 2", len(parent.Routes))
	}
	gotNames := make([]string, len(root.Receivers))
	for i, r := range root.Receivers {
		gotNames[i] = r.Name
	}
	sort.Strings(gotNames)
	want := []string{"payments-critical", "payments-fallback", "payments-warning"}
	if !reflect.DeepEqual(gotNames, want) {
		t.Errorf("Receiver names = %v, want %v", gotNames, want)
	}
}
