package main

import	"github.com/gofiber/fiber/v2"

func setupRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	app.Put("/:locationId", func(c *fiber.Ctx) error {
		locationId, ok := c.AllParams()["locationId"]
		if !ok {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		payload := c.BodyRaw()

		err := processUpdateRequest(locationId, payload)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	app.Put("/internal/:locationId", func(c *fiber.Ctx) error {
		locationId := c.Params("locationId", "")

		if locationId == "" {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		payload := c.BodyRaw()

		updateDataStore(locationId, payload)

		return c.SendStatus(fiber.StatusCreated)
	})
}
