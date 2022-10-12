package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"

	"github.com/rnovatorov/go-routine"
)

type Server struct {
	logger        *log.Logger
	listenAddress string
}

func NewServer(logger *log.Logger, listenAddress string) *Server {
	name := fmt.Sprintf("server[%s]", listenAddress)

	return &Server{
		logger:        log.New(logger.Writer(), name+" ", logger.Flags()),
		listenAddress: listenAddress,
	}
}

func (s *Server) Run(ctx context.Context) error {
	rg := routine.NewGroup(ctx)

	listenerChan := s.listen(rg)
	connChan := s.acceptConns(rg, listenerChan)
	s.handleConns(rg, connChan)

	s.logger.Print("started")
	defer s.logger.Print("stopped")

	return rg.Wait()
}

func (s *Server) listen(rg *routine.Group) <-chan net.Listener {
	listenerChan := make(chan net.Listener, 1)

	rg.Go("listen", func(ctx context.Context) error {
		var config net.ListenConfig

		listener, err := config.Listen(ctx, "tcp", s.listenAddress)
		if err != nil {
			return err
		}

		rg.Go("close listener", func(ctx context.Context) error {
			<-ctx.Done()
			return listener.Close()
		})

		listenerChan <- listener

		return nil
	})

	return listenerChan
}

func (s *Server) acceptConns(rg *routine.Group, listenerChan <-chan net.Listener) <-chan net.Conn {
	connChan := make(chan net.Conn)

	rg.Go("accept conns", func(ctx context.Context) error {
		var listener net.Listener

		select {
		case <-ctx.Done():
			return nil
		case listener = <-listenerChan:
		}

		for {
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return nil
				}
				return err
			}

			select {
			case <-ctx.Done():
				return conn.Close()
			case connChan <- conn:
			}
		}
	})

	return connChan
}

func (s *Server) handleConns(rg *routine.Group, connChan <-chan net.Conn) {
	rg.Go("handle conns", func(ctx context.Context) error {
		for {
			select {
			case <-ctx.Done():
				return nil
			case conn := <-connChan:
				s.startNewSession(rg, conn)
			}
		}
	})
}

func (s *Server) startNewSession(rg *routine.Group, conn net.Conn) {
	name := fmt.Sprintf("session[%s->%s]", conn.LocalAddr(), conn.RemoteAddr())
	session := NewSession(name, s.logger, conn)
	stopped := make(chan struct{})

	rg.Go(name, func(ctx context.Context) error {
		defer close(stopped)
		return session.Run(ctx)
	})

	rg.Go(name+" conn closer", func(ctx context.Context) error {
		select {
		case <-ctx.Done():
		case <-stopped:
		}
		return conn.Close()
	})
}
