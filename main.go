package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/klauspost/reedsolomon"
)

const NUMBER_OF_DATA_SHARDS = 4
const NUMBER_OF_PARITY_SHARDS = 3
const TOTAL_SHARDS = NUMBER_OF_DATA_SHARDS + NUMBER_OF_PARITY_SHARDS
const TOTAL_NODES = TOTAL_SHARDS

type Payload struct {
	Id               string
	Seismic_activity float32
	Temperature_c    float32
	Radiation_level  float32
}

var payload = Payload{Id: "id1", Seismic_activity: 12.3, Temperature_c: 23.4, Radiation_level: 45.3}

var dataStore map[string]LocationData = make(map[string]LocationData)

type LocationData struct {
	data              []byte
	modificationCount int
}

// 7,3
// Note that number of parity shards will give you maximum tolerated failures, so here 3 failures is the maximum tolerated.
func processPayload(payload []byte) ([][]byte, error) {
	enc, _ := reedsolomon.New(NUMBER_OF_DATA_SHARDS, NUMBER_OF_PARITY_SHARDS)
	data := make([][]byte, TOTAL_SHARDS)

	chunkSize := len(payload) / NUMBER_OF_DATA_SHARDS

	// Create all shards, size them at chunkSize each
	for i := range TOTAL_SHARDS {
		data[i] = make([]byte, chunkSize)
	}

	populateDataChunks(payload, chunkSize, data)

	fmt.Println("***************** Initial Data")
	printData(data)

	err := enc.Encode(data)
	return data, err
}

var currentNodeIp string
var nodeIpMap map[int]string

func loadEnv() {
	currentNodeIp = os.Getenv("CURRENT_NODE_IP")
	ALL_NODE_IPS := os.Getenv("ALL_NODE_IPS")

	if len(ALL_NODE_IPS) == 0 || len(currentNodeIp) == 0 {
		panic("Oh no we are doomed!")
	}
	nodeIps := strings.Split(ALL_NODE_IPS, ",")
	nodeIpMap = make(map[int]string)
	for index, ip := range nodeIps {
		nodeIpMap[index] = ip
	}
}

func main() {
	loadEnv()
	app := fiber.New()

	// // // Define a route for the Hello World message
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
			return c.SendStatus(500)
		}

		return c.SendStatus(fiber.StatusCreated)
	})

	app.Put("/internal/:locationId", func(c *fiber.Ctx) error {
		locationId, ok := c.AllParams()["locationId"]

		if !ok {
			return c.SendStatus(fiber.ErrBadRequest.Code)
		}

		payload := c.BodyRaw()

		updateDataStore(locationId, payload)

		return c.SendStatus(fiber.StatusCreated)
	})
}

func processUpdateRequest(locationId string, payload []byte) error {
	encodedPayload, err := processPayload(payload)

	if err != nil {
		return err
	}

	yourShare, err := replicateData(encodedPayload)

	if err != nil {
		return err
	}

	updateDataStore(locationId, yourShare)	

	return nil
}

func updateDataStore(locationId string, dataShard []byte) {
	existingValue, exists := dataStore[locationId]
	if exists {
		dataStore[locationId] = LocationData{data: dataShard, modificationCount: existingValue.modificationCount + 1}
	} else {
		dataStore[locationId] = LocationData{data: dataShard, modificationCount: 1}
	}
}

func makeRequest(c *fiber.Ctx, url string) (error) {
	agent := fiber.Post(url)
	agent.Body(c.Body()) // set body received by request
	statusCode, body, errs := agent.Bytes()
	if len(errs) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"errs": errs,
		})
	}

    // pass status code and body received by the proxy
	return c.Status(statusCode).Send(body)
}

func replicateData(encodedPayload [][]byte) ([]byte, error) {
	var myShare []byte
	for index, value := range encodedPayload {
		nodeIp := nodeIpMap[index]
		if nodeIp != currentNodeIp {
		} else {
			myShare = value
		}
	}
	return myShare, nil
}

func populateDataChunks(out []byte, chunkSize int, data [][]byte) {
	var index = 0
	for value := range slices.Chunk(out, chunkSize) {
		data[index] = value
		index++
	}
}

func printData(data [][]byte) {
	for i, j := range data[:TOTAL_SHARDS] {
		fmt.Println("Index ", i, " Data", j)
	}
}

func reconstruct(data [][]byte) Payload {
	var payload Payload
	var byteArr []byte

	for _, value := range data[:NUMBER_OF_DATA_SHARDS] {
		byteArr = append(byteArr, value...)
	}

	json.Unmarshal(byteArr, &payload)

	return payload
}
