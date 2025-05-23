package connection_pool

import (
	pb "big_o/protobuf_helper"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
)

const MAX_RETRIES int = 3
const RETRY_TIMEOUT = 5 * time.Second
const CLIENT_DOES_NOT_EXIST = "client does not exist."
const CONNECTION_DOES_NOT_EXIST = "connection does not exist."

var log = logrus.New()

type Connections = map[string]*grpc.ClientConn
type clientMap = map[string]pb.InternalClient

var connectionPool ConnectionPool = ConnectionPool{
	connections: make(map[string]*grpc.ClientConn),
	clients:     make(map[string]pb.InternalClient),
}

type ConnectionPool struct {
	mu          sync.Mutex
	connections Connections
	clients     clientMap
}

func GetClientFor(url string) (pb.InternalClient, error) {
	client, exists := connectionPool.clients[url]
	if exists {
		return client, nil
	}
	connectionPool.mu.Lock()
	defer connectionPool.mu.Unlock()

	client = pb.NewInternalClient(connectToServerWithRetries(url))
	connectionPool.clients[url] = client

	return client, nil
}

func connectToServerWithRetries(url string) *grpc.ClientConn {
	retriedTimes := 0

	for retriedTimes <= MAX_RETRIES {
		conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err == nil {
			warmupConnections(conn, url)
			return conn
		} else {
			log.Errorln("Acquiring connection for ", url, " failed. Retry number: ", retriedTimes)
			retriedTimes++
			log.Errorln("Connection failed due to: ", err)
			time.Sleep(RETRY_TIMEOUT)
		}
	}
	panic("Acquiring connections for " + url + " failed. Shutting down systems.")
}

func warmupConnections(conn *grpc.ClientConn, url string) {
	log.Infoln("Starting Waiting for Connection establishment for ", url)
	for conn.GetState() != connectivity.Ready {
		conn.Connect()
		log.Infoln("Waiting for connection to be ready for ", url, conn.GetState().String())
		time.Sleep(500 * time.Microsecond)
	}
	log.Infoln("Connection established for ", url, conn.GetState().String())
}