package pkg

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
)

func ReadData(inputFile string) (*TrackData, error) {
	trackInfo, measures, err := readTrackMeasures(inputFile)
	if err != nil {
		return nil, err
	}

	laps := extractLaps(measures, trackInfo)
	return &TrackData{TrackInformation: trackInfo, GPSMeasurement: measures, Laps: laps}, nil
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
				fmt.Printf("Found Start/End GPS coordinate: [%f/%f]\n", trackInfo.startLatLng[0], trackInfo.startLatLng[1])
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
			})

		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	// TODO flag
	measures = PredictKalmanFilteredMeasures(measures)

	return &trackInfo, measures, nil
}

func mustParseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalf("can't parse float: %s", s)
	}
	return f
}
