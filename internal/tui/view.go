package tui

import (
	"fmt"
	"github.com/karol-broda/snitch/internal/collector"
	"github.com/karol-broda/snitch/internal/resolver"
	"strconv"
	"strings"
	"time"

	"github.com/mattn/go-runewidth"
)

func (m model) renderMain() string {
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString(m.renderTitle())
	b.WriteString("\n")
	b.WriteString(m.renderFilters())
	b.WriteString("\n\n")
	b.WriteString(m.renderTableHeader())
	b.WriteString(m.renderSeparator())
	b.WriteString(m.renderConnections())
	b.WriteString("\n")
	b.WriteString(m.renderStatusLine())

	return b.String()
}

func (m model) renderTitle() string {
	visible := m.visibleConnections()
	total := len(m.connections)

	left := m.theme.Styles.Header.Render("snitch")

	ago := time.Since(m.lastRefresh).Round(time.Millisecond * 100)
	right := m.theme.Styles.Normal.Render(fmt.Sprintf("%d/%d connections  %s %s", len(visible), total, SymbolRefresh, formatDuration(ago)))

	w := m.safeWidth()
	gap := w - len(stripAnsi(left)) - len(stripAnsi(right)) - 2
	if gap < 0 {
		gap = 0
	}

	return "  " + left + strings.Repeat(" ", gap) + right
}

func (m model) renderFilters() string {
	var parts []string

	parts = append(parts, m.renderFilterLabel("t", "cp", m.showTCP))
	parts = append(parts, m.renderFilterLabel("u", "dp", m.showUDP))

	parts = append(parts, m.theme.Styles.Border.Render(BoxVertical))

	parts = append(parts, m.renderFilterLabel("l", "isten", m.showListening))
	parts = append(parts, m.renderFilterLabel("e", "stab", m.showEstablished))
	parts = append(parts, m.renderFilterLabel("o", "ther", m.showOther))

	left := "  " + strings.Join(parts, "  ")

	sortLabel := sortFieldLabel(m.sortField)
	sortDir := SymbolArrowUp
	if m.sortReverse {
		sortDir = SymbolArrowDown
	}

	var right string
	if m.searchActive {
		right = m.theme.Styles.Warning.Render(fmt.Sprintf("/%s▌", m.searchQuery))
	} else if m.searchQuery != "" {
		right = m.theme.Styles.Normal.Render(fmt.Sprintf("filter: %s", m.searchQuery))
	} else {
		right = m.theme.Styles.Normal.Render(fmt.Sprintf("sort: %s %s", sortLabel, sortDir))
	}

	w := m.safeWidth()
	gap := w - len(stripAnsi(left)) - len(stripAnsi(right)) - 2
	if gap < 0 {
		gap = 0
	}

	return left + strings.Repeat(" ", gap) + right + "  "
}

func (m model) renderTableHeader() string {
	cols := m.columnWidths()

	header := fmt.Sprintf("  %-*s  %-*s  %-*s  %-*s  %-*s  %s",
		cols.process, "PROCESS",
		cols.port, "PORT",
		cols.proto, "PROTO",
		cols.state, "STATE",
		cols.local, "LOCAL",
		"REMOTE")

	return m.theme.Styles.Header.Render(header) + "\n"
}

func (m model) renderFilterLabel(firstChar, rest string, active bool) string {
	baseStyle := m.theme.Styles.Normal
	if active {
		baseStyle = m.theme.Styles.Success
	}

	underlinedFirst := baseStyle.Underline(true).Render(firstChar)
	restPart := baseStyle.Render(rest)

	return underlinedFirst + restPart
}

func (m model) renderSeparator() string {
	w := m.width - 4
	if w < 1 {
		w = 76
	}
	line := "  " + strings.Repeat(BoxHorizontal, w)
	return m.theme.Styles.Border.Render(line) + "\n"
}

func (m model) renderConnections() string {
	var b strings.Builder
	visible := m.visibleConnections()
	pageSize := m.pageSize()

	if len(visible) == 0 {
		b.WriteString("  " + m.theme.Styles.Normal.Render("no connections match filters") + "\n")
		for i := 1; i < pageSize; i++ {
			b.WriteString("\n")
		}
		return b.String()
	}

	start := m.scrollOffset(pageSize, len(visible))

	for i := 0; i < pageSize; i++ {
		idx := start + i
		if idx >= len(visible) {
			b.WriteString("\n")
			continue
		}

		isSelected := idx == m.cursor
		b.WriteString(m.renderRow(visible[idx], isSelected))
	}

	return b.String()
}

func (m model) renderRow(c collector.Connection, selected bool) string {
	cols := m.columnWidths()

	indicator := "  "
	if selected {
		indicator = m.theme.Styles.Success.Render(SymbolSelected + " ")
	} else if m.isWatched(c.PID) {
		indicator = m.theme.Styles.Watched.Render(SymbolWatched + " ")
	}

	process := truncate(c.Process, cols.process)
	if process == "" {
		process = SymbolDash
	}

	port := truncate(m.resolvePort(c.Lport, c.Proto), cols.port)
	proto := c.Proto
	state := c.State
	if state == "" {
		state = SymbolDash
	}

	local := truncate(m.resolveAddr(c.Laddr), cols.local)
	if local == "*" || local == "" {
		local = "*"
	}

	remote := truncate(m.formatRemote(c.Raddr, c.Rport, c.Proto), cols.remote)

	// apply styling
	protoStyled := m.theme.Styles.GetProtoStyle(proto).Render(fmt.Sprintf("%-*s", cols.proto, proto))
	stateStyled := m.theme.Styles.GetStateStyle(state).Render(fmt.Sprintf("%-*s", cols.state, truncate(state, cols.state)))

	row := fmt.Sprintf("%s%-*s  %-*s  %s  %s  %-*s  %s",
		indicator,
		cols.process, process,
		cols.port, port,
		protoStyled,
		stateStyled,
		cols.local, local,
		remote)

	if selected {
		return m.theme.Styles.Selected.Render(row) + "\n"
	}

	return m.theme.Styles.Normal.Render(row) + "\n"
}

func (m model) renderStatusLine() string {
	// show status message if present
	if m.statusMessage != "" {
		return "  " + m.theme.Styles.Warning.Render(m.statusMessage)
	}

	left := "  " + m.theme.Styles.Normal.Render("t/u proto  l/e/o state  n/N dns  w watch  K kill  s sort  / search  ? help  q quit")

	// show watched count if any
	if m.watchedCount() > 0 {
		watchedInfo := fmt.Sprintf("  watching: %d", m.watchedCount())
		left += m.theme.Styles.Watched.Render(watchedInfo)
	}

	// show dns resolution status
	var resolveStatus string
	if m.resolveAddrs && m.resolvePorts {
		resolveStatus = "all"
	} else if m.resolveAddrs {
		resolveStatus = "addrs"
	} else if m.resolvePorts {
		resolveStatus = "ports"
	} else {
		resolveStatus = "off"
	}
	if resolveStatus != "addrs" { // addrs is the default, don't show
		left += m.theme.Styles.Normal.Render(fmt.Sprintf("  dns: %s", resolveStatus))
	}

	return left
}

func (m model) renderError() string {
	return fmt.Sprintf("\n  %s\n\n  press q to quit\n",
		m.theme.Styles.Error.Render(fmt.Sprintf("error: %v", m.err)))
}

func (m model) renderHelp() string {
	help := `
  navigation
  ──────────
  j/k ↑/↓      move cursor
  g/G          jump to top/bottom
  ctrl+d/u     half page down/up
  enter        show connection details

  filters
  ───────
  t            toggle tcp
  u            toggle udp
  l            toggle listening
  e            toggle established
  o            toggle other states
  a            reset all filters

  sorting
  ───────
  s            cycle sort field
  S            reverse sort order

  display
  ───────
  n            toggle address resolution (dns)
  N            toggle port resolution (service names)

  process management
  ──────────────────
  w            watch/unwatch process (highlight & track)
  W            clear all watched processes
  K            kill process (with confirmation)

  other
  ─────
  /            search
  r            refresh now
  q            quit

  press ? or esc to close
`
	return m.theme.Styles.Normal.Render(help)
}

func (m model) renderDetail() string {
	if m.selected == nil {
		return ""
	}

	c := m.selected
	var b strings.Builder

	b.WriteString("\n")
	b.WriteString("  " + m.theme.Styles.Header.Render("connection details") + "\n")
	b.WriteString("  " + m.theme.Styles.Border.Render(strings.Repeat(BoxHorizontal, 40)) + "\n\n")

	localAddr := m.resolveAddr(c.Laddr)
	localPort := m.resolvePort(c.Lport, c.Proto)
	remoteAddr := m.resolveAddr(c.Raddr)
	remotePort := m.resolvePort(c.Rport, c.Proto)

	fields := []struct {
		label string
		value string
	}{
		{"process", c.Process},
		{"pid", fmt.Sprintf("%d", c.PID)},
		{"cmdline", c.Cmdline},
		{"cwd", c.Cwd},
		{"user", c.User},
		{"protocol", c.Proto},
		{"state", c.State},
		{"local", fmt.Sprintf("%s:%s", localAddr, localPort)},
		{"remote", fmt.Sprintf("%s:%s", remoteAddr, remotePort)},
		{"interface", c.Interface},
		{"inode", fmt.Sprintf("%d", c.Inode)},
	}

	for _, f := range fields {
		val := f.value
		if val == "" || val == "0" || val == ":0" {
			val = SymbolDash
		}
		line := fmt.Sprintf("  %-12s  %s\n", m.theme.Styles.Header.Render(f.label), val)
		b.WriteString(line)
	}

	b.WriteString("\n")
	b.WriteString("  " + m.theme.Styles.Normal.Render("press esc to close") + "\n")

	return b.String()
}

func (m model) renderKillModal() string {
	if m.killTarget == nil {
		return ""
	}

	c := m.killTarget
	processName := c.Process
	if processName == "" {
		processName = "(unknown)"
	}

	// count how many connections this process has
	connCount := 0
	for _, conn := range m.connections {
		if conn.PID == c.PID {
			connCount++
		}
	}

	// build modal content
	var lines []string
	lines = append(lines, "")
	lines = append(lines, m.theme.Styles.Error.Render("  "+SymbolWarning+"  KILL PROCESS?  "))
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  process:  %s", m.theme.Styles.Header.Render(processName)))
	lines = append(lines, fmt.Sprintf("  pid:      %s", m.theme.Styles.Header.Render(fmt.Sprintf("%d", c.PID))))
	lines = append(lines, fmt.Sprintf("  user:     %s", c.User))
	lines = append(lines, fmt.Sprintf("  conns:    %d", connCount))
	lines = append(lines, "")
	lines = append(lines, m.theme.Styles.Warning.Render("  sends SIGTERM to process"))
	if connCount > 1 {
		lines = append(lines, m.theme.Styles.Warning.Render(fmt.Sprintf("  will close all %d connections", connCount)))
	}
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("  %s confirm   %s cancel",
		m.theme.Styles.Success.Render("[y]"),
		m.theme.Styles.Error.Render("[n]")))
	lines = append(lines, "")

	return strings.Join(lines, "\n")
}

func (m model) overlayModal(background, modal string) string {
	bgLines := strings.Split(background, "\n")
	modalLines := strings.Split(modal, "\n")

	// find max modal line width using runewidth for proper unicode handling
	modalWidth := 0
	for _, line := range modalLines {
		w := stringWidth(line)
		if w > modalWidth {
			modalWidth = w
		}
	}
	modalWidth += 4 // padding for box

	modalHeight := len(modalLines)
	boxWidth := modalWidth + 2 // include border chars │ │

	// calculate modal position (centered)
	startRow := (m.height - modalHeight) / 2
	if startRow < 2 {
		startRow = 2
	}
	startCol := (m.width - boxWidth) / 2
	if startCol < 0 {
		startCol = 0
	}

	// build result
	result := make([]string, len(bgLines))
	copy(result, bgLines)

	// ensure we have enough lines
	for len(result) < startRow+modalHeight+2 {
		result = append(result, strings.Repeat(" ", m.width))
	}

	// helper to build a line with modal overlay
	buildLine := func(bgLine, modalContent string) string {
		modalVisibleWidth := stringWidth(modalContent)
		endCol := startCol + modalVisibleWidth

		leftBg := visibleSubstring(bgLine, 0, startCol)
		rightBg := visibleSubstring(bgLine, endCol, m.width)

		// pad left side if needed
		leftLen := stringWidth(leftBg)
		if leftLen < startCol {
			leftBg = leftBg + strings.Repeat(" ", startCol-leftLen)
		}

		return leftBg + modalContent + rightBg
	}

	// draw top border
	borderRow := startRow - 1
	if borderRow >= 0 && borderRow < len(result) {
		border := m.theme.Styles.Border.Render(BoxTopLeft + strings.Repeat(BoxHorizontal, modalWidth) + BoxTopRight)
		result[borderRow] = buildLine(result[borderRow], border)
	}

	// draw modal content with side borders
	for i, line := range modalLines {
		row := startRow + i
		if row >= 0 && row < len(result) {
			content := line
			padding := modalWidth - stringWidth(line)
			if padding > 0 {
				content = line + strings.Repeat(" ", padding)
			}
			boxedLine := m.theme.Styles.Border.Render(BoxVertical) + content + m.theme.Styles.Border.Render(BoxVertical)
			result[row] = buildLine(result[row], boxedLine)
		}
	}

	// draw bottom border
	bottomRow := startRow + modalHeight
	if bottomRow >= 0 && bottomRow < len(result) {
		border := m.theme.Styles.Border.Render(BoxBottomLeft + strings.Repeat(BoxHorizontal, modalWidth) + BoxBottomRight)
		result[bottomRow] = buildLine(result[bottomRow], border)
	}

	return strings.Join(result, "\n")
}

// stringWidth returns the display width of a string excluding ANSI codes
func stringWidth(s string) int {
	return runewidth.StringWidth(stripAnsi(s))
}

// visibleSubstring extracts a substring by visible column positions, preserving ANSI codes
func visibleSubstring(s string, start, end int) string {
	if start >= end {
		return ""
	}

	var result strings.Builder
	visiblePos := 0
	inEscape := false

	for _, r := range s {
		// detect start of ANSI escape sequence
		if r == '\x1b' {
			inEscape = true
			result.WriteRune(r)
			continue
		}

		if inEscape {
			result.WriteRune(r)
			// end of escape sequence is a letter
			if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') {
				inEscape = false
			}
			continue
		}

		// regular visible character
		w := runewidth.RuneWidth(r)
		if visiblePos >= start && visiblePos+w <= end {
			result.WriteRune(r)
		}
		visiblePos += w

		if visiblePos >= end {
			break
		}
	}

	return result.String()
}

func (m model) scrollOffset(pageSize, total int) int {
	if total <= pageSize {
		return 0
	}

	// keep cursor roughly centered
	offset := m.cursor - pageSize/2
	if offset < 0 {
		offset = 0
	}
	if offset > total-pageSize {
		offset = total - pageSize
	}
	return offset
}

type columns struct {
	process int
	port    int
	proto   int
	state   int
	local   int
	remote  int
}

func (m model) columnWidths() columns {
	// minimum widths (header lengths + padding)
	c := columns{
		process: 7,  // "PROCESS"
		port:    4,  // "PORT"
		proto:   5,  // "PROTO"
		state:   5,  // "STATE"
		local:   5,  // "LOCAL"
		remote:  6,  // "REMOTE"
	}

	// scan visible connections to find max content width for each column
	visible := m.visibleConnections()
	for _, conn := range visible {
		if len(conn.Process) > c.process {
			c.process = len(conn.Process)
		}

		port := m.resolvePort(conn.Lport, conn.Proto)
		if len(port) > c.port {
			c.port = len(port)
		}

		if len(conn.Proto) > c.proto {
			c.proto = len(conn.Proto)
		}

		if len(conn.State) > c.state {
			c.state = len(conn.State)
		}

		local := m.resolveAddr(conn.Laddr)
		if len(local) > c.local {
			c.local = len(local)
		}

		remote := m.formatRemote(conn.Raddr, conn.Rport, conn.Proto)
		if len(remote) > c.remote {
			c.remote = len(remote)
		}
	}

	// calculate total and available width
	spacing := 12 // 2 spaces between each of 6 columns
	indicator := 2
	margin := 2
	available := m.safeWidth() - spacing - indicator - margin

	total := c.process + c.port + c.proto + c.state + c.local + c.remote

	// if content fits, we're done
	if total <= available {
		return c
	}

	// content exceeds available space - need to shrink columns proportionally
	// fixed columns that shouldn't shrink much: port, proto, state
	fixedWidth := c.port + c.proto + c.state
	flexibleAvailable := available - fixedWidth

	// distribute flexible space between process, local, remote
	flexibleTotal := c.process + c.local + c.remote
	if flexibleTotal > 0 && flexibleAvailable > 0 {
		ratio := float64(flexibleAvailable) / float64(flexibleTotal)
		c.process = max(7, int(float64(c.process)*ratio))
		c.local = max(5, int(float64(c.local)*ratio))
		c.remote = max(6, int(float64(c.remote)*ratio))
	}

	return c
}

func (m model) safeWidth() int {
	if m.width < 80 {
		return 80
	}
	return m.width
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	if d < time.Minute {
		return fmt.Sprintf("%.1fs", d.Seconds())
	}
	return fmt.Sprintf("%.0fm", d.Minutes())
}

func (m model) resolveAddr(addr string) string {
	if !m.resolveAddrs {
		return addr
	}
	if addr == "" || addr == "*" {
		return addr
	}
	return resolver.ResolveAddr(addr)
}

func (m model) resolvePort(port int, proto string) string {
	if !m.resolvePorts {
		return strconv.Itoa(port)
	}
	return resolver.ResolvePort(port, proto)
}

func (m model) formatRemote(addr string, port int, proto string) string {
	if addr == "" || addr == "*" || port == 0 {
		return "-"
	}
	resolvedAddr := m.resolveAddr(addr)
	resolvedPort := m.resolvePort(port, proto)
	return fmt.Sprintf("%s:%s", resolvedAddr, resolvedPort)
}
