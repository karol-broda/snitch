package cmd

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/tabwriter"

	"github.com/charmbracelet/lipgloss"
	"github.com/spf13/cobra"

	"github.com/karol-broda/snitch/internal/collector"
	"github.com/karol-broda/snitch/internal/color"
	"github.com/karol-broda/snitch/internal/config"
	"github.com/karol-broda/snitch/internal/errutil"
	"github.com/karol-broda/snitch/internal/resolver"
	"github.com/tidwall/pretty"
	"golang.org/x/term"
)

// ls-specific flags
var (
	outputFormat  string
	noHeaders     bool
	showTimestamp bool
	sortBy        string
	fields        string
	colorMode     string
	plainOutput   bool
)

var lsCmd = &cobra.Command{
	Use:   "ls [filters...]",
	Short: "One-shot listing of connections",
	Long: `One-shot listing of connections.

Filters are specified in key=value format. For example:
  snitch ls proto=tcp state=established

Available filters:
  proto, state, pid, proc, lport, rport, user, laddr, raddr, contains, if, mark, namespace, inode, since
`,
	Run: func(cmd *cobra.Command, args []string) {
		runListCommand(outputFormat, args)
	},
}

func runListCommand(outputFormat string, args []string) {
	rt, err := NewRuntime(args, colorMode)
	if err != nil {
		log.Fatal(err)
	}

	// apply sorting
	if sortBy != "" {
		rt.SortConnections(collector.ParseSortOptions(sortBy))
	} else {
		rt.SortConnections(collector.SortOptions{
			Field:     collector.SortByLport,
			Direction: collector.SortAsc,
		})
	}

	selectedFields := []string{}
	if fields != "" {
		selectedFields = strings.Split(fields, ",")
	}

	renderList(rt.Connections, outputFormat, selectedFields)
}

func renderList(connections []collector.Connection, format string, selectedFields []string) {
	switch format {
	case "json":
		printJSON(connections)
	case "csv":
		printCSV(connections, !noHeaders, showTimestamp, selectedFields)
	case "table", "wide":
		if plainOutput {
			printPlainTable(connections, !noHeaders, showTimestamp, selectedFields)
		} else {
			printStyledTable(connections, !noHeaders, selectedFields)
		}
	default:
		log.Fatalf("Invalid output format: %s. Valid formats are: table, wide, json, csv", format)
	}
}


func getFieldMap(c collector.Connection) map[string]string {
	laddr := c.Laddr
	raddr := c.Raddr
	lport := strconv.Itoa(c.Lport)
	rport := strconv.Itoa(c.Rport)
	
	// apply address resolution
	if resolveAddrs {
		if resolvedLaddr := resolver.ResolveAddr(c.Laddr); resolvedLaddr != c.Laddr {
			laddr = resolvedLaddr
		}
		if resolvedRaddr := resolver.ResolveAddr(c.Raddr); resolvedRaddr != c.Raddr && c.Raddr != "*" && c.Raddr != "" {
			raddr = resolvedRaddr
		}
	}

	// apply port resolution
	if resolvePorts {
		if resolvedLport := resolver.ResolvePort(c.Lport, c.Proto); resolvedLport != strconv.Itoa(c.Lport) {
			lport = resolvedLport
		}
		if resolvedRport := resolver.ResolvePort(c.Rport, c.Proto); resolvedRport != strconv.Itoa(c.Rport) && c.Rport != 0 {
			rport = resolvedRport
		}
	}
	
	return map[string]string{
		"pid":       strconv.Itoa(c.PID),
		"process":   c.Process,
		"user":      c.User,
		"uid":       strconv.Itoa(c.UID),
		"proto":     c.Proto,
		"ipversion": c.IPVersion,
		"state":     c.State,
		"laddr":     laddr,
		"lport":     lport,
		"raddr":     raddr,
		"rport":     rport,
		"if":        c.Interface,
		"rx_bytes":  strconv.FormatInt(c.RxBytes, 10),
		"tx_bytes":  strconv.FormatInt(c.TxBytes, 10),
		"rtt_ms":    strconv.FormatFloat(c.RttMs, 'f', 1, 64),
		"mark":      c.Mark,
		"namespace": c.Namespace,
		"inode":     strconv.FormatInt(c.Inode, 10),
		"ts":        c.TS.Format("2006-01-02T15:04:05.000Z07:00"),
	}
}

func printJSON(conns []collector.Connection) {
	jsonOutput, err := json.MarshalIndent(conns, "", "  ")
	if err != nil {
		log.Fatalf("Error marshaling to JSON: %v", err)
	}

	if color.IsColorDisabled() {
		fmt.Println(string(jsonOutput))
	} else {
		colored := pretty.Color(jsonOutput, nil)
		fmt.Println(string(colored))
	}
}

func printCSV(conns []collector.Connection, headers bool, timestamp bool, selectedFields []string) {
	writer := csv.NewWriter(os.Stdout)
	defer writer.Flush()

	if len(selectedFields) == 0 {
		selectedFields = []string{"pid", "process", "user", "uid", "proto", "state", "laddr", "lport", "raddr", "rport"}
		if timestamp {
			selectedFields = append([]string{"ts"}, selectedFields...)
		}
	}

	if headers {
		headerRow := []string{}
		for _, field := range selectedFields {
			headerRow = append(headerRow, strings.ToUpper(field))
		}
		_ = writer.Write(headerRow)
	}

	for _, conn := range conns {
		fieldMap := getFieldMap(conn)
		row := []string{}
		for _, field := range selectedFields {
			row = append(row, fieldMap[field])
		}
		_ = writer.Write(row)
	}
}

func printPlainTable(conns []collector.Connection, headers bool, timestamp bool, selectedFields []string) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	defer errutil.Flush(w)

	if len(selectedFields) == 0 {
		selectedFields = []string{"pid", "process", "user", "proto", "state", "laddr", "lport", "raddr", "rport"}
		if timestamp {
			selectedFields = append([]string{"ts"}, selectedFields...)
		}
	}

	if headers {
		headerRow := []string{}
		for _, field := range selectedFields {
			headerRow = append(headerRow, strings.ToUpper(field))
		}
		errutil.Ignore(fmt.Fprintln(w, strings.Join(headerRow, "\t")))
	}

	for _, conn := range conns {
		fieldMap := getFieldMap(conn)
		row := []string{}
		for _, field := range selectedFields {
			row = append(row, fieldMap[field])
		}
		errutil.Ignore(fmt.Fprintln(w, strings.Join(row, "\t")))
	}
}

func printStyledTable(conns []collector.Connection, headers bool, selectedFields []string) {
	if len(selectedFields) == 0 {
		selectedFields = []string{"process", "pid", "proto", "state", "laddr", "lport", "raddr", "rport"}
	}

	// calculate column widths
	widths := make(map[string]int)
	for _, f := range selectedFields {
		widths[f] = len(strings.ToUpper(f))
	}

	for _, conn := range conns {
		fm := getFieldMap(conn)
		for _, f := range selectedFields {
			if len(fm[f]) > widths[f] {
				widths[f] = len(fm[f])
			}
		}
	}

	// cap and pad widths
	for f := range widths {
		if widths[f] > 25 {
			widths[f] = 25
		}
		widths[f] += 2 // padding
	}

	// build output
	var output strings.Builder

	// styles
	borderStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	processStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("15"))
	faintStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("245"))

	// build top border
	output.WriteString("\n")
	output.WriteString(borderStyle.Render("  ╭"))
	for i, f := range selectedFields {
		if i > 0 {
			output.WriteString(borderStyle.Render("┬"))
		}
		output.WriteString(borderStyle.Render(strings.Repeat("─", widths[f])))
	}
	output.WriteString(borderStyle.Render("╮"))
	output.WriteString("\n")

	// header row
	if headers {
		output.WriteString(borderStyle.Render("  │"))
		for i, f := range selectedFields {
			if i > 0 {
				output.WriteString(borderStyle.Render("│"))
			}
			cell := fmt.Sprintf(" %-*s", widths[f]-1, strings.ToUpper(f))
			output.WriteString(headerStyle.Render(cell))
		}
		output.WriteString(borderStyle.Render("│"))
		output.WriteString("\n")

		// header separator
		output.WriteString(borderStyle.Render("  ├"))
		for i, f := range selectedFields {
			if i > 0 {
				output.WriteString(borderStyle.Render("┼"))
			}
			output.WriteString(borderStyle.Render(strings.Repeat("─", widths[f])))
		}
		output.WriteString(borderStyle.Render("┤"))
		output.WriteString("\n")
	}

	// data rows
	for _, conn := range conns {
		fm := getFieldMap(conn)
		output.WriteString(borderStyle.Render("  │"))
		for i, f := range selectedFields {
			if i > 0 {
				output.WriteString(borderStyle.Render("│"))
			}
			val := fm[f]
			maxW := widths[f] - 2
			if len(val) > maxW {
				val = val[:maxW-1] + "…"
			}
			cell := fmt.Sprintf(" %-*s ", maxW, val)

			switch f {
			case "proto":
				c := lipgloss.Color("37") // cyan
				if strings.Contains(fm["proto"], "udp") {
					c = lipgloss.Color("135") // purple
				}
				output.WriteString(lipgloss.NewStyle().Foreground(c).Render(cell))
			case "state":
				c := lipgloss.Color("245") // gray
				switch strings.ToUpper(fm["state"]) {
				case "LISTEN":
					c = lipgloss.Color("35") // green
				case "ESTABLISHED":
					c = lipgloss.Color("33") // blue
				case "TIME_WAIT", "CLOSE_WAIT":
					c = lipgloss.Color("178") // yellow
				}
				output.WriteString(lipgloss.NewStyle().Foreground(c).Render(cell))
			case "process":
				output.WriteString(processStyle.Render(cell))
			default:
				output.WriteString(cell)
			}
		}
		output.WriteString(borderStyle.Render("│"))
		output.WriteString("\n")
	}

	// bottom border
	output.WriteString(borderStyle.Render("  ╰"))
	for i, f := range selectedFields {
		if i > 0 {
			output.WriteString(borderStyle.Render("┴"))
		}
		output.WriteString(borderStyle.Render(strings.Repeat("─", widths[f])))
	}
	output.WriteString(borderStyle.Render("╯"))
	output.WriteString("\n")

	// summary
	output.WriteString(faintStyle.Render(fmt.Sprintf("  %d connections\n", len(conns))))
	output.WriteString("\n")

	// output with pager if needed
	printWithPager(output.String())
}

func printWithPager(content string) {
	lines := strings.Count(content, "\n")

	// check if stdout is a terminal and content is long
	if term.IsTerminal(int(os.Stdout.Fd())) {
		_, height, err := term.GetSize(int(os.Stdout.Fd()))
		if err == nil && lines > height-2 {
			// use pager
			pager := os.Getenv("PAGER")
			if pager == "" {
				pager = "less"
			}

			cmd := exec.Command(pager, "-R") // -R for color support
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			stdin, err := cmd.StdinPipe()
			if err != nil {
				fmt.Print(content)
				return
			}

			if err := cmd.Start(); err != nil {
				fmt.Print(content)
				return
			}

			_, _ = io.WriteString(stdin, content)
			_ = stdin.Close()
			_ = cmd.Wait()
			return
		}
	}

	fmt.Print(content)
}

func init() {
	rootCmd.AddCommand(lsCmd)

	cfg := config.Get()

	// ls-specific flags
	lsCmd.Flags().StringVarP(&outputFormat, "output", "o", cfg.Defaults.OutputFormat, "Output format (table, wide, json, csv)")
	lsCmd.Flags().BoolVar(&noHeaders, "no-headers", cfg.Defaults.NoHeaders, "Omit headers for table/csv output")
	lsCmd.Flags().BoolVar(&showTimestamp, "ts", false, "Include timestamp in output")
	lsCmd.Flags().StringVarP(&sortBy, "sort", "s", cfg.Defaults.SortBy, "Sort by column (e.g., pid:desc)")
	lsCmd.Flags().StringVarP(&fields, "fields", "f", strings.Join(cfg.Defaults.Fields, ","), "Comma-separated list of fields to show")
	lsCmd.Flags().StringVar(&colorMode, "color", cfg.Defaults.Color, "Color mode (auto, always, never)")
	lsCmd.Flags().BoolVarP(&plainOutput, "plain", "p", false, "Plain output (parsable, no styling)")

	// shared flags
	addFilterFlags(lsCmd)
	addResolutionFlags(lsCmd)
}