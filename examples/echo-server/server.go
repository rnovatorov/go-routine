package main

import (
	"context"
	"log"
	"net"

	"github.com/rnovatorov/go-routine"
)

type Server struct {
	*routine.Routine
	logger        *log.Logger
	listenAddress string
}

func StartNewServer(parent routine.Parent, logger *log.Logger, listenAddress string) *Server {
	s := &Server{
		logger:        logger,
		listenAddress: listenAddress,
	}

	s.Routine = routine.Go(parent, s.run)

	return s
}

func (s *Server) run(ctx context.Context) error {
	s.logger.Print("server started")
	defer s.logger.Print("server stopped")

	var listenConfig net.ListenConfig
	listener, err := listenConfig.Listen(ctx, "tcp", s.listenAddress)
	if err != nil {
		return err
	}

	acceptor := StartNewAcceptor(s.Routine, s.logger, listener)
	defer s.stopAcceptor(acceptor)

	handler := StartNewHandler(s.Routine, s.logger, acceptor)
	defer s.stopHandler(handler)

	<-ctx.Done()

	return nil
}

func (s *Server) stopAcceptor(acceptor *Acceptor) {
	if err := acceptor.Stop(); err != nil {
		s.logger.Printf("failed to stop acceptor: %v", err)
	}
}

func (s *Server) stopHandler(handler *Handler) {
	if err := handler.Stop(); err != nil {
		s.logger.Printf("failed to stop handler: %v", err)
	}
}
