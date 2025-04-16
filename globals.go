package main
import (
	"github.com/sirupsen/logrus"
	"github.com/klauspost/reedsolomon"
	"flag"
)
var portPtr = flag.String("port", "8000", "send port number")
var updateChannel = make(chan UpdateChannelPayload)

var logger = logrus.New()



var dataStore map[string]LocationData = make(map[string]LocationData)
var currentNodeIp string
var currentNodeGrpcIp string
var nodeIps []string
var grpcIps []string
var enc, _ = reedsolomon.New(NUMBER_OF_DATA_SHARDS, NUMBER_OF_PARITY_SHARDS, reedsolomon.WithMaxGoroutines(25))