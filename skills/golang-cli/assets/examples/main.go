// cmd/myapp/main.go
package main

import (
	"fmt"
	"os"
)

func main() {
	if err := Execute(); err != nil {
		// The root sets SilenceErrors, so cobra printed nothing — print the error once, here.
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
