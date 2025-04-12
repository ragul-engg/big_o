package main

import (
	connectionPool "big_o/connection_pool"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"math"
	"net/http"
	_ "net/http/pprof"
	"os"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/klauspost/reedsolomon"
)

const NUMBER_OF_DATA_SHARDS int = 4
const NUMBER_OF_PARITY_SHARDS int = 3
const TOTAL_SHARDS int = NUMBER_OF_DATA_SHARDS + NUMBER_OF_PARITY_SHARDS
const TOTAL_NODES = TOTAL_SHARDS

const LOCATION_ID_NOT_FOUND = "location id not found"
const COULD_NOT_RECONSTRUCT_DATA = "could not reconstruct data"
const MEMORY_FULL = "memory is full."

var portPtr = flag.String("port", "8000", "send port number")
var updateChannel = make(chan UpdateChannelPayload)

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

type ResponsePayload struct {
	Payload
	ModificationCount int `json:"modification_count"`
}

var dataStore map[string]LocationData = make(map[string]LocationData)
var currentNodeIp string
var currentNodeGrpcIp string
var nodeIps []string
var grpcIps []string
var enc, _ = reedsolomon.New(NUMBER_OF_DATA_SHARDS, NUMBER_OF_PARITY_SHARDS, reedsolomon.WithMaxGoroutines(25))

func processPayload(payload []byte) ([][]byte, error) {

	data := make([][]byte, TOTAL_SHARDS)

	chunkSizeFloat := float64(len(payload)) / float64(NUMBER_OF_DATA_SHARDS)
	chunkSize := int(math.Ceil(chunkSizeFloat))

	for i := range TOTAL_SHARDS {
		data[i] = make([]byte, chunkSize)
	}

	populateDataChunks(payload, chunkSize, data)

	err := enc.Encode(data)

	return data, err
}

func loadEnv() {
	currentNodeIp = os.Getenv("CURRENT_NODE_IP")
	allNodeIps := os.Getenv("ALL_NODE_IPS")

	if len(allNodeIps) == 0 || len(currentNodeIp) == 0 {
		panic("Oh no we are doomed!")
	}
	// nodeIps := strings.Split(allNodeIps, ",")
	nodeIps = strings.Split(allNodeIps, ",")
	grpcIps = getGrpcIps(nodeIps)
	log.Println("GRPC IPs:", grpcIps)
	grpcIp, err := getGrpcIpFor(currentNodeIp)

	if err != nil {
		panic("Grpc conversion failed, something wrong.")
	}

	currentNodeGrpcIp = grpcIp
	log.Println("loading with current ip and node ips", currentNodeIp, nodeIps)
}

func main() {
	loadEnv()
	runtime.GOMAXPROCS(runtime.NumCPU())

	go dataStoreWriter()
	readFlags(portPtr)
	var port = ":" + *portPtr

	app := fiber.New(fiberConfig)

	setupRoutes(app)
	grpcServerPort, err := strconv.ParseInt(*portPtr, 10, 32)
	if err != nil {
		panic("Unable to start grpc server at: " + string(grpcServerPort))
	}

	go startGrpcServer(strconv.Itoa(int(grpcServerPort + 1000)))

	app.Listen(port)
}

func processUpdateRequest(locationId string, payload []byte) error {
	if !allowWrites() {
		return errors.New(MEMORY_FULL)
	}
	encodedPayload, err := processPayload(payload)

	if err != nil {
		return err
	}

	log.Println("Full data: ", encodedPayload)
	yourShare, err := replicateDataGrpc(locationId, encodedPayload)

	if err != nil {
		return err
	}
	// do go routine that puts into a channel to update data store, channel takes in update payload
	updateChannel <- UpdateChannelPayload{locationId: locationId, encodedPayload: yourShare}
	// updateDataStore(locationId, yourShare)

	return nil
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
	var myShare []byte
	for index, value := range encodedPayload {
		nodeIp := nodeIps[index]
		if nodeIp != currentNodeIp {
			err := makePutRequest(constructInternalUrl(nodeIp, locationId), value)
			if err != nil {
				fmt.Println("Something went wrong with post requests: ", err)
			}
		} else {
			log.Println("Taking my share: ", nodeIp)
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

func processGetRequest(locationId string) (ResponsePayload, error) {
	enc, _ := reedsolomon.New(NUMBER_OF_DATA_SHARDS, NUMBER_OF_PARITY_SHARDS)

	data := make([][]byte, TOTAL_SHARDS)
	myLocationData, ok := dataStore[locationId]

	if !ok {
		return ResponsePayload{}, errors.New(LOCATION_ID_NOT_FOUND)
	}

	chunkSize := len(myLocationData.data)

	for i := range TOTAL_SHARDS {
		data[i] = make([]byte, chunkSize)
	}

	getAllShards(data, locationId, chunkSize)

	err := enc.Reconstruct(data)

	if err != nil {
		fmt.Println(err.Error())
		return ResponsePayload{}, errors.New(COULD_NOT_RECONSTRUCT_DATA)
	}

	reconstructedData := reconstruct(data)
	return ResponsePayload{Payload: reconstructedData, ModificationCount: myLocationData.modificationCount}, nil
}

func getAllShards(data [][]byte, locationId string, chunkSize int) {
	fmt.Println("Getting all Data!")
	for index, nodeIp := range nodeIps {
		internalUrl := constructInternalUrl(nodeIp, locationId)
		fmt.Println("running for", nodeIp, "index", index, "url", internalUrl)
		if nodeIp != currentNodeIp {
			res, err := makeGetRequest(internalUrl)
			if err != nil {
				fmt.Println("Something went wrong with Get requests: ", err)
				data[index] = nil
			} else {
				data[index] = padRightWithZeros(res, chunkSize)
			}
		} else {
			data[index] = padRightWithZeros(dataStore[locationId].data, chunkSize)
		}
	}
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

	trimmedByteArr := removeTrailingZeros(byteArr)
	err := json.Unmarshal(trimmedByteArr, &payload)

	fmt.Println("Error while unmarshalling", err)

	return payload
}

func readFlags(portPtr *string) {
	flag.Parse()
	fmt.Println("port:", *portPtr)
}
