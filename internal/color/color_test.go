package color

import (
	"os"
	"testing"

	"github.com/fatih/color"

	"github.com/karol-broda/snitch/internal/errutil"
)

func TestInit(t *testing.T) {
	testCases := []struct {
		name      string
		mode      string
		noColor   string
		term      string
		expected  bool
	}{
		{"Always", "always", "", "", false},
		{"Never", "never", "", "", true},
		{"Auto no env", "auto", "", "xterm-256color", false},
		{"Auto with NO_COLOR", "auto", "1", "xterm-256color", true},
		{"Auto with TERM=dumb", "auto", "", "dumb", true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Save original env vars
			origNoColor := os.Getenv("NO_COLOR")
			origTerm := os.Getenv("TERM")

			// Set test env vars
			errutil.Setenv("NO_COLOR", tc.noColor)
			errutil.Setenv("TERM", tc.term)

			Init(tc.mode)

			if color.NoColor != tc.expected {
				t.Errorf("Expected color.NoColor to be %v, but got %v", tc.expected, color.NoColor)
			}

			// Restore original env vars
			errutil.Setenv("NO_COLOR", origNoColor)
			errutil.Setenv("TERM", origTerm)
		})
	}
}
