package connection_pool

import (
	pb "big_o/protobuf_helper"
	"errors"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const MAX_RETRIES int = 3
const RETRY_TIMEOUT = 5 * time.Second
const CLIENT_DOES_NOT_EXIST = "client does not exist."
const CONNECTION_DOES_NOT_EXIST = "connection does not exist."

type Connections = map[string]*grpc.ClientConn
type clientMap = map[string]pb.InternalClient

var connectionPool ConnectionPool

type ConnectionPool struct {
	connections Connections
	clients     clientMap
}

func InitialiseConnectionPool(urls []string) {
	connections := acquireConnections(urls)
	clients := generateClients(connections)
	connectionPool = ConnectionPool{
		connections: connections,
		clients:     clients,
	}
}

func GetConnectionFor(url string) (*grpc.ClientConn, error) {
	conn, exists := connectionPool.connections[url]
	if exists {
		return conn, nil
	}
	return nil, errors.New("connection does not exist.")
}

func GetClientFor(url string) (pb.InternalClient, error) {
	client, exists := connectionPool.clients[url]
	if exists {
		return client, nil
	}
	return nil, errors.New("client does not exist.")
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

func generateClients(connections Connections) clientMap {
	clientMap := make(map[string](pb.InternalClient))
	for url, connection := range connections {
		clientMap[url] = pb.NewInternalClient(connection)
	}
	return clientMap
}
