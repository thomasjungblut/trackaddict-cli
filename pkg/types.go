package pkg

var EARTH_RADIUS_IN_METERS = 6372797.560856
var DIST_TOLERANCE_IN_METERS = 30.0
var NUM_LAP_COOLDOWN_MEASURES = 100

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
	accelXYZ	 []float64
}
