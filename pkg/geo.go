package pkg

import "math"

var EarthRadiusInMeters = 6372797.560856

func haversineDistance(aInit []float64, bInit []float64) float64 {
	a := make([]float64, len(aInit))
	copy(a, aInit)
	b := make([]float64, len(bInit))
	copy(b, bInit)

	a[0] = a[0] / 180.0 * math.Pi
	a[1] = a[1] / 180.0 * math.Pi
	b[0] = b[0] / 180.0 * math.Pi
	b[1] = b[1] / 180.0 * math.Pi

	return math.Acos(math.Sin(a[0])*math.Sin(b[0])+math.Cos(a[0])*math.Cos(b[0])*math.Cos(a[1]-b[1])) * EarthRadiusInMeters
}
