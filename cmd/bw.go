package cmd

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/karol-broda/snitch/internal/bandwidth"
	"github.com/karol-broda/snitch/internal/errutil"
	"github.com/spf13/cobra"
)

var (
	bwInterval     time.Duration
	bwCount        int
	bwOutputFormat string
	bwNoHeaders    bool
	bwInterfaces   []string
)

type BWOutput struct {
	Timestamp  time.Time                 `json:"ts"`
	Duration   string                    `json:"duration"`
	Interfaces []bandwidth.InterfaceRate `json:"interfaces"`
	TotalRx    float64                   `json:"total_rx_bytes_per_sec"`
	TotalTx    float64                   `json:"total_tx_bytes_per_sec"`
}

var bwCmd = &cobra.Command{
	Use:   "bw [interfaces...]",
	Short: "Live bandwidth monitoring per interface",
	Long: `Live bandwidth monitoring showing RX/TX rates per network interface.

Examples:
  snitch bw                    # monitor all interfaces
  snitch bw eth0 wlan0         # monitor specific interfaces
  snitch bw -i 500ms           # faster refresh
  snitch bw -o json            # json output for scripting
`,
	Run: func(cmd *cobra.Command, args []string) {
		bwInterfaces = args
		runBWCommand()
	},
}

func runBWCommand() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		cancel()
	}()

	tracker := bandwidth.NewTracker()
	count := 0

	_, err := tracker.Update()
	if err != nil {
		log.Fatalf("Error getting initial stats: %v", err)
	}

	select {
	case <-ctx.Done():
		return
	case <-time.After(bwInterval):
	}

	for {
		result, err := tracker.Update()
		if err != nil {
			log.Printf("Error updating stats: %v", err)
			continue
		}

		filtered := filterInterfaces(result.Interfaces, bwInterfaces)
		output := buildBWOutput(result.Timestamp, result.Duration, filtered)

		switch bwOutputFormat {
		case "json":
			printBWJSON(output)
		case "csv":
			printBWCSV(output, !bwNoHeaders && count == 0)
		default:
			printBWTable(output, !bwNoHeaders && count == 0)
		}

		count++
		if bwCount > 0 && count >= bwCount {
			return
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(bwInterval):
			continue
		}
	}
}

func filterInterfaces(rates []bandwidth.InterfaceRate, filter []string) []bandwidth.InterfaceRate {
	if len(filter) == 0 {
		return rates
	}

	filterMap := make(map[string]bool)
	for _, name := range filter {
		filterMap[name] = true
	}

	filtered := make([]bandwidth.InterfaceRate, 0)
	for _, rate := range rates {
		if filterMap[rate.Name] {
			filtered = append(filtered, rate)
		}
	}
	return filtered
}

func buildBWOutput(timestamp time.Time, duration time.Duration, rates []bandwidth.InterfaceRate) BWOutput {
	sort.Slice(rates, func(i, j int) bool {
		return rates[i].Name < rates[j].Name
	})

	var totalRx, totalTx float64
	for _, r := range rates {
		totalRx += r.RxBytesPerSec
		totalTx += r.TxBytesPerSec
	}

	return BWOutput{
		Timestamp:  timestamp,
		Duration:   duration.Round(time.Millisecond).String(),
		Interfaces: rates,
		TotalRx:    totalRx,
		TotalTx:    totalTx,
	}
}

func printBWJSON(output BWOutput) {
	jsonBytes, err := json.Marshal(output)
	if err != nil {
		log.Printf("Error marshaling JSON: %v", err)
		return
	}
	fmt.Println(string(jsonBytes))
}

func printBWCSV(output BWOutput, headers bool) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	if headers {
		err := writer.Write([]string{"timestamp", "interface", "rx_bytes_sec", "tx_bytes_sec", "rx_total", "tx_total"})
		if err != nil {
			log.Printf("Error writing CSV headers: %v", err)
			return
		}
	}

	ts := output.Timestamp.Format(time.RFC3339)
	for _, iface := range output.Interfaces {
		err := writer.Write([]string{
			ts,
			iface.Name,
			fmt.Sprintf("%.0f", iface.RxBytesPerSec),
			fmt.Sprintf("%.0f", iface.TxBytesPerSec),
			fmt.Sprintf("%d", iface.RxBytes),
			fmt.Sprintf("%d", iface.TxBytes),
		})
		if err != nil {
			log.Printf("Error writing CSV row: %v", err)
			return
		}
	}
}

func printBWTable(output BWOutput, headers bool) {
	if !headers {
		fmt.Print("\033[H\033[2J")
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer errutil.Flush(w)

	if headers {
		errutil.Ignore(fmt.Fprintf(w, "INTERFACE\tRX/s\tTX/s\tRX TOTAL\tTX TOTAL\n"))
		errutil.Ignore(fmt.Fprintf(w, "---------\t----\t----\t--------\t--------\n"))
	}

	for _, iface := range output.Interfaces {
		errutil.Ignore(fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
			iface.Name,
			bandwidth.FormatRate(iface.RxBytesPerSec),
			bandwidth.FormatRate(iface.TxBytesPerSec),
			bandwidth.FormatBytes(float64(iface.RxBytes)),
			bandwidth.FormatBytes(float64(iface.TxBytes)),
		))
	}

	if len(output.Interfaces) > 1 {
		errutil.Ignore(fmt.Fprintf(w, "---------\t----\t----\t\t\n"))
		errutil.Ignore(fmt.Fprintf(w, "TOTAL\t%s\t%s\t\t\n",
			bandwidth.FormatRate(output.TotalRx),
			bandwidth.FormatRate(output.TotalTx),
		))
	}

	errutil.Ignore(fmt.Fprintf(w, "\n[%s] interval: %s\n",
		output.Timestamp.Format("15:04:05"),
		output.Duration,
	))
}

func sparkline(values []float64) string {
	if len(values) == 0 {
		return ""
	}

	chars := []rune{'▁', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	maxVal := values[0]
	for _, v := range values {
		if v > maxVal {
			maxVal = v
		}
	}

	if maxVal == 0 {
		return strings.Repeat(string(chars[0]), len(values))
	}

	var result strings.Builder
	for _, v := range values {
		idx := int((v / maxVal) * float64(len(chars)-1))
		if idx >= len(chars) {
			idx = len(chars) - 1
		}
		result.WriteRune(chars[idx])
	}

	return result.String()
}

func init() {
	rootCmd.AddCommand(bwCmd)

	bwCmd.Flags().DurationVarP(&bwInterval, "interval", "i", time.Second, "Refresh interval")
	bwCmd.Flags().IntVarP(&bwCount, "count", "c", 0, "Number of iterations (0 = unlimited)")
	bwCmd.Flags().StringVarP(&bwOutputFormat, "output", "o", "table", "Output format (table, json, csv)")
	bwCmd.Flags().BoolVar(&bwNoHeaders, "no-headers", false, "Omit headers")
}
