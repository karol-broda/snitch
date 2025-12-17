package collector

import (
	"sort"
	"strings"
)

// SortField represents a field to sort by
type SortField string

const (
	SortByPID        SortField = "pid"
	SortByProcess    SortField = "process"
	SortByUser       SortField = "user"
	SortByProto      SortField = "proto"
	SortByState      SortField = "state"
	SortByLaddr      SortField = "laddr"
	SortByLport      SortField = "lport"
	SortByRaddr      SortField = "raddr"
	SortByRport      SortField = "rport"
	SortByInterface  SortField = "if"
	SortByRxBytes    SortField = "rx_bytes"
	SortByTxBytes    SortField = "tx_bytes"
	SortByRttMs      SortField = "rtt_ms"
	SortByTimestamp  SortField = "ts"
)

// SortDirection represents ascending or descending order
type SortDirection int

const (
	SortAsc SortDirection = iota
	SortDesc
)

// SortOptions configures how connections are sorted
type SortOptions struct {
	Field     SortField
	Direction SortDirection
}

// ParseSortOptions parses a sort string like "pid:desc" or "lport"
func ParseSortOptions(s string) SortOptions {
	if s == "" {
		return SortOptions{Field: SortByLport, Direction: SortAsc}
	}

	parts := strings.SplitN(s, ":", 2)
	field := SortField(strings.ToLower(parts[0]))
	direction := SortAsc

	if len(parts) > 1 && strings.ToLower(parts[1]) == "desc" {
		direction = SortDesc
	}

	return SortOptions{Field: field, Direction: direction}
}

// SortConnections sorts a slice of connections in place
func SortConnections(conns []Connection, opts SortOptions) {
	if len(conns) < 2 {
		return
	}

	sort.SliceStable(conns, func(i, j int) bool {
		less := compareConnections(conns[i], conns[j], opts.Field)
		if opts.Direction == SortDesc {
			return !less
		}
		return less
	})
}

func compareConnections(a, b Connection, field SortField) bool {
	switch field {
	case SortByPID:
		return a.PID < b.PID
	case SortByProcess:
		return strings.ToLower(a.Process) < strings.ToLower(b.Process)
	case SortByUser:
		return strings.ToLower(a.User) < strings.ToLower(b.User)
	case SortByProto:
		return a.Proto < b.Proto
	case SortByState:
		return stateOrder(a.State) < stateOrder(b.State)
	case SortByLaddr:
		return a.Laddr < b.Laddr
	case SortByLport:
		return a.Lport < b.Lport
	case SortByRaddr:
		return a.Raddr < b.Raddr
	case SortByRport:
		return a.Rport < b.Rport
	case SortByInterface:
		return a.Interface < b.Interface
	case SortByRxBytes:
		return a.RxBytes < b.RxBytes
	case SortByTxBytes:
		return a.TxBytes < b.TxBytes
	case SortByRttMs:
		return a.RttMs < b.RttMs
	case SortByTimestamp:
		return a.TS.Before(b.TS)
	default:
		return a.Lport < b.Lport
	}
}

// stateOrder returns a numeric order for connection states
// puts LISTEN first, then ESTABLISHED, then others
func stateOrder(state string) int {
	order := map[string]int{
		"LISTEN":      0,
		"UNCONNECTED": 1, // udp sockets bound but not connected to a specific peer
		"ESTABLISHED": 2,
		"SYN_SENT":    3,
		"SYN_RECV":    4,
		"FIN_WAIT1":   5,
		"FIN_WAIT2":   6,
		"TIME_WAIT":   7,
		"CLOSE_WAIT":  8,
		"LAST_ACK":    9,
		"CLOSING":     10,
		"CLOSED":      11,
	}

	if o, exists := order[strings.ToUpper(state)]; exists {
		return o
	}
	return 99
}

