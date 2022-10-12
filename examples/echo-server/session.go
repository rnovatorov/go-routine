package main

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
)

type Session struct {
	logger *log.Logger
	conn   net.Conn
}

func NewSession(name string, logger *log.Logger, conn net.Conn) *Session {
	return &Session{
		logger: log.New(logger.Writer(), name+" ", logger.Flags()),
		conn:   conn,
	}
}

func (s *Session) Run(ctx context.Context) error {
	s.logger.Print("started")
	defer s.logger.Print("stopped")

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

		switch m := string(line); m {
		case "exit\n":
			return nil
		case "panic\n":
			var s []string
			s[42] = "oops"
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
