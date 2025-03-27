package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
)

func setupRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	app.Put("/:locationId", func(c *fiber.Ctx) error {
		log.Println("PUT Recieved: ", dataStore)
		locationId, ok := c.AllParams()["locationId"]
		if !ok {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		payload := c.BodyRaw()

		err := processUpdateRequest(locationId, payload)

		if err != nil {
			switch err.Error() {
			case MEMORY_FULL:
				return c.Status(507).SendString(err.Error())
			default:
				return c.Status(500).SendString(err.Error())
			}
		}

		log.Println("PUT Ended: ", dataStore)
		return c.SendStatus(fiber.StatusCreated)
	})

	app.Put("/internal/:locationId", func(c *fiber.Ctx) error {
		locationId := c.Params("locationId", "")

		if locationId == "" {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		payload := c.BodyRaw()
		fmt.Println("Got internal call to replicate data: ", payload)

		updateDataStore(locationId, payload)

		return c.SendStatus(fiber.StatusCreated)
	})

	app.Get("/:locationId", func(c *fiber.Ctx) error {
		locationId := c.Params("locationId", "")
		log.Println("Get Started for: ", locationId, "Data Store: ", dataStore)

		if locationId == "" {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		response, err := processGetRequest(locationId)
		log.Println("Get location errors, response :", response, err)
		if err != nil {
			switch err.Error() {
			case LOCATION_ID_NOT_FOUND:
				return c.Status(404).SendString(err.Error())
			case COULD_NOT_RECONSTRUCT_DATA:
				return c.Status(500).SendString(err.Error())
			default:
				return c.SendStatus(500)
			}
		}

		fmt.Println("Process get response", response)

		body, err := json.Marshal(response)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		log.Println("Get Ended for: ", locationId, "Data Store: ", dataStore)
		return c.Status(200).Send(body)
	})

	app.Get("/internal/:locationId", func(c *fiber.Ctx) error {
		locationId := c.Params("locationId", "")
		fmt.Println("Got internal call to get data!")

		if locationId == "" {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		body, ok := dataStore[locationId]

		if !ok {
			return c.Status(404).SendString("location data not found")
		}

		return c.Status(200).Send(body.data)
	})
}
