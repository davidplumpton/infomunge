package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"infomunge/internal/playground"
)

func main() {
	listenAddress := flag.String("listen", "127.0.0.1:8081", "address for the standalone playground server")
	rootDirectory := flag.String("root", "docs/playground", "directory containing the standalone playground assets")
	flag.Parse()

	if !playground.DirectoryExists(*rootDirectory) {
		fmt.Fprintf(os.Stderr, "playground directory %q does not exist; run this command from the repository root\n", *rootDirectory)
		os.Exit(1)
	}

	server := &http.Server{
		Addr:              *listenAddress,
		Handler:           playground.Handler(*rootDirectory),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("serving standalone playground at http://%s", *listenAddress)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
