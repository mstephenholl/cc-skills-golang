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

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "myapp %s (commit: %s, built: %s)\n", resolvedVersion(), commit, date)

			if info, ok := debug.ReadBuildInfo(); ok {
				fmt.Fprintf(cmd.OutOrStdout(), "go: %s\n", info.GoVersion)
			}
		},
	}
}

// resolvedVersion falls back to the module version in the build info — `go install module@version`
// runs no ldflags, so without it those binaries would report "dev".
func resolvedVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

// Build with — for package main the -X path is main.<var>, not the import path;
// the linker silently ignores an unknown symbol and the binary ships "dev":
//
//   go build -ldflags "-X main.version=1.2.3 \
//     -X main.commit=$(git rev-parse --short HEAD) \
//     -X main.date=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
//     -o bin/myapp ./cmd/myapp
