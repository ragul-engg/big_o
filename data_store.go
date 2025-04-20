package main

import (
	"runtime"
	"sync/atomic"
	"unsafe"
)

/*
Flow:

	PUT:
		Request comes -> Performs encoding and then pushes to shared write queue
	GET:
		Request comes -> block the channel for the go routine handling that API response till we fetch and decode for that,
		perform gets from all across the other pods and wait till we get the data through the shared channel
*/
const BYTES_IN_GB uint64 = 2500000000

var dataStore map[string]LocationData = make(map[string]LocationData)

var totalSize uintptr

type UpdateChannelPayload struct {
	locationId     string
	encodedPayload []byte
}

func InitLoop() {
	updateChan := make(chan UpdateChannelPayload)
	updateChan <- UpdateChannelPayload{}
}

func dataStoreWriter() {
	for {
		select {
		case val, ok := <-updateChannel:
			if !ok {
				return
			}
			logger.Debugln("Starting update internally: ", val.locationId)
			updateDataStore(val.locationId, val.encodedPayload)
		}
	}
}

func listenBackgroundSyncChannel() {
	for {
		select {
		case val, ok := <-backgroundSyncChannel:
			if !ok {
				return
			}
			replicateDataGrpc(val.locationId, val.encodedPayload)
		}
	}
}

func updateDataStore(locationId string, dataShard []byte) {
	logger.Debugln("Updating data store for: ", locationId, dataShard)
	existingValue, exists := dataStore[locationId]
	if exists {
		dataStore[locationId] = LocationData{data: dataShard, modificationCount: existingValue.modificationCount + 1}
	} else {
		dataStore[locationId] = LocationData{data: dataShard, modificationCount: 1}
	}
	size := unsafe.Sizeof(dataStore)
	atomic.StoreUintptr(&totalSize, size)
}

func allowWrites() bool {
	runtime.ReadMemStats(&datastoreMemoryUsage)

	logger.Debug("Alloc:", datastoreMemoryUsage.Alloc,
		"\nTotalAlloc:", datastoreMemoryUsage.TotalAlloc,
		"\nHeapAlloc:", datastoreMemoryUsage.HeapAlloc)

	if datastoreMemoryUsage.Alloc <= BYTES_IN_GB {
		return true
	} else {
		return false
	}
}
