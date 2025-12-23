package cmd

import (
	"fmt"
	"github.com/karol-broda/snitch/internal/collector"
	"github.com/karol-broda/snitch/internal/color"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

// Runtime holds the shared state for all commands.
// it handles common filter logic, fetching, and filtering connections.
type Runtime struct {
	// filter options built from flags and args
	Filters collector.FilterOptions

	// filtered connections ready for rendering
	Connections []collector.Connection

	// common settings
	ColorMode    string
	ResolveAddrs bool
	ResolvePorts bool
}

// shared filter flags - used by all commands
var (
	filterTCP    bool
	filterUDP    bool
	filterListen bool
	filterEstab  bool
	filterIPv4   bool
	filterIPv6   bool
)

// BuildFilters constructs FilterOptions from command args and shortcut flags.
func BuildFilters(args []string) (collector.FilterOptions, error) {
	filters, err := ParseFilterArgs(args)
	if err != nil {
		return filters, err
	}

	// apply ipv4/ipv6 flags
	filters.IPv4 = filterIPv4
	filters.IPv6 = filterIPv6

	// apply protocol shortcut flags
	if filterTCP && !filterUDP {
		filters.Proto = "tcp"
	} else if filterUDP && !filterTCP {
		filters.Proto = "udp"
	}

	// apply state shortcut flags
	if filterListen && !filterEstab {
		filters.State = "LISTEN"
	} else if filterEstab && !filterListen {
		filters.State = "ESTABLISHED"
	}

	return filters, nil
}

// FetchConnections gets connections from the collector and applies filters.
func FetchConnections(filters collector.FilterOptions) ([]collector.Connection, error) {
	connections, err := collector.GetConnections()
	if err != nil {
		return nil, err
	}

	return collector.FilterConnections(connections, filters), nil
}

// NewRuntime creates a runtime with fetched and filtered connections.
func NewRuntime(args []string, colorMode string) (*Runtime, error) {
	color.Init(colorMode)

	filters, err := BuildFilters(args)
	if err != nil {
		return nil, fmt.Errorf("failed to parse filters: %w", err)
	}

	connections, err := FetchConnections(filters)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch connections: %w", err)
	}

	return &Runtime{
		Filters:      filters,
		Connections:  connections,
		ColorMode:    colorMode,
		ResolveAddrs: resolveAddrs,
		ResolvePorts: resolvePorts,
	}, nil
}

// SortConnections sorts the runtime's connections in place.
func (r *Runtime) SortConnections(opts collector.SortOptions) {
	collector.SortConnections(r.Connections, opts)
}

// ParseFilterArgs parses key=value filter arguments.
// exported for testing.
func ParseFilterArgs(args []string) (collector.FilterOptions, error) {
	filters := collector.FilterOptions{}
	for _, arg := range args {
		parts := strings.SplitN(arg, "=", 2)
		if len(parts) != 2 {
			return filters, fmt.Errorf("invalid filter format: %s (expected key=value)", arg)
		}
		key, value := parts[0], parts[1]
		if err := applyFilter(&filters, key, value); err != nil {
			return filters, err
		}
	}
	return filters, nil
}

// applyFilter applies a single key=value filter to FilterOptions.
func applyFilter(filters *collector.FilterOptions, key, value string) error {
	switch strings.ToLower(key) {
	case "proto":
		filters.Proto = value
	case "state":
		filters.State = value
	case "pid":
		pid, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid pid value: %s", value)
		}
		filters.Pid = pid
	case "proc":
		filters.Proc = value
	case "lport":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid lport value: %s", value)
		}
		filters.Lport = port
	case "rport":
		port, err := strconv.Atoi(value)
		if err != nil {
			return fmt.Errorf("invalid rport value: %s", value)
		}
		filters.Rport = port
	case "user":
		uid, err := strconv.Atoi(value)
		if err == nil {
			filters.UID = uid
		} else {
			filters.User = value
		}
	case "laddr":
		filters.Laddr = value
	case "raddr":
		filters.Raddr = value
	case "contains":
		filters.Contains = value
	case "if", "interface":
		filters.Interface = value
	case "mark":
		filters.Mark = value
	case "namespace":
		filters.Namespace = value
	case "inode":
		inode, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid inode value: %s", value)
		}
		filters.Inode = inode
	case "since":
		since, sinceRel, err := collector.ParseTimeFilter(value)
		if err != nil {
			return fmt.Errorf("invalid since value: %s", value)
		}
		filters.Since = since
		filters.SinceRel = sinceRel
	default:
		return fmt.Errorf("unknown filter key: %s", key)
	}
	return nil
}

// FilterFlagsHelp returns the help text for common filter flags.
const FilterFlagsHelp = `
Filters are specified in key=value format. For example:
  snitch ls proto=tcp state=established

Available filters:
  proto, state, pid, proc, lport, rport, user, laddr, raddr, contains, if, mark, namespace, inode, since`

// addFilterFlags adds the common filter flags to a command.
func addFilterFlags(cmd *cobra.Command) {
	cmd.Flags().BoolVarP(&filterTCP, "tcp", "t", false, "Show only TCP connections")
	cmd.Flags().BoolVarP(&filterUDP, "udp", "u", false, "Show only UDP connections")
	cmd.Flags().BoolVarP(&filterListen, "listen", "l", false, "Show only listening sockets")
	cmd.Flags().BoolVarP(&filterEstab, "established", "e", false, "Show only established connections")
	cmd.Flags().BoolVarP(&filterIPv4, "ipv4", "4", false, "Only show IPv4 connections")
	cmd.Flags().BoolVarP(&filterIPv6, "ipv6", "6", false, "Only show IPv6 connections")
}

