package main

import (
	"os"

	"github.com/marineverse/globe-nmea-server/internal/config"
	"github.com/marineverse/globe-nmea-server/internal/server"
	"github.com/marineverse/globe-nmea-server/pkg/logger"
)

func main() {
	log := logger.New()
	
	cfg, err := config.Parse()
	if err != nil {
		log.Printf("Failed to parse config: %v", err)
		os.Exit(1)
	}

	srv := server.New(cfg, log)

	// Start UDP broadcast in background if enabled
	if cfg.UDP {
		go func() {
			if err := srv.StartUDP(); err != nil {
				log.Printf("UDP error: %v", err)
			}
		}()
	}

	if err := srv.Start(); err != nil {
		log.Printf("Server error: %v", err)
		os.Exit(1)
	}
}