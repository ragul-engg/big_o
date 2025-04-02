package main

import (
	pb "big_o/protobuf_helper"
	"context"
	"errors"
	"net"

	"google.golang.org/grpc"
)

type protobufServer struct {
	pb.UnimplementedInternalServer
}

var protobufServer1 = protobufServer{}

func (s *protobufServer) UpsertLocationData(ctx context.Context, req *pb.UpsertPayload) (*pb.Empty, error) {
	// log.Printf("Received upsert request for location ID: %s\n", req.LocationId)

	if req.LocationId == "" {
		return &pb.Empty{}, errors.New("locationId cannot be empty")
	}

	updateChannel <- UpdateChannelPayload{locationId: req.LocationId, encodedPayload: req.EncodedPayload}
	// log.Println("Internal update call done!")
	return &pb.Empty{}, nil
}

func startGrpcServer(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		// log.Fatalf("failed to listen: %v", err)
	}

	// Create a gRPC server object
	grpcServer := grpc.NewServer()

	// Register our service with the gRPC server
	pb.RegisterInternalServer(grpcServer, &protobufServer{})

	// log.Println("Starting gRPC server on port ", port)

	// Start the server
	if err := grpcServer.Serve(lis); err != nil {
		// log.Fatalf("failed to serve: %v", err)
	}

}
