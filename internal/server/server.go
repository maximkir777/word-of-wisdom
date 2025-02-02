package server

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"

	"github.com/maximkir777/word-of-wisdom/pkg/protocol"
)

// Handler defines the interface for processing incoming messages.
type Handler interface {
	ProcessRequest(ctx context.Context, msgStr string, clientInfo string) (*protocol.Message, error)
}

// Server represents a TCP server.
type Server struct {
	listenAddr string
	ln         net.Listener
	handler    Handler
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	logger     *slog.Logger
}

// NewServer creates a new TCP server instance.
func NewServer(listenAddr string, handler Handler, logger *slog.Logger) *Server {
	if logger == nil {
		logger = slog.Default()
	}
	return &Server{
		listenAddr: listenAddr,
		handler:    handler,
		logger:     logger,
	}
}

// Start initializes and starts the TCP server.
func (s *Server) Start() error {
	ln, err := net.Listen("tcp", s.listenAddr)
	if err != nil {
		return fmt.Errorf("failed to start server: %w", err)
	}
	s.ln = ln
	s.ctx, s.cancel = context.WithCancel(context.Background())

	s.logger.Info("Server started", "address", s.listenAddr)

	s.wg.Add(1)
	go s.acceptLoop()

	return nil
}

// acceptLoop continuously accepts new connections.
func (s *Server) acceptLoop() {
	defer s.wg.Done()

	for {
		conn, err := s.ln.Accept()
		if errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			s.logger.Warn("Error accepting connection", "error", err)
			continue
		}

		select {
		case <-s.ctx.Done():
			conn.Close()
			return
		default:
			s.wg.Add(1)
			go s.handleConn(conn)
		}
	}
}

// handleConn processes each individual TCP connection.
func (s *Server) handleConn(conn net.Conn) {
	defer func() {
		conn.Close()
		s.wg.Done()
	}()

	clientAddr := conn.RemoteAddr().String()
	s.logger.Info("New connection", "client", clientAddr)

	reader := bufio.NewReader(conn)
	for {
		select {
		case <-s.ctx.Done():
			return
		default:
			msgStr, err := reader.ReadString('\n')
			if err != nil {
				if errors.Is(err, io.EOF) {
					s.logger.Info("Connection closed by client", "client", clientAddr)
				} else {
					s.logger.Warn("Error reading from connection", "client", clientAddr, "error", err)
				}
				return
			}

			msg, err := s.handler.ProcessRequest(s.ctx, msgStr, clientAddr)
			if err != nil {
				s.logger.Warn("Processing error", "client", clientAddr, "error", err)
				errMsg := protocol.Message{
					Header:  protocol.ResponseResource,
					Payload: "Error: " + err.Error(),
				}
				_ = s.sendMessage(errMsg, conn)
				return
			}

			if msg != nil {
				if err := s.sendMessage(*msg, conn); err != nil {
					s.logger.Warn("Send error", "client", clientAddr, "error", err)
					return
				}
			}
		}
	}
}

// sendMessage serializes and sends a message via the connection.
func (s *Server) sendMessage(msg protocol.Message, conn net.Conn) error {
	msgStr := fmt.Sprintf("%s\n", msg.Stringify())
	_, err := conn.Write([]byte(msgStr))
	return err
}

// Stop performs a graceful shutdown of the server.
func (s *Server) Stop() {
	s.logger.Info("Initiating graceful shutdown")
	s.cancel()

	if err := s.ln.Close(); err != nil {
		s.logger.Warn("Error closing listener", "error", err)
	}

	s.wg.Wait()
	s.logger.Info("Server stopped")
}
