package tui

import (
	"fmt"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/karol-broda/snitch/internal/collector"
	"github.com/karol-broda/snitch/internal/state"
	"github.com/karol-broda/snitch/internal/theme"
)

type model struct {
	connections []collector.Connection
	cursor      int
	width       int
	height      int

	// filtering
	showTCP         bool
	showUDP         bool
	showListening   bool
	showEstablished bool
	showOther       bool
	searchQuery     string
	searchActive    bool

	// sorting
	sortField   collector.SortField
	sortReverse bool

	// display options
	resolveAddrs bool // when true, resolve IP addresses to hostnames
	resolvePorts bool // when true, resolve port numbers to service names

	// ui state
	theme       *theme.Theme
	showHelp    bool
	showDetail  bool
	selected    *collector.Connection
	interval    time.Duration
	lastRefresh time.Time
	err         error

	// watched processes
	watchedPIDs map[int]bool

	// kill confirmation
	showKillConfirm bool
	killTarget      *collector.Connection

	// status message (temporary feedback)
	statusMessage string
	statusExpiry  time.Time

	// state persistence
	rememberState bool
}

type Options struct {
	Theme         string
	Interval      time.Duration
	TCP           bool
	UDP           bool
	Listening     bool
	Established   bool
	Other         bool
	FilterSet     bool // true if user specified any filter flags
	ResolveAddrs  bool // when true, resolve IP addresses to hostnames
	ResolvePorts  bool // when true, resolve port numbers to service names
	NoCache       bool // when true, disable DNS caching
	RememberState bool // when true, persist view options between sessions
}

func New(opts Options) model {
	interval := opts.Interval
	if interval == 0 {
		interval = time.Second
	}

	// default: show everything
	showTCP := true
	showUDP := true
	showListening := true
	showEstablished := true
	showOther := true
	sortField := collector.SortByLport
	sortReverse := false
	resolveAddrs := opts.ResolveAddrs
	resolvePorts := opts.ResolvePorts

	// load saved state if enabled and no CLI filter flags were specified
	if opts.RememberState && !opts.FilterSet {
		if saved := state.Load(); saved != nil {
			showTCP = saved.ShowTCP
			showUDP = saved.ShowUDP
			showListening = saved.ShowListening
			showEstablished = saved.ShowEstablished
			showOther = saved.ShowOther
			sortField = saved.SortField
			sortReverse = saved.SortReverse
			resolveAddrs = saved.ResolveAddrs
			resolvePorts = saved.ResolvePorts
		}
	}

	// if user specified filters, use those instead (CLI flags take precedence)
	if opts.FilterSet {
		showTCP = opts.TCP
		showUDP = opts.UDP
		showListening = opts.Listening
		showEstablished = opts.Established
		showOther = opts.Other

		// if only proto filters set, show all states
		if !opts.Listening && !opts.Established && !opts.Other {
			showListening = true
			showEstablished = true
			showOther = true
		}
		// if only state filters set, show all protos
		if !opts.TCP && !opts.UDP {
			showTCP = true
			showUDP = true
		}
	}

	return model{
		connections:     []collector.Connection{},
		showTCP:         showTCP,
		showUDP:         showUDP,
		showListening:   showListening,
		showEstablished: showEstablished,
		showOther:       showOther,
		sortField:       sortField,
		sortReverse:     sortReverse,
		resolveAddrs:    resolveAddrs,
		resolvePorts:    resolvePorts,
		theme:           theme.GetTheme(opts.Theme),
		interval:        interval,
		lastRefresh:     time.Now(),
		watchedPIDs:     make(map[int]bool),
		rememberState:   opts.RememberState,
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		tea.HideCursor,
		m.fetchData(),
		m.tick(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case tickMsg:
		return m, tea.Batch(m.fetchData(), m.tick())

	case dataMsg:
		m.connections = msg.connections
		m.lastRefresh = time.Now()
		m.applySorting()
		m.clampCursor()
		return m, nil

	case errMsg:
		m.err = msg.err
		return m, nil

	case killResultMsg:
		if msg.success {
			m.statusMessage = fmt.Sprintf("killed %s (pid %d)", msg.process, msg.pid)
		} else {
			m.statusMessage = fmt.Sprintf("failed to kill pid %d: %v", msg.pid, msg.err)
		}
		m.statusExpiry = time.Now().Add(3 * time.Second)
		return m, tea.Batch(m.fetchData(), clearStatusAfter(3*time.Second))

	case clearStatusMsg:
		if time.Now().After(m.statusExpiry) {
			m.statusMessage = ""
		}
		return m, nil
	}

	return m, nil
}

func (m model) View() string {
	if m.err != nil {
		return m.renderError()
	}
	if m.showHelp {
		return m.renderHelp()
	}
	if m.showDetail && m.selected != nil {
		return m.renderDetail()
	}

	main := m.renderMain()

	// overlay kill confirmation modal on top of main view
	if m.showKillConfirm && m.killTarget != nil {
		return m.overlayModal(main, m.renderKillModal())
	}

	return main
}

func (m *model) applySorting() {
	direction := collector.SortAsc
	if m.sortReverse {
		direction = collector.SortDesc
	}
	collector.SortConnections(m.connections, collector.SortOptions{
		Field:     m.sortField,
		Direction: direction,
	})
}

func (m *model) clampCursor() {
	visible := m.visibleConnections()
	if m.cursor >= len(visible) {
		m.cursor = len(visible) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m model) visibleConnections() []collector.Connection {
	var watched []collector.Connection
	var unwatched []collector.Connection

	for _, c := range m.connections {
		if !m.matchesFilters(c) {
			continue
		}
		if m.searchQuery != "" && !m.matchesSearch(c) {
			continue
		}
		if m.isWatched(c.PID) {
			watched = append(watched, c)
		} else {
			unwatched = append(unwatched, c)
		}
	}

	// watched connections appear first
	return append(watched, unwatched...)
}

func (m model) matchesFilters(c collector.Connection) bool {
	isTCP := c.Proto == "tcp" || c.Proto == "tcp6"
	isUDP := c.Proto == "udp" || c.Proto == "udp6"

	if isTCP && !m.showTCP {
		return false
	}
	if isUDP && !m.showUDP {
		return false
	}

	isListening := c.State == "LISTEN"
	isEstablished := c.State == "ESTABLISHED"
	isOther := !isListening && !isEstablished

	if isListening && !m.showListening {
		return false
	}
	if isEstablished && !m.showEstablished {
		return false
	}
	if isOther && !m.showOther {
		return false
	}

	return true
}

func (m model) matchesSearch(c collector.Connection) bool {
	return containsIgnoreCase(c.Process, m.searchQuery) ||
		containsIgnoreCase(c.Laddr, m.searchQuery) ||
		containsIgnoreCase(c.Raddr, m.searchQuery) ||
		containsIgnoreCase(c.User, m.searchQuery) ||
		containsIgnoreCase(c.Proto, m.searchQuery) ||
		containsIgnoreCase(c.State, m.searchQuery)
}

func (m model) isWatched(pid int) bool {
	if pid <= 0 {
		return false
	}
	return m.watchedPIDs[pid]
}

func (m *model) toggleWatch(pid int) {
	if pid <= 0 {
		return
	}
	if m.watchedPIDs[pid] {
		delete(m.watchedPIDs, pid)
	} else {
		m.watchedPIDs[pid] = true
	}
}

func (m model) watchedCount() int {
	return len(m.watchedPIDs)
}

// currentState returns the current view options as a TUIState for persistence
func (m model) currentState() state.TUIState {
	return state.TUIState{
		ShowTCP:         m.showTCP,
		ShowUDP:         m.showUDP,
		ShowListening:   m.showListening,
		ShowEstablished: m.showEstablished,
		ShowOther:       m.showOther,
		SortField:       m.sortField,
		SortReverse:     m.sortReverse,
		ResolveAddrs:    m.resolveAddrs,
		ResolvePorts:    m.resolvePorts,
	}
}

// saveState persists current view options in the background
func (m model) saveState() {
	if m.rememberState {
		state.SaveAsync(m.currentState())
	}
}
