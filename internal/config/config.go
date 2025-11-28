package config

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	BoatUUID string
	Port     int
	Host     string
	UDP      bool
	UDPPort  int
}

func Parse() (*Config, error) {
	cfg := &Config{}

	flag.StringVar(&cfg.Host, "host", "https://api.marineverse.com", "Host URL of the API")
	flag.IntVar(&cfg.Port, "port", 3006, "TCP port number to listen on")
	flag.BoolVar(&cfg.UDP, "udp", false, "Also broadcast via UDP")
	flag.IntVar(&cfg.UDPPort, "udp-port", 10110, "UDP broadcast port")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Print("Enter your Boat UUID (from https://www.marineverse.com/globe/boats → API tab): ")
		reader := bufio.NewReader(os.Stdin)
		uuid, err := reader.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read UUID: %v", err)
		}
		uuid = strings.TrimSpace(uuid)
		if uuid == "" {
			return nil, fmt.Errorf("boat UUID is required")
		}
		cfg.BoatUUID = uuid
	} else {
		cfg.BoatUUID = args[0]
	}

	if _, err := url.Parse(cfg.Host); err != nil {
		return nil, fmt.Errorf("invalid host URL: %v", err)
	}

	return cfg, nil
}