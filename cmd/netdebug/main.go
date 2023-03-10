package main

import (
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/ripta/netdebug/pkg/app"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "last resort logger: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cmd, cleanup := app.New()
	defer cleanup()
	return cmd.Execute()
}
