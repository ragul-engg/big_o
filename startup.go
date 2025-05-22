package main

import (
	"os"
	"strings"
)

func loadEnv() {
	currentNodeIp = os.Getenv("CURRENT_NODE_IP")
	allNodeIps := os.Getenv("ALL_NODE_IPS")
	portNum = os.Getenv("PORT")

	if len(allNodeIps) == 0 || len(currentNodeIp) == 0 {
		panic("Oh no we are doomed!")
	}
	nodeIps = strings.Split(allNodeIps, ",")
	grpcIps = getGrpcIps(nodeIps)

	logger.Infof("GRPC IPs: %v", grpcIps)
	grpcIp, err := getGrpcIpFor(currentNodeIp)

	if err != nil {
		panic("Grpc conversion failed, something wrong.")
	}

	currentNodeGrpcIp = grpcIp
	logger.Infof("Loading with current ip: %v . Node ips: %v ", currentNodeIp, nodeIps)
}

func setLoggingToFile() {
	file, err := os.OpenFile("./logrus_"+portNum+".log", os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	if err == nil {
		logger.SetOutput(file)
	} else {
		logger.Info("Failed to log to file, using default stderr")
	}

	defer file.Close()
}
