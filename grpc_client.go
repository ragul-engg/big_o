package main

import (
	pb "big_o/protobuf_helper"
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// var grpcClients = map[string](grpc.Client)
func updatePod(url string, upsertPayload *pb.UpsertPayload) error {
	log.Println("Update launched for: ", url)
	conn, err := grpc.NewClient(url, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Println("ERROR: establishing GRPC connection failed for: ", conn, err)
		return err
	}
	client := pb.NewInternalClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err = client.UpsertLocationData(ctx, upsertPayload)
	if err != nil {
		return err
	}
	log.Println("Update completed for: ", url, " Errors: ", err)
	return nil
}

func replicateDataGrpc(locationId string, encodedPayload [][]byte) ([]byte, error) {
	var myShare []byte
	for index, value := range encodedPayload {
		nodeIp := grpcIps[index]
		upsertPayload := constructUpsertPayload(locationId, value)
		if nodeIp != currentNodeGrpcIp {
			err := updatePod(nodeIp, &upsertPayload)
			if err != nil {
				log.Println("Something went wrong with GRPC update", err)
			}
		} else {
			log.Println("Taking my share: ", nodeIp)
			myShare = value
		}
	}
	return myShare, nil
}

func constructUpsertPayload(locationId string, encodedPayload []byte) pb.UpsertPayload {
	return pb.UpsertPayload{
		LocationId:     locationId,
		EncodedPayload: encodedPayload,
	}
}
