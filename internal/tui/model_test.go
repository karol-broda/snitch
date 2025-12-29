package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/exp/teatest"
	"github.com/karol-broda/snitch/internal/collector"
)

func TestTUI_InitialState(t *testing.T) {
	m := New(Options{
		Theme:    "dark",
		Interval: time.Second,
	})

	if m.showTCP != true {
		t.Error("expected showTCP to be true by default")
	}
	if m.showUDP != true {
		t.Error("expected showUDP to be true by default")
	}
	if m.showListening != true {
		t.Error("expected showListening to be true by default")
	}
	if m.showEstablished != true {
		t.Error("expected showEstablished to be true by default")
	}
}

func TestTUI_FilterOptions(t *testing.T) {
	m := New(Options{
		Theme:     "dark",
		Interval:  time.Second,
		TCP:       true,
		UDP:       false,
		FilterSet: true,
	})

	if m.showTCP != true {
		t.Error("expected showTCP to be true")
	}
	if m.showUDP != false {
		t.Error("expected showUDP to be false")
	}
}

func TestTUI_MatchesFilters(t *testing.T) {
	m := New(Options{
		Theme:       "dark",
		Interval:    time.Second,
		TCP:         true,
		UDP:         false,
		Listening:   true,
		Established: false,
		FilterSet:   true,
	})

	tests := []struct {
		name     string
		conn     collector.Connection
		expected bool
	}{
		{
			name:     "tcp listen matches",
			conn:     collector.Connection{Proto: "tcp", State: "LISTEN"},
			expected: true,
		},
		{
			name:     "tcp6 listen matches",
			conn:     collector.Connection{Proto: "tcp6", State: "LISTEN"},
			expected: true,
		},
		{
			name:     "udp listen does not match",
			conn:     collector.Connection{Proto: "udp", State: "LISTEN"},
			expected: false,
		},
		{
			name:     "tcp established does not match",
			conn:     collector.Connection{Proto: "tcp", State: "ESTABLISHED"},
			expected: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := m.matchesFilters(tc.conn)
			if result != tc.expected {
				t.Errorf("matchesFilters() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestTUI_MatchesSearch(t *testing.T) {
	m := New(Options{Theme: "dark"})
	m.searchQuery = "firefox"

	tests := []struct {
		name     string
		conn     collector.Connection
		expected bool
	}{
		{
			name:     "process name matches",
			conn:     collector.Connection{Process: "firefox"},
			expected: true,
		},
		{
			name:     "process name case insensitive",
			conn:     collector.Connection{Process: "Firefox"},
			expected: true,
		},
		{
			name:     "no match",
			conn:     collector.Connection{Process: "chrome"},
			expected: false,
		},
		{
			name:     "matches in address",
			conn:     collector.Connection{Raddr: "firefox.com"},
			expected: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := m.matchesSearch(tc.conn)
			if result != tc.expected {
				t.Errorf("matchesSearch() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestTUI_KeyBindings(t *testing.T) {
	tm := teatest.NewTestModel(t, New(Options{Theme: "dark", Interval: time.Hour}))

	// test quit with 'q'
	tm.Send(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})

	tm.WaitFinished(t, teatest.WithFinalTimeout(time.Second*3))
}

func TestTUI_ToggleFilters(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})

	// initial state: all filters on
	if m.showTCP != true || m.showUDP != true {
		t.Fatal("expected all protocol filters on initially")
	}

	// toggle TCP with 't'
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}})
	m = newModel.(model)

	if m.showTCP != false {
		t.Error("expected showTCP to be false after toggle")
	}

	// toggle UDP with 'u'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'u'}})
	m = newModel.(model)

	if m.showUDP != false {
		t.Error("expected showUDP to be false after toggle")
	}

	// toggle listening with 'l'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	m = newModel.(model)

	if m.showListening != false {
		t.Error("expected showListening to be false after toggle")
	}

	// toggle established with 'e'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	m = newModel.(model)

	if m.showEstablished != false {
		t.Error("expected showEstablished to be false after toggle")
	}
}

func TestTUI_HelpToggle(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})

	if m.showHelp != false {
		t.Fatal("expected showHelp to be false initially")
	}

	// toggle help with '?'
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(model)

	if m.showHelp != true {
		t.Error("expected showHelp to be true after toggle")
	}

	// toggle help off
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	m = newModel.(model)

	if m.showHelp != false {
		t.Error("expected showHelp to be false after second toggle")
	}
}

func TestTUI_CursorNavigation(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})

	// add some test data
	m.connections = []collector.Connection{
		{PID: 1, Process: "proc1", Proto: "tcp", State: "LISTEN"},
		{PID: 2, Process: "proc2", Proto: "tcp", State: "LISTEN"},
		{PID: 3, Process: "proc3", Proto: "tcp", State: "LISTEN"},
	}

	if m.cursor != 0 {
		t.Fatal("expected cursor at 0 initially")
	}

	// move down with 'j'
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(model)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after down, got %d", m.cursor)
	}

	// move down again
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(model)

	if m.cursor != 2 {
		t.Errorf("expected cursor at 2 after second down, got %d", m.cursor)
	}

	// move up with 'k'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(model)

	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after up, got %d", m.cursor)
	}

	// go to top with 'g'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	m = newModel.(model)

	if m.cursor != 0 {
		t.Errorf("expected cursor at 0 after 'g', got %d", m.cursor)
	}

	// go to bottom with 'G'
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	m = newModel.(model)

	if m.cursor != 2 {
		t.Errorf("expected cursor at 2 after 'G', got %d", m.cursor)
	}
}

func TestTUI_WindowResize(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})

	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	m = newModel.(model)

	if m.width != 120 {
		t.Errorf("expected width 120, got %d", m.width)
	}
	if m.height != 40 {
		t.Errorf("expected height 40, got %d", m.height)
	}
}

func TestTUI_ViewRenders(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.width = 120
	m.height = 40

	m.connections = []collector.Connection{
		{PID: 1234, Process: "nginx", Proto: "tcp", State: "LISTEN", Laddr: "0.0.0.0", Lport: 80},
	}

	// main view should render without panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	// help view
	m.showHelp = true
	helpView := m.View()
	if helpView == "" {
		t.Error("expected non-empty help view")
	}
}

func TestTUI_ResolutionOptions(t *testing.T) {
	// test default resolution settings
	m := New(Options{Theme: "dark", Interval: time.Hour})

	if m.resolveAddrs != false {
		t.Error("expected resolveAddrs to be false by default (must be explicitly set)")
	}
	if m.resolvePorts != false {
		t.Error("expected resolvePorts to be false by default")
	}

	// test with explicit options
	m2 := New(Options{
		Theme:        "dark",
		Interval:     time.Hour,
		ResolveAddrs: true,
		ResolvePorts: true,
	})

	if m2.resolveAddrs != true {
		t.Error("expected resolveAddrs to be true when set")
	}
	if m2.resolvePorts != true {
		t.Error("expected resolvePorts to be true when set")
	}
}

func TestTUI_ToggleResolution(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour, ResolveAddrs: true})

	if m.resolveAddrs != true {
		t.Fatal("expected resolveAddrs to be true initially")
	}

	// toggle address resolution with 'n'
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(model)

	if m.resolveAddrs != false {
		t.Error("expected resolveAddrs to be false after toggle")
	}

	// toggle back
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = newModel.(model)

	if m.resolveAddrs != true {
		t.Error("expected resolveAddrs to be true after second toggle")
	}

	// toggle port resolution with 'N'
	if m.resolvePorts != false {
		t.Fatal("expected resolvePorts to be false initially")
	}

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	m = newModel.(model)

	if m.resolvePorts != true {
		t.Error("expected resolvePorts to be true after toggle")
	}

	// toggle back
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'N'}})
	m = newModel.(model)

	if m.resolvePorts != false {
		t.Error("expected resolvePorts to be false after second toggle")
	}
}

func TestTUI_ResolveAddrHelper(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.resolveAddrs = false

	// when resolution is off, should return original address
	addr := m.resolveAddr("192.168.1.1")
	if addr != "192.168.1.1" {
		t.Errorf("expected original address when resolution off, got %s", addr)
	}

	// empty and wildcard addresses should pass through unchanged
	if m.resolveAddr("") != "" {
		t.Error("expected empty string to pass through")
	}
	if m.resolveAddr("*") != "*" {
		t.Error("expected wildcard to pass through")
	}
}

func TestTUI_ResolvePortHelper(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.resolvePorts = false

	// when resolution is off, should return port number as string
	port := m.resolvePort(80, "tcp")
	if port != "80" {
		t.Errorf("expected '80' when resolution off, got %s", port)
	}

	port = m.resolvePort(443, "tcp")
	if port != "443" {
		t.Errorf("expected '443' when resolution off, got %s", port)
	}
}

func TestTUI_FormatRemoteHelper(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.resolveAddrs = false
	m.resolvePorts = false

	// empty/wildcard addresses should return dash
	if m.formatRemote("", 80, "tcp") != "-" {
		t.Error("expected dash for empty address")
	}
	if m.formatRemote("*", 80, "tcp") != "-" {
		t.Error("expected dash for wildcard address")
	}
	if m.formatRemote("192.168.1.1", 0, "tcp") != "-" {
		t.Error("expected dash for zero port")
	}

	// valid address:port should format correctly
	result := m.formatRemote("192.168.1.1", 443, "tcp")
	if result != "192.168.1.1:443" {
		t.Errorf("expected '192.168.1.1:443', got %s", result)
	}
}

func TestTUI_MatchesSearchPort(t *testing.T) {
	m := New(Options{Theme: "dark"})

	tests := []struct {
		name        string
		searchQuery string
		conn        collector.Connection
		expected    bool
	}{
		{
			name:        "matches local port",
			searchQuery: "3000",
			conn:        collector.Connection{Lport: 3000},
			expected:    true,
		},
		{
			name:        "matches remote port",
			searchQuery: "443",
			conn:        collector.Connection{Rport: 443},
			expected:    true,
		},
		{
			name:        "matches pid",
			searchQuery: "1234",
			conn:        collector.Connection{PID: 1234},
			expected:    true,
		},
		{
			name:        "partial port match",
			searchQuery: "80",
			conn:        collector.Connection{Lport: 8080},
			expected:    true,
		},
		{
			name:        "no match",
			searchQuery: "9999",
			conn:        collector.Connection{Lport: 80, Rport: 443, PID: 1234},
			expected:    false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m.searchQuery = tc.searchQuery
			result := m.matchesSearch(tc.conn)
			if result != tc.expected {
				t.Errorf("matchesSearch() = %v, want %v", result, tc.expected)
			}
		})
	}
}

func TestTUI_SortCycleIncludesRemote(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})

	// start at default (Lport)
	if m.sortField != collector.SortByLport {
		t.Fatalf("expected initial sort field to be lport, got %v", m.sortField)
	}

	// cycle through all fields and verify raddr and rport are included
	foundRaddr := false
	foundRport := false
	seenFields := make(map[collector.SortField]bool)

	for i := 0; i < 10; i++ {
		m.cycleSort()
		seenFields[m.sortField] = true

		if m.sortField == collector.SortByRaddr {
			foundRaddr = true
		}
		if m.sortField == collector.SortByRport {
			foundRport = true
		}

		if foundRaddr && foundRport {
			break
		}
	}

	if !foundRaddr {
		t.Error("expected sort cycle to include SortByRaddr")
	}
	if !foundRport {
		t.Error("expected sort cycle to include SortByRport")
	}
}

func TestTUI_ExportModal(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.width = 120
	m.height = 40

	// initially export modal should not be shown
	if m.showExportModal {
		t.Fatal("expected showExportModal to be false initially")
	}

	// press 'x' to open export modal
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = newModel.(model)

	if !m.showExportModal {
		t.Error("expected showExportModal to be true after pressing 'x'")
	}

	// type filename
	for _, c := range "test.csv" {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{c}})
		m = newModel.(model)
	}

	if m.exportFilename != "test.csv" {
		t.Errorf("expected exportFilename to be 'test.csv', got '%s'", m.exportFilename)
	}

	// escape should close modal
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	m = newModel.(model)

	if m.showExportModal {
		t.Error("expected showExportModal to be false after escape")
	}
	if m.exportFilename != "" {
		t.Error("expected exportFilename to be cleared after escape")
	}
}

func TestTUI_ExportModalDefaultFilename(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.width = 120
	m.height = 40

	// add test data
	m.connections = []collector.Connection{
		{PID: 1234, Process: "nginx", Proto: "tcp", State: "LISTEN", Laddr: "0.0.0.0", Lport: 80},
	}

	// open export modal
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = newModel.(model)

	// render export modal should show default filename hint
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view with export modal")
	}
}

func TestTUI_ExportModalBackspace(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.width = 120
	m.height = 40

	// open export modal
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = newModel.(model)

	// type filename
	for _, c := range "test.csv" {
		newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{c}})
		m = newModel.(model)
	}

	// backspace should remove last character
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = newModel.(model)

	if m.exportFilename != "test.cs" {
		t.Errorf("expected 'test.cs' after backspace, got '%s'", m.exportFilename)
	}
}

func TestTUI_ExportConnectionsCSV(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})

	m.connections = []collector.Connection{
		{PID: 1234, Process: "nginx", User: "www-data", Proto: "tcp", State: "LISTEN", Laddr: "0.0.0.0", Lport: 80, Raddr: "*", Rport: 0},
		{PID: 5678, Process: "node", User: "node", Proto: "tcp", State: "ESTABLISHED", Laddr: "192.168.1.1", Lport: 3000, Raddr: "10.0.0.1", Rport: 443},
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "test_export.csv")
	m.exportFilename = csvPath

	err := m.exportConnections()
	if err != nil {
		t.Fatalf("exportConnections() failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 lines (header + 2 data), got %d", len(lines))
	}

	if !strings.Contains(lines[0], "PID") || !strings.Contains(lines[0], "PROCESS") {
		t.Error("header line should contain PID and PROCESS")
	}

	if !strings.Contains(lines[1], "nginx") || !strings.Contains(lines[1], "1234") {
		t.Error("first data line should contain nginx and 1234")
	}

	if !strings.Contains(lines[2], "node") || !strings.Contains(lines[2], "5678") {
		t.Error("second data line should contain node and 5678")
	}
}

func TestTUI_ExportConnectionsTSV(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})

	m.connections = []collector.Connection{
		{PID: 1234, Process: "nginx", User: "www-data", Proto: "tcp", State: "LISTEN", Laddr: "0.0.0.0", Lport: 80, Raddr: "*", Rport: 0},
	}

	tmpDir := t.TempDir()
	tsvPath := filepath.Join(tmpDir, "test_export.tsv")
	m.exportFilename = tsvPath

	err := m.exportConnections()
	if err != nil {
		t.Fatalf("exportConnections() failed: %v", err)
	}

	content, err := os.ReadFile(tsvPath)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	// TSV should use tabs
	if !strings.Contains(lines[0], "\t") {
		t.Error("TSV file should use tabs as delimiters")
	}

	// CSV delimiter should not be present between fields
	fields := strings.Split(lines[1], "\t")
	if len(fields) < 9 {
		t.Errorf("expected at least 9 tab-separated fields, got %d", len(fields))
	}
}

func TestTUI_ExportWithFilters(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.showTCP = true
	m.showUDP = false

	m.connections = []collector.Connection{
		{PID: 1, Process: "tcp_proc", Proto: "tcp", State: "LISTEN", Laddr: "0.0.0.0", Lport: 80},
		{PID: 2, Process: "udp_proc", Proto: "udp", State: "LISTEN", Laddr: "0.0.0.0", Lport: 53},
	}

	tmpDir := t.TempDir()
	csvPath := filepath.Join(tmpDir, "filtered_export.csv")
	m.exportFilename = csvPath

	err := m.exportConnections()
	if err != nil {
		t.Fatalf("exportConnections() failed: %v", err)
	}

	content, err := os.ReadFile(csvPath)
	if err != nil {
		t.Fatalf("failed to read exported file: %v", err)
	}

	lines := strings.Split(strings.TrimSpace(string(content)), "\n")

	// should only have header + 1 TCP connection (UDP filtered out)
	if len(lines) != 2 {
		t.Errorf("expected 2 lines (header + 1 TCP), got %d", len(lines))
	}

	if strings.Contains(string(content), "udp_proc") {
		t.Error("UDP connection should not be exported when UDP filter is off")
	}
}

func TestTUI_ExportFormatToggle(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.width = 120
	m.height = 40

	// open export modal
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = newModel.(model)

	// default format should be csv
	if m.exportFormat != "csv" {
		t.Errorf("expected default format 'csv', got '%s'", m.exportFormat)
	}

	// tab should toggle to tsv
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(model)

	if m.exportFormat != "tsv" {
		t.Errorf("expected format 'tsv' after tab, got '%s'", m.exportFormat)
	}

	// tab again should toggle back to csv
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(model)

	if m.exportFormat != "csv" {
		t.Errorf("expected format 'csv' after second tab, got '%s'", m.exportFormat)
	}
}

func TestTUI_ExportModalRenderWithStats(t *testing.T) {
	m := New(Options{Theme: "dark", Interval: time.Hour})
	m.width = 120
	m.height = 40

	m.connections = []collector.Connection{
		{PID: 1, Process: "nginx", Proto: "tcp", State: "LISTEN", Laddr: "0.0.0.0", Lport: 80},
		{PID: 2, Process: "postgres", Proto: "tcp", State: "LISTEN", Laddr: "127.0.0.1", Lport: 5432},
		{PID: 3, Process: "node", Proto: "tcp", State: "ESTABLISHED", Laddr: "192.168.1.1", Lport: 3000},
	}

	m.showExportModal = true
	m.exportFormat = "csv"

	view := m.View()

	// modal should contain summary info
	if !strings.Contains(view, "3") {
		t.Error("modal should show connection count")
	}

	// modal should show format options
	if !strings.Contains(view, "CSV") || !strings.Contains(view, "TSV") {
		t.Error("modal should show format options")
	}
}

