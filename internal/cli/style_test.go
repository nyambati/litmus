package cli

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestStyler_DisabledIsPlain(t *testing.T) {
	s := &styler{enabled: false}
	require.Equal(t, "text", s.red("text"))
	require.Equal(t, "text", s.green("text"))
	require.Equal(t, "[FAIL]", s.status("FAIL"))
	require.Equal(t, "[OK]", s.status("OK"))
}

func TestStyler_EnabledWrapsWithANSI(t *testing.T) {
	s := &styler{enabled: true}
	require.Equal(t, ansiRed+"text"+ansiReset, s.red("text"))
	require.Equal(t, ansiGreen+"[PASS]"+ansiReset, s.status("PASS"))
	require.Equal(t, ansiYellow+"[WARN]"+ansiReset, s.status("WARN"))
	require.Equal(t, ansiDim+"[SKIP]"+ansiReset, s.status("SKIP"))
}

func TestStyler_StatusTokenAlwaysContainsBracketText(t *testing.T) {
	// Substring checks and column alignment rely on the bracket text being
	// present verbatim regardless of color state.
	for _, kind := range []string{"OK", "PASS", "FAIL", "WARN", "SKIP"} {
		require.Contains(t, (&styler{enabled: true}).status(kind), "["+kind+"]")
		require.Contains(t, (&styler{enabled: false}).status(kind), "["+kind+"]")
	}
}

func TestNewStyler_NoColorEnvDisables(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	// os.Stdout in the test harness is not a terminal, but NO_COLOR must keep
	// it disabled regardless.
	require.False(t, newStyler(nil).enabled)
}
