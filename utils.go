package main

import (
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
)

func padRightWithZeros(arr []byte, length int) []byte {
	// Calculate how many zeros need to be added
	zerosToAdd := length - len(arr)

	// Append zeros to the array
	for i := 0; i < zerosToAdd; i++ {
		arr = append(arr, 0)
	}

	// fmt.Println("Padded Array: ", arr)

	return arr
}

func constructInternalUrl(nodeIp string, locationId string) string {
	return nodeIp + "/internal/" + locationId
}

func removeTrailingZeros(byteArr []byte) []byte {
	for i := len(byteArr) - 1; i >= 0; i-- {
		if byteArr[i] != 0 {
			return byteArr[:i+1]
		}
	}
	return []byte{}
}

func getGrpcIps(nodeIps []string) []string {
	convertedIPs := make([]string, 0, len(nodeIps))

	for _, ip := range nodeIps {
		// Parse the URL
		grpcIp, err := getGrpcIpFor(ip)
		if err != nil {
			logger.Errorf("Error parsing URL %s: %v\n", ip, err)
			continue
		}

		convertedIPs = append(convertedIPs, grpcIp)
	}

	return convertedIPs
}

func getGrpcIpFor(ip string) (string, error) {
	parsedUrl, err := url.Parse(ip)

	if err != nil {
		return "", err
	}

	hostPort := parsedUrl.Host
	parts := strings.Split(hostPort, ":")

	if len(parts) != 2 {
		logger.Errorf("Invalid host:port format in %s\n", ip)
		return "", errors.New("invalid host:port format")
	}

	host := parts[0]
	portStr := parts[1]

	port, err := strconv.Atoi(portStr)
	if err != nil {
		logger.Errorf("Error converting port %s: %v\n", portStr, err)
		return "", err
	}

	newPort := port + 1000

	newHostPort := fmt.Sprintf("%s:%d", host, newPort)
	logger.Infof("GRPC IP for ip: %v is %v", ip, newHostPort)
	return newHostPort, nil
}
