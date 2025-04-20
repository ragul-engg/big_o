package main

import (
	"flag"
	"runtime"

	"github.com/klauspost/reedsolomon"
	"github.com/sirupsen/logrus"
)

var portPtr = flag.String("port", "8000", "send port number")
var updateChannel = make(chan UpdateChannelPayload)

var logger = logrus.New()

var currentNodeIp string
var currentNodeGrpcIp string
var nodeIps []string
var grpcIps []string
var enc, _ = reedsolomon.New(NUMBER_OF_DATA_SHARDS, NUMBER_OF_PARITY_SHARDS, reedsolomon.WithMaxGoroutines(25))

var datastoreMemoryUsage runtime.MemStats
