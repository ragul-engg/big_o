package main

import (
	"encoding/json"
	"log"
	"sync/atomic"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/pprof"
)

func setupRoutes(app *fiber.App) {
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	externalRoutes(app)

	internalRoutes(app)

	app.Use(pprof.New(pprof.Config{Prefix: "/logs"}))
}

func externalRoutes(app *fiber.App) {
	app.Put("/:locationId", func(c *fiber.Ctx) error {
		locationId, ok := c.AllParams()["locationId"]
		// log.Println("PUT Recieved: ", locationId)
		if !ok {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		payload := c.BodyRaw()

		err := processUpdateRequest(locationId, payload)

		if err != nil {
			log.Println("ERROR PUT endpoint: ", err)
			switch err.Error() {
			case MEMORY_FULL:
				return c.Status(507).SendString(err.Error())
			default:
				return c.Status(500).SendString(err.Error())
			}
		}

		// log.Println("PUT Ended: ", locationId)
		return c.SendStatus(fiber.StatusCreated)
	})

	app.Get("/:locationId", func(c *fiber.Ctx) error {
		locationId := c.Params("locationId", "")
		// // log.Println("Get Started for: ", locationId, "Data Store: ", dataStore)

		if locationId == "" {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		response, err := processGetRequest(locationId)
		// log.Println("Get location errors, response :", response, err)
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

		// log.Println("Process get response", response)

		body, err := json.Marshal(response)

		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		// log.Println("Get Ended for: ", locationId)
		return c.Status(200).Send(body)
	})

}

func internalRoutes(app *fiber.App) {

	app.Get("/internal/size", func(c *fiber.Ctx) error {
		size := atomic.LoadUintptr(&totalSize)
		// log.Println("Size: ", size)
		jsonn, err := json.Marshal(name{Size: size})
		if err != nil {
			return c.Status(500).SendString(err.Error())
		}
		return c.Status(200).Send(jsonn)
	})

	app.Get("/internal/:locationId", func(c *fiber.Ctx) error {
		locationId := c.Params("locationId", "")
		// log.Println("Got internal call to get data!")

		if locationId == "" {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		body, ok := dataStore[locationId]

		if !ok {
			return c.Status(404).SendString("location data not found")
		}

		return c.Status(200).Send(body.data)
	})

	app.Put("/internal/:locationId", func(c *fiber.Ctx) error {
		locationId := c.Params("locationId", "")
		// log.Println("Internal PUT Recieved: ", locationId)
		if locationId == "" {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		payload := c.BodyRaw()
		// // fmt.Println("Got internal call to replicate data: ", payload)

		updateChannel <- UpdateChannelPayload{locationId: locationId, encodedPayload: payload}
		// updateDataStore(locationId, payload)

		// log.Println("Internal PUT Ended: ", locationId)
		return c.SendStatus(fiber.StatusCreated)
	})
}

type name struct {
	Size uintptr
}
