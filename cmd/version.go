package cmd

import (
	"fmt"
	"runtime"
	"time"

	"github.com/spf13/cobra"
)

var (
	buildCode      = ""
	buildTimestamp = ""
	coreVersion    = ""
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		rd, _ := time.Parse(time.RFC3339, buildTimestamp)
		fmt.Printf("version: %s\n", coreVersion)
		fmt.Printf("build: %s\n", buildCode)
		fmt.Printf("release date: %s\n", rd.Format(time.RFC822))
		fmt.Printf("go: %s\n", runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
