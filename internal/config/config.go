package config

import (
	"flag"
	"fmt"
	"net/url"
)

type Config struct {
	BoatUUID string
	Port     int
	Host     string
}

func Parse() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Host, "host", "https://api.marineverse.com", "Host URL of the API")
	flag.IntVar(&cfg.Port, "port", 3006, "Port number to listen on")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		return nil, fmt.Errorf("boat UUID is required")
	}
	cfg.BoatUUID = args[0]

	if _, err := url.Parse(cfg.Host); err != nil {
		return nil, fmt.Errorf("invalid host URL: %v", err)
	}

	return cfg, nil
}