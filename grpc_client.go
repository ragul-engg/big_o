package main

import (
	connectionPool "big_o/connection_pool"
	pb "big_o/protobuf_helper"
	"context"
	"errors"
	"time"
)

func updatePod(url string, upsertPayload *pb.UpsertPayload) error {
	logger.Debugln("Update launched for: ", url)
	client, err := connectionPool.GetClientFor(url)

	if err != nil {
		logger.Errorln("ERROR: client not found for: ", url, " Error: ", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err = client.UpsertLocationData(ctx, upsertPayload)
	if err != nil {
		logger.Errorln("Error: Upsert failed.", err)
		return err
	}
	logger.Debugln("Update completed for: ", url)
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
				logger.Errorf("GRPC update to %s failed with error: %s\n", nodeIp, err)
				return nil, err
			}
		} else {
			logger.Debugln("Taking my share: ", nodeIp)
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

func healthCheck(url string) error {
	client, err := connectionPool.GetClientFor(url)
	if err != nil {
		logger.Errorln("ERROR: client not found for: ", url, " Error: ", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	res, err := client.HealthCheck(ctx, &pb.Empty{})
	
	if err != nil {
		logger.Errorln("Error: Health check failed.", err)
		return err
	}

	if(res.IsHealthy) {
		logger.Debugln("Health check passed for: ", url)
		return nil
	} else {
		logger.Errorln("Health check failed for: ", url)
		return errors.New("health check failed for " + url)
	}
}
