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

func extractLaps(measures []GPSMeasurement, trackInfo *TrackInformation) []Lap {
	var laps []Lap
	currentLap := Lap{measureStartIndex: 0}
	for i := 0; i < len(measures); i++ {
		// fmt.Printf("i=%d reltime=%f timeSeconds=%f dist=%f\n", index, measure.relativeTime, measure.utcTimestamp, dist)
		measure := measures[i]
		dist := haversineDistance(trackInfo.startLatLng, measure.latLng)
		// simple thresholding algorithm with cooldown
		if dist < DIST_TOLERANCE_IN_METERS && (i-currentLap.measureStartIndex) > NUM_LAP_COOLDOWN_MEASURES {
			currentLap.measureEndIndexExclusive = i + 1
			currentLap.timeSeconds = measure.relativeTime - measures[currentLap.measureStartIndex].relativeTime
			laps = append(laps, currentLap)
			currentLap = Lap{measureStartIndex: currentLap.measureEndIndexExclusive}
		}
	}

	// finish the outlap
	currentLap.measureEndIndexExclusive = len(measures)
	currentLap.timeSeconds = measures[len(measures)-1].relativeTime - measures[currentLap.measureStartIndex].relativeTime
	laps = append(laps, currentLap)

	return laps
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
				relativeTime: mustParseFloat64(split[0]),
				utcTimestamp: mustParseFloat64(split[1]),
				latLng:       []float64{mustParseFloat64(split[7]), mustParseFloat64(split[8])},
				accelXYZ:     []float64{mustParseFloat64(split[14]), mustParseFloat64(split[15]), mustParseFloat64(split[16])},
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
