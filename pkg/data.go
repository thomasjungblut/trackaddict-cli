package pkg

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func ReadData(config DataConfig) (*TrackData, error) {
	trackInfo, measures, err := readTrackMeasures(config.InputFile)
	if err != nil {
		return nil, err
	}

	filteredMeasures := PredictKalmanFilteredMeasures(measures)
	data := &TrackData{TrackInformation: trackInfo, GPSMeasurement: measures, FilteredGPSMeasurement: filteredMeasures}
	laps := extractLaps(config, data)
	data.Laps = laps
	return data, nil
}

func readTrackMeasures(inputFile string) (*TrackInformation, []GPSMeasurement, error) {

	file, err := os.Open(inputFile)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()

	trackInfo := TrackInformation{}
	var measures []GPSMeasurement
	scanner := bufio.NewScanner(file)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "# End Point") {
				r := regexp.MustCompile(`# End Point: ([0-9\.]+), ([0-9\.]+).*`)

				matches := r.FindAllStringSubmatch(line, -1)
				if matches == nil || len(matches) != 1 || len(matches[0]) != 3 {
					return nil, nil, errors.New("can't parse end point lat/lng")
				}

				trackInfo.startLatLng = []float64{mustParseFloat64(matches[0][1]), mustParseFloat64(matches[0][2])}
				// fmt.Printf("Found Start/End GPS coordinate: [%f/%f]\n", trackInfo.startLatLng[0], trackInfo.startLatLng[1])
			}
		} else if strings.HasPrefix(line, "\"Time\"") {
			// skip the header
		} else {
			split := strings.Split(line, ",")
			if len(split) != 20 {
				return nil, nil, fmt.Errorf("not enough columns in line %d", lineCount)
			}

			measures = append(measures, GPSMeasurement{
				relativeTime:       mustParseFloat64(split[0]),
				utcTimestamp:       mustParseFloat64(split[1]),
				latLng:             []float64{mustParseFloat64(split[7]), mustParseFloat64(split[8])},
				altitudeMeters:     mustParseFloat64(split[9]),
				speedKph:           mustParseFloat64(split[11]),
				headingDegrees:     mustParseFloat64(split[12]),
				accuracyMeter:      mustParseFloat64(split[13]),
				accelerationVector: []float64{mustParseFloat64(split[14]), mustParseFloat64(split[15]), mustParseFloat64(split[16])},
				trackAddictLap:     mustParseInt(split[2]),
			})

		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return &trackInfo, measures, nil
}

func PredictKalmanFilteredMeasures(measurement []GPSMeasurement) []GPSMeasurement {
	init := measurement[0]

	latFilter := NewKalmanFilterFusedPositionAccelerometer(latToMeter(init.latLng[0]), 10, 0.5, init.utcTimestamp)
	lngFilter := NewKalmanFilterFusedPositionAccelerometer(lngToMeter(init.latLng[1]), 10, 0.5, init.utcTimestamp)

	var output []GPSMeasurement
	for i := 1; i < len(measurement); i++ {
		data := measurement[i]

		speedMetersPerSecond := data.speedKph / 3.6
		xVel := speedMetersPerSecond * math.Cos(data.headingDegrees)
		yVel := speedMetersPerSecond * math.Sin(data.headingDegrees)

		latFilter.Update(latToMeter(data.latLng[0]), xVel, &data.accuracyMeter, 0)
		lngFilter.Update(lngToMeter(data.latLng[1]), yVel, &data.accuracyMeter, 0)

		latFilter.Predict(data.accelerationVector[0], init.utcTimestamp)
		lngFilter.Predict(data.accelerationVector[1], init.utcTimestamp)

		point := metersToGeoPoint(latFilter.GetPredictedPosition(), lngFilter.GetPredictedPosition())
		//fmt.Printf("[%f] vs. [%f]\n", data.latLng, point)
		output = append(output, GPSMeasurement{
			latLng:             point,
			altitudeMeters:     data.altitudeMeters,
			relativeTime:       data.relativeTime,
			accelerationVector: data.accelerationVector,
			speedKph:           data.speedKph,
			utcTimestamp:       data.utcTimestamp,
			trackAddictLap:     data.trackAddictLap,
			accuracyMeter:      data.accuracyMeter,
			headingDegrees:     data.headingDegrees,
		})
	}

	return output
}

func mustParseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalf("can't parse float: %s", s)
	}
	return f
}

func mustParseInt(s string) int {
	i, err := strconv.ParseInt(s, 10, 32)
	if err != nil {
		log.Fatalf("can't parse int: %s", s)
	}
	return int(i)
}
