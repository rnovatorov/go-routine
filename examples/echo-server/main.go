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

	logger := log.New(os.Stderr, "", log.Ldate|log.Ltime|log.Lmsgprefix)

	ctx = routine.NewPanicHookContext(ctx, func(v interface{}) {
		logger.Println(":(")
	})

	listenAddress := os.Getenv("LISTEN_ADDRESS")

	server := NewServer(logger, listenAddress)
	return server.Run(ctx)
}
