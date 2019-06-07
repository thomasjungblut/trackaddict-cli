package pkg

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"os"
	"time"
)

func PrintLaps(inputFile string) error {
	trackInfo, measures, err := readTrackMeasures(inputFile)
	if err != nil {
		return err
	}

	laps := extractLaps(measures, trackInfo)
	prettyPrintLaps(laps)

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
