package main

import (
	"context"
	"errors"
	"log"
	"net"
	"sync"

	"github.com/rnovatorov/go-routine"
)

type Acceptor struct {
	*routine.Routine
	logger   *log.Logger
	listener net.Listener
	conns    chan net.Conn
	stopOnce sync.Once
}

func StartNewAcceptor(
	parent routine.Parent, logger *log.Logger, listener net.Listener,
) *Acceptor {
	a := &Acceptor{
		logger:   logger,
		listener: listener,
		conns:    make(chan net.Conn),
	}

	a.Routine = routine.Go(parent, a.run)

	return a
}

func (a *Acceptor) Stop() error {
	a.stopOnce.Do(func() {
		if err := a.listener.Close(); err != nil {
			a.logger.Printf("failed to close listener: %v", err)
		}
	})

	return a.Routine.Stop()
}

func (a *Acceptor) Conns() <-chan net.Conn {
	return a.conns
}

func (a *Acceptor) run(ctx context.Context) error {
	a.logger.Print("acceptor started")
	defer a.logger.Print("acceptor stopped")

	for {
		conn, err := a.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return err
		}

		select {
		case <-ctx.Done():
			a.closeConn(conn)
			return nil
		case a.conns <- conn:
		}
	}
}

func (a *Acceptor) closeConn(conn net.Conn) {
	if err := conn.Close(); err != nil {
		a.logger.Printf("failed to close conn: %v", err)
	}
}
