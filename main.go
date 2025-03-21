package main

import (
	// "bytes"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"net/http"
	"os"
	"slices"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/klauspost/reedsolomon"
)

const NUMBER_OF_DATA_SHARDS = 4
const NUMBER_OF_PARITY_SHARDS = 3
const TOTAL_SHARDS = NUMBER_OF_DATA_SHARDS + NUMBER_OF_PARITY_SHARDS
const TOTAL_NODES = TOTAL_SHARDS

type Payload struct {
	Id               string `json:"id"`
	Seismic_activity float32 `json:"seismic_activity"`
	Temperature_c    float32 `json:"temperature_c"`
	Radiation_level  float32 `json:"radiation_level"`
}

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

	chunkSizeFloat := float64(len(payload)) / float64(NUMBER_OF_DATA_SHARDS)
	chunkSize := int(math.Ceil(chunkSizeFloat))
	fmt.Println("Payload Size: ", len(payload))
	fmt.Println("Chunk size: ", chunkSize)
	// Create all shards, size them at chunkSize each

	for i := range TOTAL_SHARDS {
		data[i] = make([]byte, chunkSize)
	}

	populateDataChunks(payload, chunkSize, data)

	fmt.Println("***************** Initial Data")
	printData(data)

	err := enc.Encode(data)

	fmt.Println("Processing payload: ", err)
	return data, err
}

var currentNodeIp string
var nodeIpMap map[int]string

func loadEnv() {
	currentNodeIp = os.Getenv("CURRENT_NODE_IP")
	allNodeIps := os.Getenv("ALL_NODE_IPS")

	if len(allNodeIps) == 0 || len(currentNodeIp) == 0 {
		panic("Oh no we are doomed!")
	}
	nodeIps := strings.Split(allNodeIps, ",")
	nodeIpMap = make(map[int]string)
	for index, ip := range nodeIps {
		nodeIpMap[index] = ip
	}
}


func main() {
	loadEnv()
	portPtr := flag.String("port", "8000", "send port number")
	readFlags(portPtr)
	var port = ":" + *portPtr

	app := fiber.New()
	setupRoutes(app)
	app.Listen(port)
}

func processUpdateRequest(locationId string, payload []byte) error {
	encodedPayload, err := processPayload(payload)

	if err != nil {
		return err
	}

	yourShare, err := replicateData(locationId, encodedPayload)

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

func makePutRequest(url string, payload []byte) error {
	fmt.Println("Making request to: ", url)
	client := http.Client{
		Timeout: 3 * time.Second,
	}
	byteBuffer := bytes.NewBuffer(payload)
	request, err := http.NewRequest(
		http.MethodPut,
		url,
		byteBuffer,
	)
	if err != nil {
		return err
	}

	_, err = client.Do(request)

	if err != nil {
		return err
	}

	return nil
}

func replicateData(locationId string, encodedPayload [][]byte) ([]byte, error) {
	fmt.Println("replicating Data!")
	var myShare []byte
	for index, value := range encodedPayload {
		nodeIp := nodeIpMap[index]
		if nodeIp != currentNodeIp {
			err := makePutRequest(constructUrl(nodeIp, locationId), value)
			if err != nil {
				fmt.Println("Something went wrong with post requests: ", err)
			}
		} else {
			myShare = value
		}
	}
	return myShare, nil
}

func populateDataChunks(out []byte, chunkSize int, data [][]byte) {
	var index = 0
	for value := range slices.Chunk(out, chunkSize) {
		data[index] = padRightWithZeros(value, chunkSize)
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

func constructUrl(nodeIp string, locationId string) string {
	return nodeIp + "/internal/" + locationId
}

func readFlags(portPtr *string) {
	flag.Parse()
	fmt.Println("port:", *portPtr)
}
