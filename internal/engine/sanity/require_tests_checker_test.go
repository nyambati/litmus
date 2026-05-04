package sanity

import (
	"testing"

	litconfig "github.com/nyambati/litmus/internal/config"
	"github.com/nyambati/litmus/internal/fragment"
	"github.com/nyambati/litmus/internal/types"
	amconfig "github.com/prometheus/alertmanager/config"
	"github.com/stretchr/testify/require"
)

func makeTestFrag(namespace string, routeCount int, testCount int) *fragment.Fragment {
	routes := make([]*amconfig.Route, routeCount)
	for i := range routes {
		routes[i] = &amconfig.Route{Receiver: "r"}
	}
	tests := make([]*types.TestCase, testCount)
	for i := range tests {
		tests[i] = &types.TestCase{}
	}
	return &fragment.Fragment{
		Namespace: namespace,
		Routes:    routes,
		Tests:     tests,
	}
}

func makeNestedFrag(namespace string, topLevel int, childrenPerRoute int, testCount int) *fragment.Fragment {
	routes := make([]*amconfig.Route, topLevel)
	for i := range routes {
		children := make([]*amconfig.Route, childrenPerRoute)
		for j := range children {
			children[j] = &amconfig.Route{Receiver: "child"}
		}
		routes[i] = &amconfig.Route{Receiver: "parent", Routes: children}
	}
	tests := make([]*types.TestCase, testCount)
	for i := range tests {
		tests[i] = &types.TestCase{}
	}
	return &fragment.Fragment{
		Namespace: namespace,
		Routes:    routes,
		Tests:     tests,
	}
}

func runRequireTests(frags []*fragment.Fragment, skipRoot bool) []string {
	skipRootTypes := []litconfig.PolicyType{}
	if skipRoot {
		skipRootTypes = append(skipRootTypes, litconfig.PolicyTypeTests)
	}
	ctx := CheckContext{
		Policy: litconfig.PolicyConfig{
			Require:  litconfig.RequireConfig{Tests: true},
			SkipRoot: skipRootTypes,
		},
		Fragments: frags,
	}
	return (&RequireTestsChecker{}).Run(ctx)
}

func TestRequireTestsChecker(t *testing.T) {
	tests := []struct {
		name      string
		frags     []*fragment.Fragment
		skipRoot  bool
		wantCount int
		wantFrag  string
	}{
		{
			name:      "no_routes_no_tests_passes",
			frags:     []*fragment.Fragment{{Namespace: "team-a"}},
			wantCount: 0,
		},
		{
			name:      "routes_no_tests_fails",
			frags:     []*fragment.Fragment{makeTestFrag("team-a", 3, 0)},
			wantCount: 1,
			wantFrag:  "team-a",
		},
		{
			name:      "tests_fewer_than_routes_fails",
			frags:     []*fragment.Fragment{makeTestFrag("team-a", 3, 2)},
			wantCount: 1,
			wantFrag:  "team-a",
		},
		{
			name:      "tests_equal_to_routes_passes",
			frags:     []*fragment.Fragment{makeTestFrag("team-a", 3, 3)},
			wantCount: 0,
		},
		{
			name:      "tests_more_than_routes_passes",
			frags:     []*fragment.Fragment{makeTestFrag("team-a", 3, 5)},
			wantCount: 0,
		},
		{
			name:      "single_route_single_test_passes",
			frags:     []*fragment.Fragment{makeTestFrag("team-a", 1, 1)},
			wantCount: 0,
		},
		{
			name:      "nested_routes_counted_recursively_fails",
			frags:     []*fragment.Fragment{makeNestedFrag("team-a", 1, 2, 2)}, // 3 total routes, 2 tests
			wantCount: 1,
			wantFrag:  "team-a",
		},
		{
			name:      "nested_routes_counted_recursively_passes",
			frags:     []*fragment.Fragment{makeNestedFrag("team-a", 1, 2, 3)}, // 3 total routes, 3 tests
			wantCount: 0,
		},
		{
			name: "multiple_frags_one_fails",
			frags: []*fragment.Fragment{
				makeTestFrag("team-a", 2, 2),
				makeTestFrag("team-b", 2, 1),
			},
			wantCount: 1,
			wantFrag:  "team-b",
		},
		{
			name: "multiple_frags_both_fail",
			frags: []*fragment.Fragment{
				makeTestFrag("team-a", 2, 1),
				makeTestFrag("team-b", 2, 0),
			},
			wantCount: 2,
		},
		{
			name:      "root_checked_by_default",
			frags:     []*fragment.Fragment{makeTestFrag("root", 2, 1)},
			skipRoot:  false,
			wantCount: 1,
			wantFrag:  "root",
		},
		{
			name:      "root_skipped_when_skip_root_set",
			frags:     []*fragment.Fragment{makeTestFrag("root", 2, 0)},
			skipRoot:  true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := runRequireTests(tt.frags, tt.skipRoot)
			require.Len(t, issues, tt.wantCount, "issues: %v", issues)
			if tt.wantFrag != "" {
				require.True(t, containsAny(issues, `"`+tt.wantFrag+`"`),
					"expected fragment %q in issues %v", tt.wantFrag, issues)
			}
		})
	}
}

func TestRequireTestsChecker_PolicyDisabled(t *testing.T) {
	ctx := CheckContext{
		Policy:    litconfig.PolicyConfig{Require: litconfig.RequireConfig{Tests: false}},
		Fragments: []*fragment.Fragment{makeTestFrag("team-a", 3, 0)},
	}
	require.Len(t, (&RequireTestsChecker{}).Run(ctx), 0)
}

func TestRequireTestsChecker_MessageFormat(t *testing.T) {
	issues := runRequireTests([]*fragment.Fragment{makeTestFrag("team-a", 3, 1)}, false)
	require.Len(t, issues, 1)
	require.Contains(t, issues[0], `"team-a"`)
	require.Contains(t, issues[0], "1") // test count
	require.Contains(t, issues[0], "3") // route count
}
