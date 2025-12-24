package testutil

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/karol-broda/snitch/internal/collector"
	"github.com/karol-broda/snitch/internal/errutil"
)

// TestCollector wraps MockCollector for use in tests
type TestCollector struct {
	*collector.MockCollector
}

// NewTestCollector creates a new test collector with default data
func NewTestCollector() *TestCollector {
	return &TestCollector{
		MockCollector: collector.NewMockCollector(),
	}
}

// NewTestCollectorWithFixture creates a test collector with a specific fixture
func NewTestCollectorWithFixture(fixtureName string) *TestCollector {
	fixtures := collector.GetTestFixtures()
	for _, fixture := range fixtures {
		if fixture.Name == fixtureName {
			mock := collector.NewMockCollector()
			mock.SetConnections(fixture.Connections)
			return &TestCollector{MockCollector: mock}
		}
	}
	
	// Fallback to default if fixture not found
	return NewTestCollector()
}

// SetupTestEnvironment sets up a clean test environment
func SetupTestEnvironment(t *testing.T) (string, func()) {
	// Create temporary directory for test files
	tempDir, err := os.MkdirTemp("", "snitch-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	
	// Set test environment variables
	oldConfig := os.Getenv("SNITCH_CONFIG")
	oldNoColor := os.Getenv("SNITCH_NO_COLOR")
	
	errutil.Setenv("SNITCH_NO_COLOR", "1")
	
	// Cleanup function
	cleanup := func() {
		errutil.RemoveAll(tempDir)
		errutil.Setenv("SNITCH_CONFIG", oldConfig)
		errutil.Setenv("SNITCH_NO_COLOR", oldNoColor)
	}
	
	return tempDir, cleanup
}

// CreateFixtureFile creates a JSON fixture file with the given connections
func CreateFixtureFile(t *testing.T, dir string, name string, connections []collector.Connection) string {
	mock := collector.NewMockCollector()
	mock.SetConnections(connections)
	
	filePath := filepath.Join(dir, name+".json")
	if err := mock.SaveToFile(filePath); err != nil {
		t.Fatalf("Failed to create fixture file %s: %v", filePath, err)
	}
	
	return filePath
}

// LoadFixtureFile loads connections from a JSON fixture file
func LoadFixtureFile(t *testing.T, filePath string) []collector.Connection {
	mock, err := collector.NewMockCollectorFromFile(filePath)
	if err != nil {
		t.Fatalf("Failed to load fixture file %s: %v", filePath, err)
	}
	
	connections, err := mock.GetConnections()
	if err != nil {
		t.Fatalf("Failed to get connections from fixture: %v", err)
	}
	
	return connections
}

// AssertConnectionsEqual compares two slices of connections for equality
func AssertConnectionsEqual(t *testing.T, expected, actual []collector.Connection) {
	if len(expected) != len(actual) {
		t.Errorf("Connection count mismatch: expected %d, got %d", len(expected), len(actual))
		return
	}
	
	for i, exp := range expected {
		act := actual[i]
		
		// Compare key fields (timestamps may vary slightly)
		if exp.PID != act.PID {
			t.Errorf("Connection %d PID mismatch: expected %d, got %d", i, exp.PID, act.PID)
		}
		if exp.Process != act.Process {
			t.Errorf("Connection %d Process mismatch: expected %s, got %s", i, exp.Process, act.Process)
		}
		if exp.Proto != act.Proto {
			t.Errorf("Connection %d Proto mismatch: expected %s, got %s", i, exp.Proto, act.Proto)
		}
		if exp.State != act.State {
			t.Errorf("Connection %d State mismatch: expected %s, got %s", i, exp.State, act.State)
		}
		if exp.Laddr != act.Laddr {
			t.Errorf("Connection %d Laddr mismatch: expected %s, got %s", i, exp.Laddr, act.Laddr)
		}
		if exp.Lport != act.Lport {
			t.Errorf("Connection %d Lport mismatch: expected %d, got %d", i, exp.Lport, act.Lport)
		}
	}
}

// GetTestConfig returns a test configuration with safe defaults
func GetTestConfig() map[string]interface{} {
	return map[string]interface{}{
		"defaults": map[string]interface{}{
			"interval":      "1s",
			"numeric":       true, // Disable resolution in tests
			"fields":        []string{"pid", "process", "proto", "state", "laddr", "lport"},
			"theme":         "mono", // Use monochrome theme in tests
			"units":         "auto",
			"color":         "never",
			"resolve":       false,
			"ipv4":          false,
			"ipv6":          false,
			"no_headers":    false,
			"output_format": "table",
			"sort_by":       "",
		},
	}
}

// CaptureOutput captures stdout/stderr during test execution
type OutputCapture struct {
	stdout *os.File
	stderr *os.File
	oldStdout *os.File
	oldStderr *os.File
	stdoutFile string
	stderrFile string
}

// NewOutputCapture creates a new output capture
func NewOutputCapture(t *testing.T) *OutputCapture {
	tempDir, err := os.MkdirTemp("", "snitch-output-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir for output capture: %v", err)
	}
	
	stdoutFile := filepath.Join(tempDir, "stdout")
	stderrFile := filepath.Join(tempDir, "stderr")
	
	stdout, err := os.Create(stdoutFile)
	if err != nil {
		t.Fatalf("Failed to create stdout file: %v", err)
	}
	
	stderr, err := os.Create(stderrFile)
	if err != nil {
		t.Fatalf("Failed to create stderr file: %v", err)
	}
	
	return &OutputCapture{
		stdout:     stdout,
		stderr:     stderr,
		oldStdout:  os.Stdout,
		oldStderr:  os.Stderr,
		stdoutFile: stdoutFile,
		stderrFile: stderrFile,
	}
}

// Start begins capturing output
func (oc *OutputCapture) Start() {
	os.Stdout = oc.stdout
	os.Stderr = oc.stderr
}

// Stop stops capturing and returns the captured output
func (oc *OutputCapture) Stop() (string, string, error) {
	// Restore original stdout/stderr
	os.Stdout = oc.oldStdout
	os.Stderr = oc.oldStderr
	
	// Close files
	errutil.Close(oc.stdout)
	errutil.Close(oc.stderr)
	
	// Read captured content
	stdoutContent, err := os.ReadFile(oc.stdoutFile)
	if err != nil {
		return "", "", err
	}
	
	stderrContent, err := os.ReadFile(oc.stderrFile)
	if err != nil {
		return "", "", err
	}
	
	// Cleanup
	errutil.Remove(oc.stdoutFile)
	errutil.Remove(oc.stderrFile)
	errutil.Remove(filepath.Dir(oc.stdoutFile))
	
	return string(stdoutContent), string(stderrContent), nil
}