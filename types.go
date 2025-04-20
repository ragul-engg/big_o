package main

type Payload struct {
	Id               string  `json:"id"`
	Seismic_activity float32 `json:"seismic_activity"`
	Temperature_c    float32 `json:"temperature_c"`
	Radiation_level  float32 `json:"radiation_level"`
}
type LocationData struct {
	data              []byte
	modificationCount int
}

type ResponsePayload struct {
	Payload
	ModificationCount int `json:"modification_count"`
}

type BackgroundSyncPayload struct {
	locationId     string
	encodedPayload [][]byte
}
