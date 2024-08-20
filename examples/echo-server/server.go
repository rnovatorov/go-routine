package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/rnovatorov/go-routine"
)

type Server struct {
	*routine.Routine
	logger    *log.Logger
	listener  net.Listener
	idCounter int
	mu        sync.Mutex
	sessions  map[int]*routine.Routine
}

func StartServer(
	ctx context.Context, logger *log.Logger, listenAddress string,
) (*Server, error) {
	listener, err := (&net.ListenConfig{}).Listen(ctx, "tcp", listenAddress)
	if err != nil {
		return nil, err
	}

	s := &Server{
		logger:   childLogger(logger, fmt.Sprintf("server[%s]", listenAddress)),
		listener: listener,
		sessions: make(map[int]*routine.Routine),
	}
	s.Routine = routine.Go(ctx, s.run)

	return s, nil
}

func (s *Server) run(ctx context.Context) error {
	s.logger.Print("started")
	defer s.logger.Print("stopped")

	listenerCloser := routine.Go(ctx, func(ctx context.Context) error {
		<-ctx.Done()
		if err := s.listener.Close(); err != nil {
			s.logger.Printf("failed to close listener: %v", err)
		}
		return nil
	})
	defer listenerCloser.Stop()

	defer s.stopAllSessions()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			return fmt.Errorf("accept conn: %w", err)
		}
		s.startSession(ctx, conn)
	}
}

func (s *Server) startSession(ctx context.Context, conn net.Conn) {
	id := s.newSessionID()

	logger := childLogger(s.logger, fmt.Sprintf(
		"session[%d][%s]", id, conn.RemoteAddr()))

	s.mu.Lock()
	defer s.mu.Unlock()

	s.sessions[id] = routine.Go(ctx, func(ctx context.Context) error {
		defer func() {
			s.mu.Lock()
			defer s.mu.Unlock()
			delete(s.sessions, id)
		}()

		logger.Print("started")
		defer logger.Print("stopped")

		connCloser := routine.Go(ctx, func(ctx context.Context) error {
			<-ctx.Done()
			if err := conn.Close(); err != nil {
				logger.Printf("failed to close conn: %v", err)
			}
			return nil
		})
		defer connCloser.Stop()

		if err := s.echo(conn); err != nil {
			logger.Printf("echo failed: %v", err)
		}
		return nil
	})
}

func (s *Server) stopAllSessions() {
	sessions := make(map[int]*routine.Routine)
	s.mu.Lock()
	for id, session := range s.sessions {
		sessions[id] = session
	}
	s.mu.Unlock()

	for id, session := range sessions {
		if err := session.Stop(); err != nil {
			s.logger.Printf("failed to stop session[%d]: %v", id, err)
		}
	}
}

func (s *Server) echo(conn net.Conn) error {
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

func (s *Server) newSessionID() int {
	s.idCounter++
	return s.idCounter
}

func childLogger(parent *log.Logger, prefix string) *log.Logger {
	prefix = fmt.Sprintf("%s%s ", parent.Prefix(), prefix)
	return log.New(parent.Writer(), prefix, parent.Flags())
}
