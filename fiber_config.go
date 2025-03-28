package main

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

var fiberConfig = fiber.Config{
		Prefork:                 true,  // Enable process forking
		// ServerHeader:            "OptimizedServer",
		DisableStartupMessage:   false,
		ReadTimeout:             10 * time.Second,
		WriteTimeout:            10 * time.Second,
		IdleTimeout:             120 * time.Second,
		ReadBufferSize:          4096,  // Optimize buffer sizes
		WriteBufferSize:         4096,
		// CompressedFileSuffix:    ".fiber.gz",
		// ProxyHeader:             "X-Forwarded-For",
		
		// Network performance tuning
		Network:                 "tcp4", // Prefer IPv4 for performance
		DisableKeepalive:        false,  // Enable keepalive connections
		
		// Request body limit
		BodyLimit:               10 * 1024 * 1024, // 10MB
		Immutable: true,
}