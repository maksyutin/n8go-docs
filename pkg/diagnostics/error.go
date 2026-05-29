package diagnostics

import (
	"fmt"
	"log"
	"os"
)

func HandleError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}

func PrintError(err error, msg string) {
	if err != nil {
		log.Printf("error: %s: %s\n", msg, err.Error())
	}
}
