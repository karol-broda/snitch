package tui

import (
	"fmt"
	"regexp"
	"snitch/internal/collector"
	"strings"
)

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	if max <= 2 {
		return s[:max]
	}
	return s[:max-1] + SymbolEllipsis
}

var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func stripAnsi(s string) string {
	return ansiRegex.ReplaceAllString(s, "")
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

func sortFieldLabel(f collector.SortField) string {
	switch f {
	case collector.SortByLport:
		return "port"
	case collector.SortByProcess:
		return "proc"
	case collector.SortByPID:
		return "pid"
	case collector.SortByState:
		return "state"
	case collector.SortByProto:
		return "proto"
	default:
		return "port"
	}
}

func formatRemote(addr string, port int) string {
	if addr == "" || addr == "*" || port == 0 {
		return "-"
	}
	return fmt.Sprintf("%s:%d", addr, port)
}

