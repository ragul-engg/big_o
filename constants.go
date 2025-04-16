package main


const NUMBER_OF_DATA_SHARDS int = 4
const NUMBER_OF_PARITY_SHARDS int = 3
const TOTAL_SHARDS int = NUMBER_OF_DATA_SHARDS + NUMBER_OF_PARITY_SHARDS
const TOTAL_NODES = TOTAL_SHARDS

const LOCATION_ID_NOT_FOUND = "location id not found"
const COULD_NOT_RECONSTRUCT_DATA = "could not reconstruct data"
const MEMORY_FULL = "memory is full"