package main

import (
	"fmt"
	// "github.com/gofiber/fiber/v2"
	"github.com/klauspost/reedsolomon"
)

func main() {
	enc, _ := reedsolomon.New(10, 3)
	data := make([][]byte, 13)

	// Create all shards, size them at 500 each
	for i := range 13 {
		data[i] = make([]byte, 500)
		data[i] = []byte(RandStringBytes(500))
	}
	// The above allocations can also be done by the encoder:
	// data := enc.(reedsolomon.Extended).AllocAligned(50000)

	// Fill some data into the data shards
	// for i, in := range data[:10] {
	// 	for j := range in {
	// 		in[j] = byte((i + j) & 0xff)
	// 	}
	// }
	fmt.Println("***************** Initial Data")

	printData(data)

	err := enc.Encode(data)
	fmt.Println("Encoding output: ", err)
	ok, err := enc.Verify(data)
	fmt.Println("Verified: ", ok, err)


	// Delete two data shards
	data[3] = nil
	data[7] = nil

	fmt.Println("********************************Deleted Data: ")
	printData(data)

	// Reconstruct the missing shards
	err = enc.Reconstruct(data)
	fmt.Println("Error or nah : ", err)
	fmt.Println("*************************Reconstructed Data: ")
	printData(data)
	// app := fiber.New()

	// // Define a route for the Hello World message
	// app.Get("/health", func(c *fiber.Ctx) error {
	// 	return c.SendStatus(200)
	// })

	// // Start the server on port 3000
	// app.Listen(":3000")
}

func printData(data [][]byte ) {
	for i, j := range data[:10] {
		fmt.Println("Index ",i, " Data", j)
	}
}