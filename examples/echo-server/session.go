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

type Session struct {
	*routine.Routine
	logger   *log.Logger
	conn     net.Conn
	stopOnce sync.Once
}

func StartNewSession(parent routine.Parent, logger *log.Logger, conn net.Conn) *Session {
	s := &Session{
		logger: logger,
		conn:   conn,
	}

	s.Routine = routine.Go(parent, s.run)

	return s
}

func (s *Session) Stop() error {
	s.stopOnce.Do(func() {
		if err := s.conn.Close(); err != nil {
			s.logger.Printf("failed to close session: %v", err)
		}
	})

	return s.Routine.Stop()
}

func (s *Session) run(ctx context.Context) error {
	s.logger.Print("session started")
	defer s.logger.Print("session stopped")

	r := bufio.NewReader(s.conn)
	w := bufio.NewWriter(s.conn)

	for {
		line, err := r.ReadBytes('\n')
		if err != nil {
			if errors.Is(err, net.ErrClosed) || errors.Is(err, io.EOF) {
				return nil
			}
			return fmt.Errorf("read: %w", err)
		}

		if v := line[0]; v == 'X' {
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
