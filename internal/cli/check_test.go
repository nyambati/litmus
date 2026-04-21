package cli

import (
	"testing"

	"github.com/nyambati/litmus/internal/types"
	"github.com/stretchr/testify/require"
)

func TestFilterByTags_NoTags(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
	}

	result := filterByTags(tests, []string{})
	require.Equal(t, tests, result)
}

func TestFilterByTags_SingleTag(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
		{Name: "test3", Tags: []string{"critical", "smoke"}},
	}

	result := filterByTags(tests, []string{"critical"})
	require.Len(t, result, 2)
	require.Equal(t, "test1", result[0].Name)
	require.Equal(t, "test3", result[1].Name)
}

func TestFilterByTags_MultipleTagsOr(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
		{Name: "test3", Tags: []string{"critical", "smoke"}},
		{Name: "test4", Tags: []string{"e2e"}},
	}

	result := filterByTags(tests, []string{"critical", "smoke"})
	require.Len(t, result, 3)
	require.Equal(t, "test1", result[0].Name)
	require.Equal(t, "test2", result[1].Name)
	require.Equal(t, "test3", result[2].Name)
}

func TestFilterByTags_NoMatches(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{"critical"}},
		{Name: "test2", Tags: []string{"smoke"}},
	}

	result := filterByTags(tests, []string{"nonexistent"})
	require.Len(t, result, 0)
}

func TestFilterByTags_NoTagsOnTest(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: []string{}},
		{Name: "test2", Tags: []string{"critical"}},
	}

	result := filterByTags(tests, []string{"critical"})
	require.Len(t, result, 1)
	require.Equal(t, "test2", result[0].Name)
}

func TestFilterByTags_NilTagsOnTest(t *testing.T) {
	tests := []*types.TestCase{
		{Name: "test1", Tags: nil},
		{Name: "test2", Tags: []string{"critical"}},
	}

	result := filterByTags(tests, []string{"critical"})
	require.Len(t, result, 1)
	require.Equal(t, "test2", result[0].Name)
}
