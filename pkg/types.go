package pkg

type TrackData struct {
	Laps                   []Lap
	TrackInformation       *TrackInformation
	GPSMeasurement         []GPSMeasurement
	FilteredGPSMeasurement []GPSMeasurement
}

type Lap struct {
	timeSeconds              float64
	measureStartIndex        int
	measureEndIndexExclusive int
}

type TrackInformation struct {
	startLatLng []float64
}

type GPSMeasurement struct {
	latLng             []float64
	relativeTime       float64
	utcTimestamp       float64
	accelerationVector []float64
	altitudeMeters     float64
	speedKph           float64
	accuracyMeter      float64
	headingDegrees     float64
}
