package collector

import (
	"strings"
	"time"
)

type FilterOptions struct {
	Proto     string
	State     string
	Pid       int
	Proc      string
	Lport     int
	Rport     int
	User      string
	UID       int
	Laddr     string
	Raddr     string
	Contains  string
	IPv4      bool
	IPv6      bool
	Interface string
	Mark      string
	Namespace string
	Inode     int64
	Since     time.Time
	SinceRel  time.Duration
}

func (f *FilterOptions) IsEmpty() bool {
	return f.Proto == "" && f.State == "" && f.Pid == 0 && f.Proc == "" &&
		f.Lport == 0 && f.Rport == 0 && f.User == "" && f.UID == 0 &&
		f.Laddr == "" && f.Raddr == "" && f.Contains == "" &&
		f.Interface == "" && f.Mark == "" && f.Namespace == "" && f.Inode == 0 &&
		f.Since.IsZero() && f.SinceRel == 0 && !f.IPv4 && !f.IPv6
}

func (f *FilterOptions) Matches(c Connection) bool {
	if f.Proto != "" && !matchesProto(c.Proto, f.Proto) {
		return false
	}
	if f.State != "" && !strings.EqualFold(c.State, f.State) {
		return false
	}
	if f.Pid != 0 && c.PID != f.Pid {
		return false
	}
	if f.Proc != "" && !containsIgnoreCase(c.Process, f.Proc) {
		return false
	}
	if f.Lport != 0 && c.Lport != f.Lport {
		return false
	}
	if f.Rport != 0 && c.Rport != f.Rport {
		return false
	}
	if f.User != "" && !strings.EqualFold(c.User, f.User) {
		return false
	}
	if f.UID != 0 && c.UID != f.UID {
		return false
	}
	if f.Laddr != "" && !strings.EqualFold(c.Laddr, f.Laddr) {
		return false
	}
	if f.Raddr != "" && !strings.EqualFold(c.Raddr, f.Raddr) {
		return false
	}
	if f.Contains != "" && !matchesContains(c, f.Contains) {
		return false
	}
	if f.IPv4 && c.IPVersion != "IPv4" {
		return false
	}
	if f.IPv6 && c.IPVersion != "IPv6" {
		return false
	}
	if f.Interface != "" && !strings.EqualFold(c.Interface, f.Interface) {
		return false
	}
	if f.Mark != "" && !strings.EqualFold(c.Mark, f.Mark) {
		return false
	}
	if f.Namespace != "" && !strings.EqualFold(c.Namespace, f.Namespace) {
		return false
	}
	if f.Inode != 0 && c.Inode != f.Inode {
		return false
	}
	if !f.Since.IsZero() && c.TS.Before(f.Since) {
		return false
	}
	if f.SinceRel != 0 {
		threshold := time.Now().Add(-f.SinceRel)
		if c.TS.Before(threshold) {
			return false
		}
	}

	return true
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// checks if a connection's protocol matches the filter.
// treats "tcp" as matching "tcp" and "tcp6", same for "udp"/"udp6"
func matchesProto(connProto, filterProto string) bool {
	connLower := strings.ToLower(connProto)
	filterLower := strings.ToLower(filterProto)

	// exact match
	if connLower == filterLower {
		return true
	}

	// "tcp" matches both "tcp" and "tcp6"
	if filterLower == "tcp" && (connLower == "tcp" || connLower == "tcp6") {
		return true
	}

	// "udp" matches both "udp" and "udp6"
	if filterLower == "udp" && (connLower == "udp" || connLower == "udp6") {
		return true
	}

	return false
}

func matchesContains(c Connection, query string) bool {
	q := strings.ToLower(query)
	return containsIgnoreCase(c.Process, q) ||
		containsIgnoreCase(c.Laddr, q) ||
		containsIgnoreCase(c.Raddr, q) ||
		containsIgnoreCase(c.User, q)
}

// ParseTimeFilter parses a time filter string (RFC3339 or relative like "5s", "2m", "1h")
func ParseTimeFilter(timeStr string) (time.Time, time.Duration, error) {
	// Try parsing as RFC3339 first
	if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
		return t, 0, nil
	}

	// Try parsing as relative duration
	if dur, err := time.ParseDuration(timeStr); err == nil {
		return time.Time{}, dur, nil
	}

	return time.Time{}, 0, nil // Invalid format, but don't error
}

