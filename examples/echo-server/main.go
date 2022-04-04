package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rnovatorov/go-routine"
)

func main() {
	if err := run(context.Background()); err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", os.Args[0], err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	ctx, cancel := signal.NotifyContext(ctx, syscall.SIGINT)
	defer cancel()

	mainRoutine := routine.Main(ctx)
	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmsgprefix)
	listenAddress := os.Getenv("LISTEN_ADDRESS")

	server := StartNewServer(mainRoutine, logger, listenAddress)
	return server.Wait()
}
