package main

import (
	"log"
	"os"

	"github.com/soumadeep-sarkar/multi-threaded-proxy-server/config"
	"github.com/soumadeep-sarkar/multi-threaded-proxy-server/internal/cache"
)

// ProxyServer represents the main proxy server application
type ProxyServer struct {
	infoLogger  *log.Logger
	errorLogger *log.Logger
	cache       *cache.LRUCache
	config      *config.Config
	semaphore   chan struct{}
}

// NewProxyServer creates a new instance of the proxy server
func NewProxyServer(cfg *config.Config) *ProxyServer {
	infoLogger := log.New(os.Stdout, "[INFO] ", log.Ldate|log.Ltime|log.Lshortfile)
	errorLogger := log.New(os.Stderr, "[ERROR] ", log.Ldate|log.Ltime|log.Lshortfile)

	return &ProxyServer{
		infoLogger:  infoLogger,
		errorLogger: errorLogger,
		cache:       cache.NewLRUCache(cfg.CacheSize),
		config:      cfg,
		semaphore:   make(chan struct{}, cfg.MaxClients),
	}
}
