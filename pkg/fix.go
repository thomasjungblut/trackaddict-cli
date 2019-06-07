package pkg

import (
	"bufio"
	"errors"
	"fmt"
	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"github.com/olekukonko/tablewriter"
	"image/color"
	"log"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

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

func Fix(inputFile string, plotFlag bool) error {
	trackInfo, measures, err := readTrackMeasures(inputFile)
	if err != nil {
		return err
	}

	laps := extractLaps(measures, trackInfo)
	prettyPrintLaps(laps)
	if plotFlag {
		plot(measures)
	}

	return nil
}

func prettyPrintLaps(laps []Lap) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Lap Number", "Time (s)", "Measure Range"})

	for i, v := range laps {
		duration, _ := time.ParseDuration(fmt.Sprintf("%fs", v.timeSeconds))

		lapFormat := fmt.Sprintf("%d", i+1)
		if i == 0 {
			lapFormat = fmt.Sprintf("%d (Outlap)", i+1)
		} else if i == len(laps)-1 {
			lapFormat = fmt.Sprintf("%d (Inlap)", i+1)
		}

		table.Append([]string{
			lapFormat,
			duration.String(),
			fmt.Sprintf("%d-%d", v.measureStartIndex, v.measureEndIndexExclusive),
		})
	}
	table.Render()
}

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

func plot(measures []GPSMeasurement) {
	println("Plotting your map...")
	ctx := sm.NewContext()
	ctx.SetSize(2000,2000)

	positions := make([]s2.LatLng, len(measures))
	for i:= 0; i < len(measures); i++ {
		positions[i] = s2.LatLngFromDegrees(measures[i].latLng[0], measures[i].latLng[1])
	}
	path := sm.NewPath(positions, color.RGBA{0xff, 0, 0, 0xff}, 2.0)
	ctx.AddPath(path)
	img, err := ctx.Render()
	if err != nil {
		panic(err)
	}

	if err := gg.SavePNG("output.png", img); err != nil {
		panic(err)
	}

	println("Map plotted!")
}