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

func constructInternalUrl(nodeIp string, locationId string) string {
	return "http://" + nodeIp + ":8000" + "/internal/" + locationId
}

func removeTrailingZeros(byteArr []byte) []byte {
	for i := len(byteArr) - 1; i >= 0; i-- {
		if byteArr[i] != 0 {
			return byteArr[:i+1]
		}
	}
	// If all bytes are zeros, return an empty slice
	return []byte{}

}
