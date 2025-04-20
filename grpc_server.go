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

func (s *protobufServer) UpsertLocationData(ctx context.Context, req *pb.UpsertPayload) (*pb.Empty, error) {
	logger.Debugf("Received upsert request for location ID: %s\n", req.LocationId)

	if req.LocationId == "" {
		return &pb.Empty{}, errors.New("locationId cannot be empty")
	}

	updateChannel <- UpdateChannelPayload{locationId: req.LocationId, encodedPayload: req.EncodedPayload}
	logger.Debugln("Internal update call done!")
	return &pb.Empty{}, nil
}

func (s *protobufServer) HealthCheck(ctx context.Context, req *pb.Empty) (*pb.HealthCheckResponse, error) {
	logger.Debugln("Received health check request")
	return &pb.HealthCheckResponse{IsHealthy: true}, nil
}

func startGrpcServer(port string) {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		logger.Fatalf("Failed to listen to %v: %v\n", port, err)
	}

	// Create a gRPC server object
	grpcServer := grpc.NewServer()

	// Register our service with the gRPC server
	pb.RegisterInternalServer(grpcServer, &protobufServer{})

	logger.Infof("Starting gRPC server on port %v", port)

	// Start the server
	if err := grpcServer.Serve(lis); err != nil {
		logger.Fatalf("Failed to serve at %v: %v", port, err)
	}

}
