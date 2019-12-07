package cmd

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var (
	buildCode = ""
	buildTimestamp = ""
	coreVersion = ""
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Run: func(cmd *cobra.Command, args []string) {
		st, _ := strconv.ParseInt(buildTimestamp, 10, 64)
		fmt.Printf("version: %s\n", coreVersion)
		fmt.Printf("build: %s\n", buildCode)
		fmt.Printf("release date: %s\n", time.Unix(st, 0).Format(time.RFC822))
		fmt.Printf("go: %s\n", runtime.Version())
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
