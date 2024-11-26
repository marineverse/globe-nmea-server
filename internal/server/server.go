// internal/server/server.go
package server

import (
	"fmt"
	"net"
	"time"

	"github.com/marineverse/globe-nmea-server/internal/config"
	"github.com/marineverse/globe-nmea-server/internal/nmea"
	"github.com/marineverse/globe-nmea-server/pkg/logger"
)

type Server struct {
	config *config.Config
	logger *logger.Logger
	client *nmea.Client
}

func New(cfg *config.Config, log *logger.Logger) *Server {
	return &Server{
		config: cfg,
		logger: log,
		client: nmea.NewClient(cfg.Host),
	}
}

func (s *Server) Start() error {
	listener, err := net.Listen("tcp", fmt.Sprintf("0.0.0.0:%d", s.config.Port))
	if err != nil {
		return fmt.Errorf("failed to start server: %v", err)
	}
	defer listener.Close()

	s.logger.Printf("Server listening on port %d...", s.config.Port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			s.logger.Printf("Error accepting connection: %v", err)
			continue
		}

		s.logger.Printf("Connected to %s", conn.RemoteAddr())
		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.logger.Printf("Connection closed for %s", conn.RemoteAddr())
	}()

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		tcpConn.SetKeepAlive(true)
		tcpConn.SetKeepAlivePeriod(30 * time.Second)
	}

	// Send initial message immediately
	if err := s.sendNMEAMessage(conn); err != nil {
		s.logger.Printf("Error sending initial message: %v", err)
		return
	}

	ticker := time.NewTicker(600 * time.Second) // don't call more often then every 10min
	defer ticker.Stop()

	// Channel to handle connection termination
	done := make(chan struct{})
	defer close(done)

	// Start a goroutine to check for connection errors
	go func() {
		buffer := make([]byte, 1)
		for {
			select {
			case <-done:
				return
			default:
				if _, err := conn.Read(buffer); err != nil {
					s.logger.Printf("Connection monitor detected error for %s: %v", conn.RemoteAddr(), err)
					close(done)
					return
				}
			}
		}
	}()

	// Main loop for sending NMEA sentences
	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			if err := s.sendNMEAMessage(conn); err != nil {
				s.logger.Printf("Error in ticker loop: %v", err)
				return
			}
		}
	}
}

// Extract message sending logic into its own method to avoid code duplication
func (s *Server) sendNMEAMessage(conn net.Conn) error {
	nmea, err := s.client.FetchNMEASentence(s.config.BoatUUID)
	if err != nil {
		return fmt.Errorf("error fetching NMEA sentence: %v", err)
	}

	s.logger.Printf("Sending to %s: %s", conn.RemoteAddr(), nmea)
	
	// Set write deadline to prevent blocking forever
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("error setting write deadline: %v", err)
	}

	// Send the NMEA sentence with CRLF line ending
	_, err = conn.Write([]byte(nmea + "\r\n"))
	if err != nil {
		return fmt.Errorf("error sending data: %v", err)
	}

	// Reset write deadline
	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		return fmt.Errorf("error resetting write deadline: %v", err)
	}

	return nil
}