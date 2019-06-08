package pkg

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"log"
	"math"
	"os"
	"time"
)

const DistToleranceInMeters = 10
const NumLapCooldownMeasures = 1000

func MeasuresForLap(lap Lap, measures []GPSMeasurement) []GPSMeasurement {
	return measures[lap.measureStartIndex:lap.measureEndIndexExclusive]
}

func extractLaps(config DataConfig, data *TrackData) []Lap {
	trackInfo := data.TrackInformation
	measures := data.GPSMeasurement
	if config.UseSmoothedGPSData {
		measures = data.FilteredGPSMeasurement
	}

	if config.RecalculateLaps {
		return calculateLapsWithThresholding(measures, trackInfo)
	} else {
		numLaps := 0
		for _, measure := range measures {
			numLaps = Max(numLaps, measure.trackAddictLap)
		}

		// +1 since trackaddict laps are zero indexed
		laps := make([]Lap, numLaps+1)
		for i := range laps {
			laps[i].measureStartIndex = math.MaxInt32
		}
		for i, measure := range measures {
			l := laps[measure.trackAddictLap]
			laps[measure.trackAddictLap].measureStartIndex = Min(l.measureStartIndex, i)
			laps[measure.trackAddictLap].measureEndIndexExclusive = Max(l.measureEndIndexExclusive, i)
		}

		// now just fill the lap times with the indices
		for i, lap := range laps {
			laps[i].timeSeconds = measures[lap.measureEndIndexExclusive-1].relativeTime - measures[lap.measureStartIndex].relativeTime
		}

		return laps
	}
}

func calculateLapsWithThresholding(measures []GPSMeasurement, trackInfo *TrackInformation) []Lap {
	var laps []Lap
	currentLap := Lap{measureStartIndex: 0}
	for i := 0; i < len(measures); i++ {
		measure := measures[i]
		dist := haversineDistance(trackInfo.startLatLng, measure.latLng)
		// fmt.Printf("%f\t%f\n", measure.relativeTime, dist)
		// simple thresholding algorithm with some cooldown period of measurements
		if dist < DistToleranceInMeters && (i-currentLap.measureStartIndex) > NumLapCooldownMeasures {
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

func PrettyPrintLaps(laps []Lap) {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Lap Number", "Time (s)", "Measure Range"})

	for i, v := range laps {
		duration := getLapDuration(v)

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

func getLapDuration(v Lap) time.Duration {
	duration, err := time.ParseDuration(fmt.Sprintf("%fs", v.timeSeconds))
	if err != nil {
		log.Fatalf("encountered an error while parsing laptime %d, error was: %v", v.timeSeconds, err)
	}
	return duration
}
