package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"snitch/internal/collector"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
)

type StatsData struct {
	Timestamp time.Time            `json:"ts"`
	Total     int                  `json:"total"`
	ByProto   map[string]int       `json:"by_proto"`
	ByState   map[string]int       `json:"by_state"`
	ByProc    []ProcessStats       `json:"by_proc"`
	ByIf      []InterfaceStats     `json:"by_if"`
}

type ProcessStats struct {
	PID     int    `json:"pid"`
	Process string `json:"process"`
	Count   int    `json:"count"`
}

type InterfaceStats struct {
	Interface string `json:"if"`
	Count     int    `json:"count"`
}

var (
	statsOutputFormat string
	statsInterval     time.Duration
	statsCount        int
	statsNoHeaders    bool
	statsTCP          bool
	statsUDP          bool
	statsListen       bool
	statsEstab        bool
)

var statsCmd = &cobra.Command{
	Use:   "stats [filters...]",
	Short: "Aggregated connection counters",
	Long: `Aggregated connection counters.
	
Filters are specified in key=value format. For example:
  snitch stats proto=tcp state=listening

Available filters:
  proto, state, pid, proc, lport, rport, user, laddr, raddr, contains
`,
	Run: func(cmd *cobra.Command, args []string) {
		runStatsCommand(args)
	},
}

func runStatsCommand(args []string) {
	filters, err := parseFilters(args)
	if err != nil {
		log.Fatalf("Error parsing filters: %v", err)
	}
	filters.IPv4 = ipv4
	filters.IPv6 = ipv6

	// apply shortcut flags
	if statsTCP && !statsUDP {
		filters.Proto = "tcp"
	} else if statsUDP && !statsTCP {
		filters.Proto = "udp"
	}
	if statsListen && !statsEstab {
		filters.State = "LISTEN"
	} else if statsEstab && !statsListen {
		filters.State = "ESTABLISHED"
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

	count := 0
	for {
		stats, err := generateStats(filters)
		if err != nil {
			log.Printf("Error generating stats: %v", err)
			if statsCount > 0 || statsInterval == 0 {
				return
			}
			time.Sleep(statsInterval)
			continue
		}

		switch statsOutputFormat {
		case "json":
			printStatsJSON(stats)
		case "csv":
			printStatsCSV(stats, !statsNoHeaders && count == 0)
		default:
			printStatsTable(stats, !statsNoHeaders && count == 0)
		}

		count++
		if statsCount > 0 && count >= statsCount {
			return
		}

		if statsInterval == 0 {
			return // One-shot mode
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(statsInterval):
			continue
		}
	}
}

func generateStats(filters collector.FilterOptions) (*StatsData, error) {
	connections, err := collector.GetConnections()
	if err != nil {
		return nil, err
	}

	filteredConnections := collector.FilterConnections(connections, filters)

	stats := &StatsData{
		Timestamp: time.Now(),
		Total:     len(filteredConnections),
		ByProto:   make(map[string]int),
		ByState:   make(map[string]int),
		ByProc:    make([]ProcessStats, 0),
		ByIf:      make([]InterfaceStats, 0),
	}

	procCounts := make(map[string]ProcessStats)
	ifCounts := make(map[string]int)

	for _, conn := range filteredConnections {
		// Count by protocol
		stats.ByProto[conn.Proto]++

		// Count by state
		stats.ByState[conn.State]++

		// Count by process
		if conn.Process != "" {
			key := fmt.Sprintf("%d-%s", conn.PID, conn.Process)
			if existing, ok := procCounts[key]; ok {
				existing.Count++
				procCounts[key] = existing
			} else {
				procCounts[key] = ProcessStats{
					PID:     conn.PID,
					Process: conn.Process,
					Count:   1,
				}
			}
		}

		// Count by interface (placeholder since we don't have interface data yet)
		if conn.Interface != "" {
			ifCounts[conn.Interface]++
		}
	}

	// Convert process map to sorted slice
	for _, procStats := range procCounts {
		stats.ByProc = append(stats.ByProc, procStats)
	}
	sort.Slice(stats.ByProc, func(i, j int) bool {
		return stats.ByProc[i].Count > stats.ByProc[j].Count
	})

	// Convert interface map to sorted slice
	for iface, count := range ifCounts {
		stats.ByIf = append(stats.ByIf, InterfaceStats{
			Interface: iface,
			Count:     count,
		})
	}
	sort.Slice(stats.ByIf, func(i, j int) bool {
		return stats.ByIf[i].Count > stats.ByIf[j].Count
	})

	return stats, nil
}

func printStatsJSON(stats *StatsData) {
	jsonOutput, err := json.MarshalIndent(stats, "", "  ")
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
	fmt.Println(string(jsonOutput))
}

func printStatsCSV(stats *StatsData, headers bool) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	if headers {
		_ = writer.Write([]string{"timestamp", "metric", "key", "value"})
	}

	ts := stats.Timestamp.Format(time.RFC3339)

	_ = writer.Write([]string{ts, "total", "", strconv.Itoa(stats.Total)})

	for proto, count := range stats.ByProto {
		_ = writer.Write([]string{ts, "proto", proto, strconv.Itoa(count)})
	}

	for state, count := range stats.ByState {
		_ = writer.Write([]string{ts, "state", state, strconv.Itoa(count)})
	}

	for _, proc := range stats.ByProc {
		_ = writer.Write([]string{ts, "process", proc.Process, strconv.Itoa(proc.Count)})
	}

	for _, iface := range stats.ByIf {
		_ = writer.Write([]string{ts, "interface", iface.Interface, strconv.Itoa(iface.Count)})
	}
}

func printStatsTable(stats *StatsData, headers bool) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer w.Flush()

	if headers {
		fmt.Fprintf(w, "TIMESTAMP\t%s\n", stats.Timestamp.Format(time.RFC3339))
		fmt.Fprintf(w, "TOTAL CONNECTIONS\t%d\n", stats.Total)
		fmt.Fprintln(w)
	}

	// Protocol breakdown
	if len(stats.ByProto) > 0 {
		if headers {
			fmt.Fprintln(w, "BY PROTOCOL:")
			fmt.Fprintln(w, "PROTO\tCOUNT")
		}
		protocols := make([]string, 0, len(stats.ByProto))
		for proto := range stats.ByProto {
			protocols = append(protocols, proto)
		}
		sort.Strings(protocols)
		for _, proto := range protocols {
			fmt.Fprintf(w, "%s\t%d\n", strings.ToUpper(proto), stats.ByProto[proto])
		}
		fmt.Fprintln(w)
	}

	// State breakdown
	if len(stats.ByState) > 0 {
		if headers {
			fmt.Fprintln(w, "BY STATE:")
			fmt.Fprintln(w, "STATE\tCOUNT")
		}
		states := make([]string, 0, len(stats.ByState))
		for state := range stats.ByState {
			states = append(states, state)
		}
		sort.Strings(states)
		for _, state := range states {
			fmt.Fprintf(w, "%s\t%d\n", state, stats.ByState[state])
		}
		fmt.Fprintln(w)
	}

	// Process breakdown (top 10)
	if len(stats.ByProc) > 0 {
		if headers {
			fmt.Fprintln(w, "BY PROCESS (TOP 10):")
			fmt.Fprintln(w, "PID\tPROCESS\tCOUNT")
		}
		limit := 10
		if len(stats.ByProc) < limit {
			limit = len(stats.ByProc)
		}
		for i := 0; i < limit; i++ {
			proc := stats.ByProc[i]
			fmt.Fprintf(w, "%d\t%s\t%d\n", proc.PID, proc.Process, proc.Count)
		}
	}
}

func init() {
	rootCmd.AddCommand(statsCmd)
	statsCmd.Flags().StringVarP(&statsOutputFormat, "output", "o", "table", "Output format (table, json, csv)")
	statsCmd.Flags().DurationVarP(&statsInterval, "interval", "i", 0, "Refresh interval (0 = one-shot)")
	statsCmd.Flags().IntVarP(&statsCount, "count", "c", 0, "Number of iterations (0 = unlimited)")
	statsCmd.Flags().BoolVar(&statsNoHeaders, "no-headers", false, "Omit headers for table/csv output")
	statsCmd.Flags().BoolVarP(&ipv4, "ipv4", "4", false, "Only show IPv4 connections")
	statsCmd.Flags().BoolVarP(&ipv6, "ipv6", "6", false, "Only show IPv6 connections")

	// shortcut filters
	statsCmd.Flags().BoolVarP(&statsTCP, "tcp", "t", false, "Show only TCP connections")
	statsCmd.Flags().BoolVarP(&statsUDP, "udp", "u", false, "Show only UDP connections")
	statsCmd.Flags().BoolVarP(&statsListen, "listen", "l", false, "Show only listening sockets")
	statsCmd.Flags().BoolVarP(&statsEstab, "established", "e", false, "Show only established connections")
}
