package main

import (
	"fmt"
	"os"
)

// runStatus is a stub until the status checks (Phase 8) are implemented.
func runStatus(args []string) int {
	fmt.Fprintln(os.Stderr, "komit: status is not implemented yet")
	return 1
}
