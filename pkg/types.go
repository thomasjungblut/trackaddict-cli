package pkg

type Lap struct {
	timeSeconds              float64
	measureStartIndex        int
	measureEndIndexExclusive int
}

type TrackInformation struct {
	startLatLng []float64
}

type GPSMeasurement struct {
	latLng       []float64
	relativeTime float64
	utcTimestamp float64
	accelXYZ     []float64
}
