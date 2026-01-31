package main

import (
	"fmt"
	"infomunge/internal/cli"
	"os"
)

func main() {
	app := cli.NewApp()

	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

}
