package cmd

import (
	"fmt"
	"runtime"

	"github.com/fatih/color"
	"github.com/spf13/cobra"

	"github.com/karol-broda/snitch/internal/errutil"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version/build info",
	Run: func(cmd *cobra.Command, args []string) {
		bold := color.New(color.Bold)
		cyan := color.New(color.FgCyan)
		faint := color.New(color.Faint)

		errutil.Print(bold, "snitch ")
		errutil.Println(cyan, Version)
		fmt.Println()

		errutil.Print(faint, "  commit  ")
		fmt.Println(Commit)

		errutil.Print(faint, "  built   ")
		fmt.Println(Date)

		errutil.Print(faint, "  go      ")
		fmt.Println(runtime.Version())

		errutil.Print(faint, "  os      ")
		fmt.Printf("%s/%s\n", runtime.GOOS, runtime.GOARCH)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
