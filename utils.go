package main

import "fmt"

func padRightWithZeros(arr []byte, length int) []byte {
	// Calculate how many zeros need to be added
	zerosToAdd := length - len(arr)

	// Append zeros to the array
	for i := 0; i < zerosToAdd; i++ {
		arr = append(arr, 0)
	}

	fmt.Println("Padded Array: ", arr)

	return arr
}
