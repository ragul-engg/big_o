package main

import (
	connectionPool "big_o/connection_pool"
	pb "big_o/protobuf_helper"
	"context"
	"log"
	"sync"
	"time"
)

func updatePod(url string, upsertPayload *pb.UpsertPayload) error {
	log.Println("Update launched for: ", url)
	client, err := connectionPool.GetClientFor(url)

	if err != nil {
		log.Println("ERROR: client not found for: ", url, " Error: ", err)
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err = client.UpsertLocationData(ctx, upsertPayload)
	if err != nil {
		log.Println("Error: Upsert failed.", err)
		return err
	}
	log.Println("Update completed for: ", url)
	return nil
}

func replicateDataGrpc(locationId string, encodedPayload [][]byte) ([]byte, error) {
	var myShare []byte
	var wg sync.WaitGroup // WaitGroup to wait for all goroutines to complete
	errCh := make(chan error, len(encodedPayload)) // Channel to capture errors from goroutines

	// Iterate through the encodedPayload and create a goroutine for each update
	for index, value := range encodedPayload {
		nodeIp := grpcIps[index]
		upsertPayload := constructUpsertPayload(locationId, value)
		wg.Add(1) // Increment the counter for each goroutine

		// Launch a goroutine to update the pod in parallel
		go func(nodeIp string, upsertPayload pb.UpsertPayload) {
			defer wg.Done() // Decrement the counter when the goroutine completes

			if nodeIp != currentNodeGrpcIp {
				// Perform the gRPC call if not the current node
				err := updatePod(nodeIp, &upsertPayload)
				if err != nil {
					log.Println("Something went wrong with GRPC update", err)
					errCh <- err // Send error to the channel
				}
			} else {
				log.Println("Taking my share: ", nodeIp)
				// If it's the current node, store the share
				myShare = value
			}
		}(nodeIp, upsertPayload)
	}

	// Wait for all the goroutines to finish
	wg.Wait()
	close(errCh) // Close the error channel after all goroutines are done

	// Check if any goroutines reported an error
	for err := range errCh {
		if err != nil {
			return nil, err // Return the first encountered error
		}
	}

	return myShare, nil // Return the result after all parallel calls are completed
}

func constructUpsertPayload(locationId string, encodedPayload []byte) pb.UpsertPayload {
	return pb.UpsertPayload{
		LocationId:     locationId,
		EncodedPayload: encodedPayload,
	}
}
