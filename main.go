package main

import (
	// "bytes"
	"encoding/json"
	"fmt"
	"slices"

	// "github.com/gofiber/fiber/v2"
	"github.com/klauspost/reedsolomon"
)
const NUMBER_OF_DATA_SHARDS = 4
const NUMBER_OF_PARITY_SHARDS = 3
const TOTAL_SHARDS = NUMBER_OF_DATA_SHARDS + NUMBER_OF_PARITY_SHARDS
type Payload struct {
	Id string
	Seismic_activity float32
	Temperature_c float32
	Radiation_level float32
}

var payload = Payload{Id: "id1", Seismic_activity: 12.3, Temperature_c: 23.4, Radiation_level: 45.3}  

// 7,3 
// Note that number of parity shards will give you maximum tolerated failures, so here 3 failures is the maximum tolerated.
func main() {
	out,_ := json.Marshal(payload)
	fmt.Println("Marshalled output here --> ", out)
	var unmarshalled Payload
	json.Unmarshal(out, &unmarshalled)

	fmt.Println("Unmarshalled output ---> ", unmarshalled)
	
	
	enc, _ := reedsolomon.New(NUMBER_OF_DATA_SHARDS, NUMBER_OF_PARITY_SHARDS)
	data := make([][]byte, TOTAL_SHARDS)
	
	chunkSize := len(out) / NUMBER_OF_DATA_SHARDS

	// Create all shards, size them at chunkSize each
	for i := range TOTAL_SHARDS {
		data[i] = make([]byte, chunkSize)
	}

	populateDataChunks(out, chunkSize, data)

	fmt.Println("***************** Initial Data")
	printData(data)

	err := enc.Encode(data)
	fmt.Println("Encoding output: ", err)
	ok, err := enc.Verify(data)
	fmt.Println("Verified: ", ok, err)


	// // Delete 3 data shards
	data[3] = nil
	data[6] = nil
	data[1] = nil

	fmt.Println("********************************Deleted Data: ")
	// printData(data)

	// Reconstruct the missing shards
	err = enc.Reconstruct(data)
	fmt.Println("Error or nah : ", err)
	fmt.Println("*************************Reconstructed Data: ")
	// printData(data)

	decodedPayload := reconstruct(data)
	fmt.Println("Reconstructed data: ", decodedPayload)
	// // app := fiber.New()

	// // // Define a route for the Hello World message
	// // app.Get("/health", func(c *fiber.Ctx) error {
	// // 	return c.SendStatus(200)
	// // })

	// // // Start the server on port 3000
	// app.Listen(":3000")
}

func populateDataChunks(out []byte, chunkSize int, data [][]byte) {
	var index = 0
	for value := range slices.Chunk(out, chunkSize) {
		data[index] = value
		index++
	}
}

func printData(data [][]byte ) {
	for i, j := range data[:TOTAL_SHARDS] {
		fmt.Println("Index ",i, " Data", j)
	}
}

func reconstruct(data [][]byte) Payload{
	var payload Payload
	var byteArr []byte

	for _, value := range data[:NUMBER_OF_DATA_SHARDS] {
		byteArr = append(byteArr, value...)
	}

	json.Unmarshal(byteArr, &payload)

	return payload
}