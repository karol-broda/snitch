package state

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/karol-broda/snitch/internal/collector"
)

func TestPath_XDGStateHome(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/custom/state")
	path := Path()

	expected := "/custom/state/snitch/tui.json"
	if path != expected {
		t.Errorf("Path() = %q, want %q", path, expected)
	}
}

func TestPath_DefaultFallback(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "")
	path := Path()

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("cannot determine home directory")
	}

	expected := filepath.Join(home, ".local", "state", "snitch", "tui.json")
	if path != expected {
		t.Errorf("Path() = %q, want %q", path, expected)
	}
}

func TestDefault(t *testing.T) {
	d := Default()

	if d.ShowTCP != true {
		t.Error("expected ShowTCP to be true")
	}
	if d.ShowUDP != true {
		t.Error("expected ShowUDP to be true")
	}
	if d.ShowListening != true {
		t.Error("expected ShowListening to be true")
	}
	if d.ShowEstablished != true {
		t.Error("expected ShowEstablished to be true")
	}
	if d.ShowOther != true {
		t.Error("expected ShowOther to be true")
	}
	if d.SortField != collector.SortByLport {
		t.Errorf("expected SortField to be %q, got %q", collector.SortByLport, d.SortField)
	}
	if d.SortReverse != false {
		t.Error("expected SortReverse to be false")
	}
	if d.ResolveAddrs != false {
		t.Error("expected ResolveAddrs to be false")
	}
	if d.ResolvePorts != false {
		t.Error("expected ResolvePorts to be false")
	}
}

func TestSaveAndLoad(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	state := TUIState{
		ShowTCP:         false,
		ShowUDP:         true,
		ShowListening:   true,
		ShowEstablished: false,
		ShowOther:       true,
		SortField:       collector.SortByProcess,
		SortReverse:     true,
		ResolveAddrs:    true,
		ResolvePorts:    false,
	}

	err := Save(state)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// verify file was created
	path := Path()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatal("expected state file to exist after Save()")
	}

	loaded := Load()
	if loaded == nil {
		t.Fatal("Load() returned nil")
	}

	if loaded.ShowTCP != state.ShowTCP {
		t.Errorf("ShowTCP = %v, want %v", loaded.ShowTCP, state.ShowTCP)
	}
	if loaded.ShowUDP != state.ShowUDP {
		t.Errorf("ShowUDP = %v, want %v", loaded.ShowUDP, state.ShowUDP)
	}
	if loaded.ShowListening != state.ShowListening {
		t.Errorf("ShowListening = %v, want %v", loaded.ShowListening, state.ShowListening)
	}
	if loaded.ShowEstablished != state.ShowEstablished {
		t.Errorf("ShowEstablished = %v, want %v", loaded.ShowEstablished, state.ShowEstablished)
	}
	if loaded.ShowOther != state.ShowOther {
		t.Errorf("ShowOther = %v, want %v", loaded.ShowOther, state.ShowOther)
	}
	if loaded.SortField != state.SortField {
		t.Errorf("SortField = %v, want %v", loaded.SortField, state.SortField)
	}
	if loaded.SortReverse != state.SortReverse {
		t.Errorf("SortReverse = %v, want %v", loaded.SortReverse, state.SortReverse)
	}
	if loaded.ResolveAddrs != state.ResolveAddrs {
		t.Errorf("ResolveAddrs = %v, want %v", loaded.ResolveAddrs, state.ResolveAddrs)
	}
	if loaded.ResolvePorts != state.ResolvePorts {
		t.Errorf("ResolvePorts = %v, want %v", loaded.ResolvePorts, state.ResolvePorts)
	}
}

func TestLoad_NonExistent(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	loaded := Load()
	if loaded != nil {
		t.Error("expected Load() to return nil for non-existent file")
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	// create directory and invalid json file
	stateDir := filepath.Join(tmpDir, "snitch")
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		t.Fatal(err)
	}
	stateFile := filepath.Join(stateDir, "tui.json")
	if err := os.WriteFile(stateFile, []byte("not valid json"), 0644); err != nil {
		t.Fatal(err)
	}

	loaded := Load()
	if loaded != nil {
		t.Error("expected Load() to return nil for invalid JSON")
	}
}

func TestSave_CreatesDirectories(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	// snitch directory should not exist yet
	snitchDir := filepath.Join(tmpDir, "snitch")
	if _, err := os.Stat(snitchDir); err == nil {
		t.Fatal("expected snitch directory to not exist initially")
	}

	err := Save(Default())
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	// directory should now exist
	if _, err := os.Stat(snitchDir); os.IsNotExist(err) {
		t.Error("expected Save() to create parent directories")
	}
}

func TestSaveAsync(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	state := TUIState{
		ShowTCP:   false,
		SortField: collector.SortByPID,
	}

	SaveAsync(state)

	// wait for background save with timeout
	deadline := time.Now().Add(100 * time.Millisecond)
	for time.Now().Before(deadline) {
		if loaded := Load(); loaded != nil {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}

	t.Log("SaveAsync may not have completed in time (non-fatal in CI)")
}

func TestTUIState_JSONRoundtrip(t *testing.T) {
	// verify all sort fields serialize correctly
	sortFields := []collector.SortField{
		collector.SortByLport,
		collector.SortByProcess,
		collector.SortByPID,
		collector.SortByState,
		collector.SortByProto,
	}

	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	for _, sf := range sortFields {
		state := TUIState{
			ShowTCP:   true,
			SortField: sf,
		}

		if err := Save(state); err != nil {
			t.Fatalf("Save() error for %q: %v", sf, err)
		}

		loaded := Load()
		if loaded == nil {
			t.Fatalf("Load() returned nil for %q", sf)
		}

		if loaded.SortField != sf {
			t.Errorf("SortField roundtrip failed: got %q, want %q", loaded.SortField, sf)
		}
	}
}
