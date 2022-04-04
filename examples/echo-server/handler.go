package main

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/rnovatorov/go-routine"
)

type Handler struct {
	*routine.Routine
	logger   *log.Logger
	acceptor *Acceptor
	sessions []*Session
}

func StartNewHandler(
	parent routine.Parent, logger *log.Logger, acceptor *Acceptor,
) *Handler {
	h := &Handler{
		logger:   logger,
		acceptor: acceptor,
	}

	h.Routine = routine.Go(parent, h.run)

	return h
}

func (h *Handler) run(ctx context.Context) error {
	h.logger.Print("handler started")
	defer h.logger.Print("handler stopped")

	for {
		select {
		case <-ctx.Done():
			h.stopSessions()
			return nil
		case conn := <-h.acceptor.Conns():
			s := h.startNewSession(conn)
			h.sessions = append(h.sessions, s)
		}
	}
}

func (h *Handler) startNewSession(conn net.Conn) *Session {
	prefix := fmt.Sprintf("[%s->%s] ",
		conn.LocalAddr(),
		conn.RemoteAddr())

	logger := log.New(h.logger.Writer(), prefix, h.logger.Flags())

	return StartNewSession(h.Routine, logger, conn)
}

func (h *Handler) stopSessions() {
	for _, s := range h.sessions {
		if err := s.Stop(); err != nil {
			h.logger.Printf("failed to stop session: %v", err)
		}
	}
}
