package pkg

import (
	"fmt"
	"math"
)

const EarthRadiusInMeters = 6372797.560856

func degreesToRadians(degrees float64) float64 {
	return float64(degrees * math.Pi / 180.0)
}

func radiansToDegrees(radians float64) float64 {
	return float64(radians * 180.0 / math.Pi)
}

func getPointAhead(latLng []float64, distanceMeters float64, azimuth float64) []float64 {
	radiusFraction := float64(distanceMeters / EarthRadiusInMeters)

	bearing := float64(degreesToRadians(azimuth))

	lat1 := degreesToRadians(latLng[0])
	lng1 := degreesToRadians(latLng[1])

	lat2Part1 := math.Sin(lat1) * math.Cos(radiusFraction)
	lat2Part2 := math.Cos(lat1) * math.Sin(radiusFraction) * math.Cos(bearing)

	lat2 := math.Asin(lat2Part1 + lat2Part2)

	lng2Part1 := math.Sin(bearing) * math.Sin(radiusFraction) * math.Cos(lat1)
	lng2Part2 := math.Cos(radiusFraction) - (math.Sin(lat1) * math.Sin(lat2))

	lng2 := lng1 + math.Atan2(lng2Part1, lng2Part2)
	lng2 = math.Mod(lng2+3*math.Pi, 2*math.Pi) - math.Pi

	return []float64{radiansToDegrees(lat2), radiansToDegrees(lng2),}
}

func pointPlusDistanceEast(fromCoordinate []float64, distance float64) []float64 {
	return getPointAhead(fromCoordinate, distance, 90.0)
}

func pointPlusDistanceNorth(fromCoordinate []float64, distance float64) []float64 {
	return getPointAhead(fromCoordinate, distance, 0.0)
}

func metersToGeoPoint(latAsMeters float64, lonAsMeters float64) []float64 {
	point := []float64{0.0, 0.0}
	pointEast := pointPlusDistanceEast(point, lonAsMeters)
	pointNorthEast := pointPlusDistanceNorth(pointEast, latAsMeters)
	return pointNorthEast
}

func latToMeter(latitude float64) float64 {
	distance := haversineDistance([]float64{latitude, 0.0}, []float64{0.0, 0.0})
	if latitude < 0 {
		distance *= -1
	}
	return distance
}

func lngToMeter(longitude float64) float64 {
	distance := haversineDistance([]float64{longitude, 0.0}, []float64{0.0, 0.0})
	if longitude < 0 {
		distance *= -1
	}
	return distance
}

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

func PredictKalmanFilteredMeasures(measurement []GPSMeasurement) []GPSMeasurement {
	init := measurement[0]

	latFilter := NewKalmanFilterFusedPositionAccelerometer(latToMeter(init.latLng[0]), 0, 0, init.relativeTime)
	lngFilter := NewKalmanFilterFusedPositionAccelerometer(lngToMeter(init.latLng[1]), 0, 0, init.relativeTime)

	var output []GPSMeasurement
	for i := 1; i < len(measurement); i++ {
		data := measurement[i]
		speed := data.speedKph / 3.6
		xVel := speed * math.Cos(data.headingDegrees)
		yVel := speed * math.Sin(data.headingDegrees)

		latFilter.Update(latToMeter(data.latLng[0]), xVel, nil, 0)
		lngFilter.Update(lngToMeter(data.latLng[1]), yVel, nil, 0)

		latFilter.Predict(data.accelerationVector[0], data.relativeTime)
		lngFilter.Predict(data.accelerationVector[1], data.relativeTime)

		point := metersToGeoPoint(latFilter.GetPredictedPosition(), lngFilter.GetPredictedPosition())
		fmt.Printf("[%f] vs. [%f]\n", data.latLng, point)
		output = append(output, GPSMeasurement{
			latLng:             point,
			altitudeMeters:     data.altitudeMeters,
			relativeTime:       data.relativeTime,
			accelerationVector: data.accelerationVector,
			speedKph:           data.speedKph,
			utcTimestamp:       data.utcTimestamp,
		})
	}

	return output
}
