package tui

import (
	"fmt"
	"github.com/karol-broda/snitch/internal/collector"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func (m model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// search mode captures all input
	if m.searchActive {
		return m.handleSearchKey(msg)
	}

	// kill confirmation dialog
	if m.showKillConfirm {
		return m.handleKillConfirmKey(msg)
	}

	// detail view only allows closing
	if m.showDetail {
		return m.handleDetailKey(msg)
	}

	// help view only allows closing
	if m.showHelp {
		return m.handleHelpKey(msg)
	}

	return m.handleNormalKey(msg)
}

func (m model) handleSearchKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.searchActive = false
		m.searchQuery = ""
	case "enter":
		m.searchActive = false
		m.cursor = 0
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
	default:
		if len(msg.String()) == 1 {
			m.searchQuery += msg.String()
		}
	}
	return m, nil
}

func (m model) handleDetailKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q":
		m.showDetail = false
		m.selected = nil
	}
	return m, nil
}

func (m model) handleHelpKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "enter", "q", "?":
		m.showHelp = false
	}
	return m, nil
}

func (m model) handleKillConfirmKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		if m.killTarget != nil && m.killTarget.PID > 0 {
			pid := m.killTarget.PID
			process := m.killTarget.Process
			m.showKillConfirm = false
			m.killTarget = nil
			return m, killProcess(pid, process)
		}
		m.showKillConfirm = false
		m.killTarget = nil
	case "n", "N", "esc", "q":
		m.showKillConfirm = false
		m.killTarget = nil
	}
	return m, nil
}

func (m model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Sequence(tea.ShowCursor, tea.Quit)

	// navigation
	case "j", "down":
		m.moveCursor(1)
	case "k", "up":
		m.moveCursor(-1)
	case "g":
		m.cursor = 0
	case "G":
		visible := m.visibleConnections()
		if len(visible) > 0 {
			m.cursor = len(visible) - 1
		}
	case "ctrl+d":
		m.moveCursor(m.pageSize() / 2)
	case "ctrl+u":
		m.moveCursor(-m.pageSize() / 2)
	case "ctrl+f", "pgdown":
		m.moveCursor(m.pageSize())
	case "ctrl+b", "pgup":
		m.moveCursor(-m.pageSize())

	// filter toggles
	case "t":
		m.showTCP = !m.showTCP
		m.clampCursor()
	case "u":
		m.showUDP = !m.showUDP
		m.clampCursor()
	case "l":
		m.showListening = !m.showListening
		m.clampCursor()
	case "e":
		m.showEstablished = !m.showEstablished
		m.clampCursor()
	case "o":
		m.showOther = !m.showOther
		m.clampCursor()
	case "a":
		m.showTCP = true
		m.showUDP = true
		m.showListening = true
		m.showEstablished = true
		m.showOther = true

	// sorting
	case "s":
		m.cycleSort()
	case "S":
		m.sortReverse = !m.sortReverse
		m.applySorting()

	// search
	case "/":
		m.searchActive = true
		m.searchQuery = ""

	// actions
	case "enter", " ":
		visible := m.visibleConnections()
		if m.cursor < len(visible) {
			conn := visible[m.cursor]
			m.selected = &conn
			m.showDetail = true
		}
	case "r":
		return m, m.fetchData()
	case "?":
		m.showHelp = true

	// watch/monitor process
	case "w":
		visible := m.visibleConnections()
		if m.cursor < len(visible) {
			conn := visible[m.cursor]
			if conn.PID > 0 {
				wasWatched := m.isWatched(conn.PID)
				m.toggleWatch(conn.PID)

				// count connections for this pid
				connCount := 0
				for _, c := range m.connections {
					if c.PID == conn.PID {
						connCount++
					}
				}

				if wasWatched {
					m.statusMessage = fmt.Sprintf("unwatched %s (pid %d)", conn.Process, conn.PID)
				} else if connCount > 1 {
					m.statusMessage = fmt.Sprintf("watching %s (pid %d) - %d connections", conn.Process, conn.PID, connCount)
				} else {
					m.statusMessage = fmt.Sprintf("watching %s (pid %d)", conn.Process, conn.PID)
				}
				m.statusExpiry = time.Now().Add(2 * time.Second)
				return m, clearStatusAfter(2 * time.Second)
			}
		}
	case "W":
		// clear all watched
		count := len(m.watchedPIDs)
		m.watchedPIDs = make(map[int]bool)
		if count > 0 {
			m.statusMessage = fmt.Sprintf("cleared %d watched processes", count)
			m.statusExpiry = time.Now().Add(2 * time.Second)
			return m, clearStatusAfter(2 * time.Second)
		}

	// kill process
	case "K":
		visible := m.visibleConnections()
		if m.cursor < len(visible) {
			conn := visible[m.cursor]
			if conn.PID > 0 {
				m.killTarget = &conn
				m.showKillConfirm = true
			}
		}

	// toggle address resolution
	case "n":
		m.resolveAddrs = !m.resolveAddrs
		if m.resolveAddrs {
			m.statusMessage = "address resolution: on"
		} else {
			m.statusMessage = "address resolution: off"
		}
		m.statusExpiry = time.Now().Add(2 * time.Second)
		return m, clearStatusAfter(2 * time.Second)

	// toggle port resolution
	case "N":
		m.resolvePorts = !m.resolvePorts
		if m.resolvePorts {
			m.statusMessage = "port resolution: on"
		} else {
			m.statusMessage = "port resolution: off"
		}
		m.statusExpiry = time.Now().Add(2 * time.Second)
		return m, clearStatusAfter(2 * time.Second)
	}

	return m, nil
}

func (m *model) moveCursor(delta int) {
	visible := m.visibleConnections()
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(visible) {
		m.cursor = len(visible) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m model) pageSize() int {
	size := m.height - 6
	if size < 1 {
		return 10
	}
	return size
}

func (m *model) cycleSort() {
	fields := []collector.SortField{
		collector.SortByLport,
		collector.SortByProcess,
		collector.SortByPID,
		collector.SortByState,
		collector.SortByProto,
	}

	for i, f := range fields {
		if f == m.sortField {
			m.sortField = fields[(i+1)%len(fields)]
			m.applySorting()
			return
		}
	}

	m.sortField = collector.SortByLport
	m.applySorting()
}

