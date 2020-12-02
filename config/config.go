package config

import (
	"flag"
	"os"
	"sync"
	"time"
)

type Config struct {
	// Port to listen on
	Port int
	// Maximum number of concurrent requests
	MaxInRequests int
	// Maximum number of urls per request
	MaxUrls int
	// Maximum number of outgoing requests
	MaxOutRequests int
	// Outgoing request timeout
	Timeout time.Duration
}

var (
	cfg  Config
	help bool
	o    sync.Once
)

func init() {
	help := false
	flag.IntVar(&cfg.Port, "port", 3000, "Port to listen")
	flag.IntVar(&cfg.MaxInRequests, "max-in-requests", 100, "Maximum number of concurrent requests")
	flag.IntVar(&cfg.MaxUrls, "max-urls", 20, "Maximum number of urls per request")
	flag.IntVar(&cfg.MaxOutRequests, "max-out-requests", 4, "Maximum number of outgoing requests per incoming request")
	flag.DurationVar(&cfg.Timeout, "timeout", time.Second, "Timeout for outgoing request")
	flag.BoolVar(&help, "help", false, "Usage")

	if help {
		flag.Usage()
		os.Exit(-1)
	}
}

func Get() Config {
	o.Do(func() {
		flag.Parse()
	})
	return cfg
}
