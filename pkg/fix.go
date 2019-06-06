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

var EARTH_RADIUS_IN_METERS = 6372797.560856
var DIST_TOLERANCE = 0.05

type TrackInformation struct {
	startLatLng []float64
}

type GPSMeasurement struct {
	latLng       []float64
	utcTimestamp float64
}

func Fix(inputFile string) error {
	trackInfo, measures, err := readTrackMeasures(inputFile)
	if err != nil {
		return err
	}

	for index, measure := range measures {
		fmt.Printf("i=%d time=%f dist=%f\n", index,
			measure.utcTimestamp, haversineDistance(trackInfo.startLatLng, measure.latLng))
	}

	return nil
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
				utcTimestamp: mustParseFloat64(split[1]),
				latLng:       []float64{mustParseFloat64(split[7]), mustParseFloat64(split[8])},
			})

		}
		lineCount++
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, err
	}

	return &trackInfo, measures, nil
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

	return math.Acos(math.Sin(a[0])*math.Sin(b[0])+math.Cos(a[0])*math.Cos(b[0])*math.Cos(a[1]-b[1])) * EARTH_RADIUS_IN_METERS
}

func mustParseFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalf("can't parse float: %s", s)
	}
	return f
}
