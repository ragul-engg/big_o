package main

import (
	"encoding/json"
	"errors"
	"io"

	"math"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"slices"
	"strconv"
	"github.com/gofiber/fiber/v2"
	"github.com/klauspost/reedsolomon"
	logrus "github.com/sirupsen/logrus"
)

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


func main() {
	loadEnv()
	runtime.GOMAXPROCS(runtime.NumCPU())
	logger.SetLevel(logrus.InfoLevel)
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

	logger.Debug("Full data: ", encodedPayload)
	yourShare, err := replicateDataGrpc(locationId, encodedPayload)

	if err != nil {
		return err
	}
	// do go routine that puts into a channel to update data store, channel takes in update payload
	updateChannel <- UpdateChannelPayload{locationId: locationId, encodedPayload: yourShare}
	// updateDataStore(locationId, yourShare)

	return nil
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
		logger.Errorln(err.Error())
		return ResponsePayload{}, errors.New(COULD_NOT_RECONSTRUCT_DATA)
	}

	reconstructedData := reconstruct(data)
	return ResponsePayload{Payload: reconstructedData, ModificationCount: myLocationData.modificationCount}, nil
}

func getAllShards(data [][]byte, locationId string, chunkSize int) {
	logger.Debugln("Getting all Data!")
	for index, nodeIp := range nodeIps {
		internalUrl := constructInternalUrl(nodeIp, locationId)
		logger.Debugln("running for", nodeIp, "index", index, "url", internalUrl)
		if nodeIp != currentNodeIp {
			res, err := makeGetRequest(internalUrl)
			if err != nil {
				logger.Errorln("Something went wrong with Get requests: ", err)
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

	if err != nil {
		logger.Errorln("Error while unmarshalling", err)
	}

	return payload
}

