package main
import (
	"os"
	"strings"
	"flag"
)

func loadEnv() {
	currentNodeIp = os.Getenv("CURRENT_NODE_IP")
	allNodeIps := os.Getenv("ALL_NODE_IPS")

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


func readFlags(portPtr *string) {
	flag.Parse()
}