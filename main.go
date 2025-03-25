package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
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

const LOCATION_ID_NOT_FOUND = "location id not found"
const COULD_NOT_RECONSTRUCT_DATA = "could not reconstruct data"

type Payload struct {
	Id               string  `json:"id"`
	Seismic_activity float32 `json:"seismic_activity"`
	Temperature_c    float32 `json:"temperature_c"`
	Radiation_level  float32 `json:"radiation_level"`
}
type LocationData struct {
	data              []byte
	modificationCount int
}

var dataStore map[string]LocationData = make(map[string]LocationData)
var currentNodeIp string
var nodeIpMap map[int]string

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

func loadEnv() {
	currentNodeIp = os.Getenv("CURRENT_NODE_IP")
	allNodeIps := os.Getenv("ALL_NODE_IPS")

	fmt.Println(currentNodeIp)
	fmt.Println(allNodeIps)

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
			err := makePutRequest(constructInternalUrl(nodeIp, locationId), value)
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

type ResponsePayload struct {
	Payload
	ModificationCount int `json:"modification_count"`
}

func processGetRequest(locationId string) (ResponsePayload, error) {
	enc, _ := reedsolomon.New(NUMBER_OF_DATA_SHARDS, NUMBER_OF_PARITY_SHARDS)
	data := make([][]byte, TOTAL_SHARDS)
	myLocationData, ok := dataStore[locationId]

	if !ok {
		return ResponsePayload{}, errors.New(LOCATION_ID_NOT_FOUND)
	}

	chunkSize := len(myLocationData.data)

	// chunkSizeFloat := float64(len(payload)) / float64(NUMBER_OF_DATA_SHARDS)
	// chunkSize := int(math.Ceil(chunkSizeFloat))
	// fmt.Println("Payload Size: ", len(payload))
	// fmt.Println("Chunk size: ", chunkSize)
	// Create all shards, size them at chunkSize each

	for i := range TOTAL_SHARDS {
		data[i] = make([]byte, chunkSize)
	}

	getAllShards(data, locationId, chunkSize)

	err := enc.Reconstruct(data)
	fmt.Println("Reconstructed data with encoder: ", data)

	if err != nil {
		fmt.Println(err.Error())
		return ResponsePayload{}, errors.New(COULD_NOT_RECONSTRUCT_DATA)
	}

	reconstructedData := reconstruct(data)
	fmt.Println("Processing payload: ", reconstructedData, "error:", err)
	return ResponsePayload{Payload: reconstructedData, ModificationCount: myLocationData.modificationCount}, nil
}

func getAllShards(data [][]byte, locationId string, chunkSize int) {
	fmt.Println("Getting all Data!")
	for index, nodeIp := range nodeIpMap {
		internalUrl := constructInternalUrl(nodeIp, locationId)
		fmt.Println("running for", nodeIp, "index", index, "url", internalUrl)
		if nodeIp != currentNodeIp {
			res, err := makeGetRequest(internalUrl)
			if err != nil {
				fmt.Println("Something went wrong with Get requests: ", err)
			}
			data[index] = padRightWithZeros(res, chunkSize)
		} else {
			data[index] = padRightWithZeros(dataStore[locationId].data, chunkSize)
		}
	}
	printData(data)
}

func makeGetRequest(url string) ([]byte, error) {
	resp, err := http.Get(url)

	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func reconstruct(data [][]byte) Payload {
	var payload Payload
	var byteArr []byte

	for _, value := range data[:NUMBER_OF_DATA_SHARDS] {
		byteArr = append(byteArr, value...)
	}

	fmt.Println(string(byteArr))
	
	err := json.Unmarshal(byteArr, &payload)

	fmt.Println("Error while unmarshalling", err)

	return payload
}

func readFlags(portPtr *string) {
	flag.Parse()
	fmt.Println("port:", *portPtr)
}
