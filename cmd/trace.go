package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/karol-broda/snitch/internal/collector"
	"github.com/karol-broda/snitch/internal/config"
	"github.com/karol-broda/snitch/internal/resolver"

	"github.com/spf13/cobra"
)

type TraceEvent struct {
	Timestamp  time.Time             `json:"ts"`
	Event      string                `json:"event"` // "opened" or "closed"
	Connection collector.Connection  `json:"connection"`
}

var (
	traceInterval     time.Duration
	traceCount        int
	traceOutputFormat string
	traceTimestamp    bool
)

var traceCmd = &cobra.Command{
	Use:   "trace [filters...]",
	Short: "Print new/closed connections as they happen",
	Long: `Print new/closed connections as they happen.
	
Filters are specified in key=value format. For example:
  snitch trace proto=tcp state=established

Available filters:
  proto, state, pid, proc, lport, rport, user, laddr, raddr, contains
`,
	Run: func(cmd *cobra.Command, args []string) {
		runTraceCommand(args)
	},
}

func runTraceCommand(args []string) {
	cfg := config.Get()
	
	// configure resolver with cache setting
	effectiveNoCache := noCache || !cfg.Defaults.DNSCache
	resolver.SetNoCache(effectiveNoCache)

	filters, err := BuildFilters(args)
	if err != nil {
		log.Fatalf("Error parsing filters: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle interrupts gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	// Track connections using a key-based approach
	currentConnections := make(map[string]collector.Connection)
	
	// Get initial snapshot
	initialConnections, err := collector.GetConnections()
	if err != nil {
		log.Printf("Error getting initial connections: %v", err)
	} else {
		filteredInitial := collector.FilterConnections(initialConnections, filters)
		for _, conn := range filteredInitial {
			key := getConnectionKey(conn)
			currentConnections[key] = conn
		}
	}

	ticker := time.NewTicker(traceInterval)
	defer ticker.Stop()

	eventCount := 0
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			newConnections, err := collector.GetConnections()
			if err != nil {
				log.Printf("Error getting connections: %v", err)
				continue
			}

			filteredNew := collector.FilterConnections(newConnections, filters)
			newConnectionsMap := make(map[string]collector.Connection)
			
			// Build map of new connections
			for _, conn := range filteredNew {
				key := getConnectionKey(conn)
				newConnectionsMap[key] = conn
			}

			// Find newly opened connections
			for key, conn := range newConnectionsMap {
				if _, exists := currentConnections[key]; !exists {
					event := TraceEvent{
						Timestamp:  time.Now(),
						Event:      "opened",
						Connection: conn,
					}
					printTraceEvent(event)
					eventCount++
				}
			}

			// Find closed connections
			for key, conn := range currentConnections {
				if _, exists := newConnectionsMap[key]; !exists {
					event := TraceEvent{
						Timestamp:  time.Now(),
						Event:      "closed",
						Connection: conn,
					}
					printTraceEvent(event)
					eventCount++
				}
			}

			// Update current state
			currentConnections = newConnectionsMap

			if traceCount > 0 && eventCount >= traceCount {
				return
			}
		}
	}
}

func getConnectionKey(conn collector.Connection) string {
	// Create a unique key for a connection based on protocol, addresses, ports, and PID
	// This helps identify the same logical connection across snapshots
	return fmt.Sprintf("%s|%s:%d|%s:%d|%d", conn.Proto, conn.Laddr, conn.Lport, conn.Raddr, conn.Rport, conn.PID)
}

func printTraceEvent(event TraceEvent) {
	switch traceOutputFormat {
	case "json":
		printTraceEventJSON(event)
	default:
		printTraceEventHuman(event)
	}
}

func printTraceEventJSON(event TraceEvent) {
	jsonOutput, err := json.Marshal(event)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
	fmt.Println(string(jsonOutput))
}

func printTraceEventHuman(event TraceEvent) {
	conn := event.Connection
	
	timestamp := ""
	if traceTimestamp {
		timestamp = event.Timestamp.Format("15:04:05.000") + " "
	}

	eventIcon := "+"
	if event.Event == "closed" {
		eventIcon = "-"
	}

	laddr := conn.Laddr
	raddr := conn.Raddr
	lportStr := fmt.Sprintf("%d", conn.Lport)
	rportStr := fmt.Sprintf("%d", conn.Rport)
	
	// apply name resolution
	if resolveAddrs {
		if resolvedLaddr := resolver.ResolveAddr(conn.Laddr); resolvedLaddr != conn.Laddr {
			laddr = resolvedLaddr
		}
		if resolvedRaddr := resolver.ResolveAddr(conn.Raddr); resolvedRaddr != conn.Raddr && conn.Raddr != "*" && conn.Raddr != "" {
			raddr = resolvedRaddr
		}
	}
	if resolvePorts {
		if resolvedLport := resolver.ResolvePort(conn.Lport, conn.Proto); resolvedLport != fmt.Sprintf("%d", conn.Lport) {
			lportStr = resolvedLport
		}
		if resolvedRport := resolver.ResolvePort(conn.Rport, conn.Proto); resolvedRport != fmt.Sprintf("%d", conn.Rport) && conn.Rport != 0 {
			rportStr = resolvedRport
		}
	}

	// Format the connection string
	var connStr string
	if conn.Raddr != "" && conn.Raddr != "*" {
		connStr = fmt.Sprintf("%s:%s->%s:%s", laddr, lportStr, raddr, rportStr)
	} else {
		connStr = fmt.Sprintf("%s:%s", laddr, lportStr)
	}

	process := ""
	if conn.Process != "" {
		process = fmt.Sprintf(" (%s[%d])", conn.Process, conn.PID)
	}

	protocol := strings.ToUpper(conn.Proto)
	state := conn.State
	if state == "" {
		state = "UNKNOWN"
	}

	fmt.Printf("%s%s %s %s %s%s\n", timestamp, eventIcon, protocol, state, connStr, process)
}

func init() {
	rootCmd.AddCommand(traceCmd)

	// trace-specific flags
	traceCmd.Flags().DurationVarP(&traceInterval, "interval", "i", time.Second, "Polling interval (e.g., 500ms, 2s)")
	traceCmd.Flags().IntVarP(&traceCount, "count", "c", 0, "Number of events to capture (0 = unlimited)")
	traceCmd.Flags().StringVarP(&traceOutputFormat, "output", "o", "human", "Output format (human, json)")
	traceCmd.Flags().BoolVar(&traceTimestamp, "ts", false, "Include timestamp in output")

	// shared flags
	addFilterFlags(traceCmd)
	addResolutionFlags(traceCmd)
}
