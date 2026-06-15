package cli

import (
	"os"
)

// ANSI escape codes used for terminal styling.
const (
	ansiReset  = "\033[0m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiDim    = "\033[2m"
	ansiBold   = "\033[1m"
)

// styler renders optionally-colored text. Color is enabled only when the
// destination is a real terminal and the NO_COLOR environment variable is
// unset (see https://no-color.org). This keeps CI logs and piped output clean.
type styler struct {
	enabled bool
}

// newStyler builds a styler for the given output file. Detection happens at
// call time so that redirected output (tests, pipes) correctly disables color.
func newStyler(f *os.File) *styler {
	return &styler{enabled: isTerminal(f) && os.Getenv("NO_COLOR") == ""}
}

func (s *styler) wrap(code, text string) string {
	if !s.enabled {
		return text
	}
	return code + text + ansiReset
}

// status returns a colored bracketed status token (e.g. "[PASS]"). The bracket
// text is preserved verbatim so substring checks and alignment still hold when
// color is disabled.
func (s *styler) status(kind string) string {
	token := "[" + kind + "]"
	switch kind {
	case "OK", "PASS":
		return s.green(token)
	case "FAIL":
		return s.red(token)
	case "WARN":
		return s.yellow(token)
	case "SKIP":
		return s.dim(token)
	default:
		return token
	}
}

func (s *styler) red(t string) string    { return s.wrap(ansiRed, t) }
func (s *styler) green(t string) string  { return s.wrap(ansiGreen, t) }
func (s *styler) yellow(t string) string { return s.wrap(ansiYellow, t) }
func (s *styler) dim(t string) string    { return s.wrap(ansiDim, t) }
func (s *styler) bold(t string) string   { return s.wrap(ansiBold, t) }

// isTerminal reports whether f is attached to a character device (a TTY).
func isTerminal(f *os.File) bool {
	if f == nil {
		return false
	}
	fi, err := f.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
