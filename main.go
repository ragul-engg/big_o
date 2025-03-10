package main

import (
	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()

	// Define a route for the Hello World message
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	// Start the server on port 3000
	app.Listen(":3000")
}
