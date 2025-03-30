package main

import (
	"github.com/gofiber/fiber/v2"
)

var fiberConfig = fiber.Config{
	// Network performance tuning
	Network:          "tcp4", // Prefer IPv4 for performance
	DisableKeepalive: false,  // Enable keepalive connections

	// Request body limit
	Immutable: true,
}
