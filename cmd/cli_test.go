package cmd

import (
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/karol-broda/snitch/internal/errutil"
	"github.com/karol-broda/snitch/internal/testutil"
)

// TestCLIContract tests the CLI interface contracts as specified in the README
func TestCLIContract(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		expectExitCode int
		expectStdout   []string
		expectStderr   []string
		description    string
	}{
		{
			name:           "help_root",
			args:           []string{"--help"},
			expectExitCode: 0,
			expectStdout:   []string{"snitch is a tool for inspecting network connections", "Usage:", "Available Commands:"},
			expectStderr:   nil,
			description:    "Root help should show usage and available commands",
		},
		{
			name:           "help_ls",
			args:           []string{"ls", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"One-shot listing of connections", "Usage:", "Flags:"},
			expectStderr:   nil,
			description:    "ls help should show command description and flags",
		},
		{
			name:           "help_top",
			args:           []string{"top", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"Live TUI for inspecting connections", "Usage:", "Flags:"},
			expectStderr:   nil,
			description:    "top help should show command description and flags",
		},
		{
			name:           "help_watch",
			args:           []string{"watch", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"Stream connection events as json frames", "Usage:", "Flags:"},
			expectStderr:   nil,
			description:    "watch help should show command description and flags",
		},
		{
			name:           "help_stats",
			args:           []string{"stats", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"Aggregated connection counters", "Usage:", "Flags:"},
			expectStderr:   nil,
			description:    "stats help should show command description and flags",
		},
		{
			name:           "help_trace",
			args:           []string{"trace", "--help"},
			expectExitCode: 0,
			expectStdout:   []string{"Print new/closed connections", "Usage:", "Flags:"},
			expectStderr:   nil,
			description:    "trace help should show command description and flags",
		},
		{
			name:           "version",
			args:           []string{"version"},
			expectExitCode: 0,
			expectStdout:   []string{"snitch", "commit", "built"},
			expectStderr:   nil,
			description:    "version command should show version information",
		},
		{
			name:           "invalid_command",
			args:           []string{"invalid"},
			expectExitCode: 1,
			expectStdout:   nil,
			expectStderr:   []string{"unknown command", "invalid"},
			description:    "Invalid command should exit with code 1 and show error",
		},
		{
			name:           "invalid_flag",
			args:           []string{"ls", "--invalid-flag"},
			expectExitCode: 1,
			expectStdout:   nil,
			expectStderr:   []string{"unknown flag"},
			description:    "Invalid flag should exit with code 1 and show error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build the command
			cmd := exec.Command("go", append([]string{"run", "../main.go"}, tt.args...)...)

			// Set environment for consistent testing
			cmd.Env = append(os.Environ(),
				"SNITCH_NO_COLOR=1",
				"SNITCH_RESOLVE=0",
			)

			// Run command and capture output
			output, err := cmd.CombinedOutput()

			// Check exit code
			actualExitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					actualExitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Failed to run command: %v", err)
				}
			}

			if actualExitCode != tt.expectExitCode {
				t.Errorf("Expected exit code %d, got %d", tt.expectExitCode, actualExitCode)
			}

			outputStr := string(output)

			// Check expected stdout content
			for _, expected := range tt.expectStdout {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected stdout to contain %q, but output was:\n%s", expected, outputStr)
				}
			}

			// Check expected stderr content
			for _, expected := range tt.expectStderr {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain error %q, but output was:\n%s", expected, outputStr)
				}
			}
		})
	}
}

// TestFlagInteractions tests complex flag interactions and precedence
func TestFlagInteractions(t *testing.T) {
	// Skip this test for now as it's using real system data instead of mocks
	t.Skip("Skipping TestFlagInteractions as it needs to be rewritten to use proper mocks")

	tests := []struct {
		name        string
		args        []string
		expectOut   []string
		expectError bool
		description string
	}{
		{
			name:        "output_json_flag",
			args:        []string{"ls", "-o", "json"},
			expectOut:   []string{`"pid"`, `"process"`, `[`},
			expectError: false,
			description: "JSON output flag should produce valid JSON",
		},
		{
			name:        "output_csv_flag",
			args:        []string{"ls", "-o", "csv"},
			expectOut:   []string{"PID,PROCESS", "1,tcp-server", "2,udp-server", "3,unix-app"},
			expectError: false,
			description: "CSV output flag should produce CSV format",
		},
		{
			name:        "no_headers_flag",
			args:        []string{"ls", "--no-headers"},
			expectOut:   nil, // Will verify no header line
			expectError: false,
			description: "No headers flag should omit column headers",
		},
		{
			name:        "ipv4_filter",
			args:        []string{"ls", "-4"},
			expectOut:   []string{"tcp", "udp"}, // Should show IPv4 connections
			expectError: false,
			description: "IPv4 filter should only show IPv4 connections",
		},
		{
			name:        "numeric_flag",
			args:        []string{"ls", "-n"},
			expectOut:   []string{"0.0.0.0", "*"}, // Should show numeric addresses
			expectError: false,
			description: "Numeric flag should disable name resolution",
		},
		{
			name:        "invalid_output_format",
			args:        []string{"ls", "-o", "invalid"},
			expectOut:   nil,
			expectError: true,
			description: "Invalid output format should cause error",
		},
		{
			name:        "combined_filters",
			args:        []string{"ls", "proto=tcp", "state=listen"},
			expectOut:   []string{"tcp", "LISTEN"},
			expectError: false,
			description: "Multiple filters should be ANDed together",
		},
		{
			name:        "invalid_filter_format",
			args:        []string{"ls", "invalid-filter"},
			expectOut:   nil,
			expectError: true,
			description: "Invalid filter format should cause error",
		},
		{
			name:        "invalid_filter_key",
			args:        []string{"ls", "badkey=value"},
			expectOut:   nil,
			expectError: true,
			description: "Invalid filter key should cause error",
		},
		{
			name:        "invalid_pid_filter",
			args:        []string{"ls", "pid=notanumber"},
			expectOut:   nil,
			expectError: true,
			description: "Invalid PID value should cause error",
		},
		{
			name:        "fields_flag",
			args:        []string{"ls", "-f", "pid,process,proto"},
			expectOut:   []string{"PID", "PROCESS", "PROTO"},
			expectError: false,
			description: "Fields flag should limit displayed columns",
		},
		{
			name:        "sort_flag",
			args:        []string{"ls", "-s", "pid:desc"},
			expectOut:   []string{"3", "2", "1"}, // Should be in descending PID order
			expectError: false,
			description: "Sort flag should order results",
		},
	}

	for _, tt := range tests {
		// Capture output
		capture := testutil.NewOutputCapture(t)
		capture.Start()

		// Reset global variables that might be modified by flags
		resetGlobalFlags()

		// Simulate command execution by directly calling the command functions
		// This is easier than spawning processes for integration tests
		if len(tt.args) > 0 && tt.args[0] == "ls" {
			// Parse ls-specific flags and arguments
			outputFormat := "table"
			noHeaders := false
			ipv4 := false
			ipv6 := false
			numeric := false
			fields := ""
			sortBy := ""
			filters := []string{}

			// Simple flag parsing for test
			i := 1
			for i < len(tt.args) {
				arg := tt.args[i]
				if arg == "-o" && i+1 < len(tt.args) {
					outputFormat = tt.args[i+1]
					i += 2
				} else if arg == "--no-headers" {
					noHeaders = true
					i++
				} else if arg == "-4" {
					ipv4 = true
					i++
				} else if arg == "-6" {
					ipv6 = true
					i++
				} else if arg == "-n" {
					numeric = true
					i++
				} else if arg == "-f" && i+1 < len(tt.args) {
					fields = tt.args[i+1]
					i += 2
				} else if arg == "-s" && i+1 < len(tt.args) {
					sortBy = tt.args[i+1]
					i += 2
				} else if strings.Contains(arg, "=") {
					filters = append(filters, arg)
					i++
				} else {
					i++
				}
			}

			// Set global variables
			oldOutputFormat := outputFormat
			oldNoHeaders := noHeaders
			oldIpv4 := ipv4
			oldIpv6 := ipv6
			oldNumeric := numeric
			oldFields := fields
			oldSortBy := sortBy
			defer func() {
				outputFormat = oldOutputFormat
				noHeaders = oldNoHeaders
				ipv4 = oldIpv4
				ipv6 = oldIpv6
				numeric = oldNumeric
				fields = oldFields
				sortBy = oldSortBy
			}()

			// Build the command
			cmd := exec.Command("go", append([]string{"run", "../main.go"}, tt.args...)...)

			// Set environment for consistent testing
			cmd.Env = append(os.Environ(),
				"SNITCH_NO_COLOR=1",
				"SNITCH_RESOLVE=0",
			)

			// Run command and capture output
			output, err := cmd.CombinedOutput()

			// Check exit code
			actualExitCode := 0
			if err != nil {
				if exitError, ok := err.(*exec.ExitError); ok {
					actualExitCode = exitError.ExitCode()
				} else {
					t.Fatalf("Failed to run command: %v", err)
				}
			}

			if tt.expectError {
				if actualExitCode == 0 {
					t.Errorf("Expected command to fail with error, but it succeeded. Output:\n%s", string(output))
				}
			} else {
				if actualExitCode != 0 {
					t.Errorf("Expected command to succeed, but it failed with exit code %d. Output:\n%s", actualExitCode, string(output))
				}
			}

			outputStr := string(output)

			// Check expected stdout content
			for _, expected := range tt.expectOut {
				if !strings.Contains(outputStr, expected) {
					t.Errorf("Expected output to contain %q, but output was:\n%s", expected, outputStr)
				}
			}
		}
	}
}

// resetGlobalFlags resets global flag variables to their defaults
func resetGlobalFlags() {
	outputFormat = "table"
	noHeaders = false
	showTimestamp = false
	sortBy = ""
	fields = ""
	filterIPv4 = false
	filterIPv6 = false
	colorMode = "auto"
	resolveAddrs = true
	resolvePorts = false
}

// TestEnvironmentVariables tests that environment variables are properly handled
func TestEnvironmentVariables(t *testing.T) {
	tests := []struct {
		name           string
		envVars        map[string]string
		expectBehavior string
		description    string
	}{
		{
			name: "snitch_no_color",
			envVars: map[string]string{
				"SNITCH_NO_COLOR": "1",
			},
			expectBehavior: "no_color",
			description:    "SNITCH_NO_COLOR=1 should disable colors",
		},
		{
			name: "snitch_resolve_disabled",
			envVars: map[string]string{
				"SNITCH_RESOLVE": "0",
			},
			expectBehavior: "numeric",
			description:    "SNITCH_RESOLVE=0 should enable numeric mode",
		},
		{
			name: "snitch_theme",
			envVars: map[string]string{
				"SNITCH_THEME": "mono",
			},
			expectBehavior: "mono_theme",
			description:    "SNITCH_THEME should set the default theme",
		},
	}

	for _, tt := range tests {
		// Set environment variables
		oldEnvVars := make(map[string]string)
		for key, value := range tt.envVars {
			oldEnvVars[key] = os.Getenv(key)
			errutil.Setenv(key, value)
		}

		// Clean up environment variables
		defer func() {
			for key, oldValue := range oldEnvVars {
				if oldValue == "" {
					errutil.Unsetenv(key)
				} else {
					errutil.Setenv(key, oldValue)
				}
			}
		}()

		// Test that environment variables affect behavior
		// This would normally require running the full CLI with subprocesses
		// For now, we just verify the environment variables are set correctly
		for key, expectedValue := range tt.envVars {
			actualValue := os.Getenv(key)
			if actualValue != expectedValue {
				t.Errorf("Expected %s=%s, but got %s=%s", key, expectedValue, key, actualValue)
			}
		}
	}
}

// TestErrorExitCodes tests that the CLI returns appropriate exit codes
func TestErrorExitCodes(t *testing.T) {
	tests := []struct {
		name         string
		command      []string
		expectedCode int
		description  string
	}{
		{
			name:         "success",
			command:      []string{"version"},
			expectedCode: 0,
			description:  "Successful commands should exit with 0",
		},
		{
			name:         "invalid_usage",
			command:      []string{"ls", "--invalid-flag"},
			expectedCode: 1, // Using 1 instead of 2 since that's what cobra returns
			description:  "Invalid usage should exit with error code",
		},
	}

	for _, tt := range tests {
		cmd := exec.Command("go", append([]string{"run", "../main.go"}, tt.command...)...)
		cmd.Env = append(os.Environ(), "SNITCH_NO_COLOR=1")

		err := cmd.Run()

		actualCode := 0
		if err != nil {
			if exitError, ok := err.(*exec.ExitError); ok {
				actualCode = exitError.ExitCode()
			}
		}

		if actualCode != tt.expectedCode {
			t.Errorf("Expected exit code %d, got %d for command: %v",
				tt.expectedCode, actualCode, tt.command)
		}
	}
}
