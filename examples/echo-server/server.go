package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/rnovatorov/go-routine"
)

type server struct {
	routines      *routine.Group
	logger        *log.Logger
	listenAddress string
	listenerReady chan struct{}
	listener      net.Listener
}

func serve(
	ctx context.Context, logger *log.Logger, listenAddress string,
) error {
	s := &server{
		routines:      routine.NewGroup(ctx),
		logger:        childLogger(logger, fmt.Sprintf("server[%s]", listenAddress)),
		listenAddress: listenAddress,
		listenerReady: make(chan struct{}),
		listener:      nil,
	}

	s.routines.Go(s.listen)
	s.routines.Go(s.acceptConns)

	return s.routines.Wait()
}

func (s *server) listen(ctx context.Context) error {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", s.listenAddress)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	defer func() {
		if err := s.listener.Close(); err != nil {
			s.logger.Printf("failed to close listener: %v", err)
		}
	}()

	s.listener = listener
	close(s.listenerReady)

	<-ctx.Done()

	return nil
}

func (s *server) acceptConns(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return nil
	case <-s.listenerReady:
	}

	s.logger.Print("started accepting conns")
	defer s.logger.Print("stopped accepting conns")

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("accept conn: %w", err)
		}
		s.startSession(conn)
	}
}

func (s *server) startSession(conn net.Conn) {
	logger := childLogger(s.logger, fmt.Sprintf("session[%s]", conn.RemoteAddr()))

	session := s.routines.Go(func(ctx context.Context) error {
		logger.Print("session started")
		defer logger.Print("session stopped")

		if err := s.echo(conn); err != nil {
			return fmt.Errorf("echo: %w", err)
		}
		return nil
	})

	s.routines.Go(func(ctx context.Context) error {
		select {
		case <-session.Stopped():
		case <-ctx.Done():
		}
		if err := conn.Close(); err != nil {
			logger.Printf("failed to close conn: %v", err)
		}
		return nil
	})
}

func (s *server) echo(conn net.Conn) error {
	r := bufio.NewReader(conn)
	w := bufio.NewWriter(conn)

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read: %w", err)
		}

		switch m := string(line); m {
		case "exit\n":
			return nil
		case "panic\n":
			panic("oops")
		case "error\n":
			return errors.New("oops")
		}

		if _, err := w.Write(line); err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("write: %w", err)
		}

		if err := w.Flush(); err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("flush: %w", err)
		}
	}
}

func childLogger(parent *log.Logger, prefix string) *log.Logger {
	prefix = fmt.Sprintf("%s%s ", parent.Prefix(), prefix)
	return log.New(parent.Writer(), prefix, parent.Flags())
}
