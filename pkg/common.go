package pkg

import "math"

func Max(x, y int) int {
	if x < y {
		return y
	}
	return x
}

func Min(x, y int) int {
	if x > y {
		return y
	}
	return x
}

func stddev(measurement []GPSMeasurement, fnc GPSMeasureGetFunc) float64 {
	sum := 0.0
	for _, m := range measurement {
		sum += fnc(m)
	}

	mean := sum / float64(len(measurement))

	sum = 0.0
	for _, m := range measurement {
		diff := mean - fnc(m)
		sum += diff * diff
	}

	mean = sum / float64(len(measurement))
	return math.Sqrt(mean)
}
