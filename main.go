package main

import (
	"os"

	"github.com/chinnareddy578/kubectl-login/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

