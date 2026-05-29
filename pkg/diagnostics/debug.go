package diagnostics

import (
	"fmt"
	"os"
)

var verbose bool

func SetVerbose(v bool) {
	verbose = v
}

func IsVerbose() bool {
	return verbose
}

func Debug(fn func()) {
	if os.Getenv("DEBUG") != "" {
		fn()
	}
}

func Info(format string, args ...any) {
	if verbose {
		fmt.Printf("INFO -   "+format+"\n", args...)
	}
}
