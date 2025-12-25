package state

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"github.com/karol-broda/snitch/internal/collector"
)

// TUIState holds view options that can be persisted between sessions
type TUIState struct {
	ShowTCP         bool                `json:"show_tcp"`
	ShowUDP         bool                `json:"show_udp"`
	ShowListening   bool                `json:"show_listening"`
	ShowEstablished bool                `json:"show_established"`
	ShowOther       bool                `json:"show_other"`
	SortField       collector.SortField `json:"sort_field"`
	SortReverse     bool                `json:"sort_reverse"`
	ResolveAddrs    bool                `json:"resolve_addrs"`
	ResolvePorts    bool                `json:"resolve_ports"`
}

var (
	saveMu   sync.Mutex
	saveChan chan TUIState
	once     sync.Once
)

// Path returns the XDG-compliant state file path
func Path() string {
	stateDir := os.Getenv("XDG_STATE_HOME")
	if stateDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return ""
		}
		stateDir = filepath.Join(home, ".local", "state")
	}
	return filepath.Join(stateDir, "snitch", "tui.json")
}

// Load reads the TUI state from disk.
// returns nil if state file doesn't exist or can't be read.
func Load() *TUIState {
	path := Path()
	if path == "" {
		return nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}

	var state TUIState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil
	}

	return &state
}

// Save writes the TUI state to disk synchronously.
// creates parent directories if needed.
func Save(state TUIState) error {
	path := Path()
	if path == "" {
		return nil
	}

	saveMu.Lock()
	defer saveMu.Unlock()

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// SaveAsync queues a state save to happen in the background.
// only the most recent state is saved if multiple saves are queued.
func SaveAsync(state TUIState) {
	once.Do(func() {
		saveChan = make(chan TUIState, 1)
		go saveWorker()
	})

	// non-blocking send, replace pending save with newer state
	select {
	case saveChan <- state:
	default:
		// channel full, drain and replace
		select {
		case <-saveChan:
		default:
		}
		select {
		case saveChan <- state:
		default:
		}
	}
}

func saveWorker() {
	for state := range saveChan {
		_ = Save(state)
	}
}

// Default returns a TUIState with default values
func Default() TUIState {
	return TUIState{
		ShowTCP:         true,
		ShowUDP:         true,
		ShowListening:   true,
		ShowEstablished: true,
		ShowOther:       true,
		SortField:       collector.SortByLport,
		SortReverse:     false,
		ResolveAddrs:    false,
		ResolvePorts:    false,
	}
}

