package main

import (
	"sync/atomic"
)

/*
Flow:

	PUT:
		Request comes -> Performs encoding and then pushes to shared write queue
	GET:
		Request comes -> block the channel for the go routine handling that API response till we fetch and decode for that,
		perform gets from all across the other pods and wait till we get the data through the shared channel
*/
const BYTES_IN_GB uintptr = 1_00_00_00_000

var totalSize uintptr

type UpdateChannelPayload struct {
	locationId     string
	encodedPayload []byte
}

func dataStoreWriter() {
	for {
		select {
		case val, ok := <-updateChannel:
			if !ok {
				return
			}
			// log.Println("Starting update internally: ", val.locationId)
			updateDataStore(val.locationId, val.encodedPayload)
		}
	}
}

func updateDataStore(locationId string, dataShard []byte) {
	// log.Println("Updating data store for: ", locationId, dataShard)
	existingValue, exists := dataStore[locationId]
	if exists {
		dataStore[locationId] = LocationData{data: dataShard, modificationCount: existingValue.modificationCount + 1}
	} else {
		dataStore[locationId] = LocationData{data: dataShard, modificationCount: 1}
	}
}

func allowWrites() bool {
	if atomic.LoadUintptr(&totalSize) <= BYTES_IN_GB {
		return true
	} else {
		return false
	}
}
