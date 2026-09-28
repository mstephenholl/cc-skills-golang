package main

import (
	"errors"
	"fmt"
	"os"
)

// Pattern for mapping errors to exit codes.
func mainWithExitCodes() {
	if err := Execute(); err != nil {
		// The root sets SilenceErrors, so cobra printed nothing — print the error once, here.
		fmt.Fprintln(os.Stderr, "Error:", err)
		var exitErr *ExitError
		if errors.As(err, &exitErr) {
			os.Exit(exitErr.Code)
		}
		os.Exit(1)
	}
}

type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func (e *ExitError) Unwrap() error { return e.Err }
