package server

import (
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/marineverse/globe-nmea-server/internal/config"
	"github.com/marineverse/globe-nmea-server/internal/nmea"
	"github.com/marineverse/globe-nmea-server/pkg/logger"
)

type Server struct {
	config *config.Config
	logger *logger.Logger
	client *nmea.Client
	
	cachedMessage string
	cacheMutex    sync.RWMutex
	lastUpdate    time.Time
}

func New(cfg *config.Config, log *logger.Logger) *Server {
	s := &Server{
		config: cfg,
		logger: log,
		client: nmea.NewClient(cfg.Host),
	}
	
	// Start the background fetcher
	go s.backgroundFetch()
	
	return s
}

func (s *Server) backgroundFetch() {
	// Fetch initial message
	if err := s.updateCache(); err != nil {
		s.logger.Printf("Initial cache update failed: %v", err)
	}

	ticker := time.NewTicker(600 * time.Second) // 10 minutes
	defer ticker.Stop()

	for range ticker.C {
		if err := s.updateCache(); err != nil {
			s.logger.Printf("Cache update failed: %v", err)
		}
	}
}

func (s *Server) updateCache() error {
	nmea, err := s.client.FetchNMEASentence(s.config.BoatUUID)
	if err != nil {
		return fmt.Errorf("error fetching NMEA sentence: %v", err)
	}

	s.cacheMutex.Lock()
	s.cachedMessage = nmea
	s.lastUpdate = time.Now()
	s.cacheMutex.Unlock()

	s.logger.Printf("Cache updated with new NMEA data at %v", s.lastUpdate.Format(time.RFC3339))
	return nil
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
		
		// Refresh cache on new connection
		if err := s.updateCache(); err != nil {
			s.logger.Printf("Failed to refresh cache for new connection: %v", err)
			// Continue anyway with existing cache data
		}
		
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
	if err := s.sendCachedMessage(conn); err != nil {
		s.logger.Printf("Error sending initial message: %v", err)
		return
	}

	publishTicker := time.NewTicker(5 * time.Second) // Publish every 5 seconds
	defer publishTicker.Stop()

	done := make(chan struct{})
	defer close(done)

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

	for {
		select {
		case <-done:
			return
		case <-publishTicker.C:
			if err := s.sendCachedMessage(conn); err != nil {
				s.logger.Printf("Error in publish loop: %v", err)
				return
			}
		}
	}
}

func (s *Server) sendCachedMessage(conn net.Conn) error {
	s.cacheMutex.RLock()
	nmea := s.cachedMessage
	lastUpdate := s.lastUpdate
	s.cacheMutex.RUnlock()

	if nmea == "" {
		return fmt.Errorf("no cached NMEA data available")
	}

	s.logger.Printf("Sending to %s: %s (cached %s ago)", 
		conn.RemoteAddr(), 
		nmea,
		time.Since(lastUpdate).Round(time.Second),
	)
	
	if err := conn.SetWriteDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return fmt.Errorf("error setting write deadline: %v", err)
	}

	_, err := conn.Write([]byte(nmea + "\r\n"))
	if err != nil {
		return fmt.Errorf("error sending data: %v", err)
	}

	if err := conn.SetWriteDeadline(time.Time{}); err != nil {
		return fmt.Errorf("error resetting write deadline: %v", err)
	}

	return nil
}