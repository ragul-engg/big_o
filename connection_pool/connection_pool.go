package connection_pool

import (
	// pb "big_o/protobuf_helper"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MAX_RETRIES int = 3
const RETRY_TIMEOUT = 5 * time.Second

type Connections = map[string]*grpc.ClientConn

var connectionPool Connections

func InitialiseConnectionPool(urls []string) {
	connectionPool = acquireConnections(urls)
}

func GetConnectionFor(url string) (*grpc.ClientConn, error) {
	conn, ok := connectionPool[url]
	if ok {
		return conn, nil
	}
	return nil, errors.New("connection does not exist.")
}

func acquireConnections(urls []string) Connections {
	connections := make(map[string]*grpc.ClientConn)
	for _, url := range urls {
		conn := connectToServerWithRetries(url)
		connections[url] = conn
	}

	return connections
}

func connectToServerWithRetries(url string) *grpc.ClientConn {
	retriedTimes := 0
	for retriedTimes <= MAX_RETRIES {
		// grpc.WithKeepaliveParams(keepalive.ClientParameters{T})
		conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			return conn
		} else {
			log.Println("Acquiring connection for ", url, " failed. Retry number: ", retriedTimes)
			log.Println("Connection failed due to: ", err)
			time.Sleep(RETRY_TIMEOUT)
		}
	}
	panic("Acquiring connections for " + url + " failed. Shutting down systems.")
}
