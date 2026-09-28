// cmd/myapp/version.go
package main

import (
	"fmt"
	"runtime/debug"

	"github.com/spf13/cobra"
)

// Set via ldflags
var (
	version = "dev"
	commit  = "unknown"
	date    = "unknown"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintf(cmd.OutOrStdout(), "myapp %s (commit: %s, built: %s)\n", version, commit, date)

		if info, ok := debug.ReadBuildInfo(); ok {
			fmt.Fprintf(cmd.OutOrStdout(), "go: %s\n", info.GoVersion)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// Build with — for package main the -X path is main.<var>, not the import path;
// the linker silently ignores an unknown symbol and the binary ships "dev":
//
//   go build -ldflags "-X main.version=1.2.3 \
//     -X main.commit=$(git rev-parse --short HEAD) \
//     -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
//     -o bin/myapp ./cmd/myapp
