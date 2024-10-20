package main

import (
	"context"
	"errors"
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

	ctx = routine.WithMiddleware(ctx, routine.NewRecoverMiddleware(func(v any) {
		logger.Panicf("recover middleware: %v", v)
	}))

	listenAddress := os.Getenv("LISTEN_ADDRESS")
	if listenAddress == "" {
		return errListenAddressEmpty
	}

	return serve(ctx, logger, listenAddress)
}

var errListenAddressEmpty = errors.New("listen address empty")
