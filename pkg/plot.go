package pkg

import (
	"fmt"
	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"image/color"
	"math/rand"
	"strings"
)

func Plot(data *TrackData, config PlotConfig) error {
	fmt.Printf("Plotting your map in [Width/Height] [%d, %d]\n", config.ImageWidth, config.ImageHeight)
	laps := data.Laps
	if config.FastestLapOnly {
		// this method is guaranteed to only have a single lap
		laps = filterFastestLap(laps)
		fmt.Printf("Plotting the fastest Lap [%s]\n", getLapDuration(laps[0]).String())
	}

	measures := data.GPSMeasurement
	if config.UseSmoothedGPSData {
		measures = data.FilteredGPSMeasurement
	}

	outputFile := config.OutputFile
	ctx := sm.NewContext()
	ctx.SetSize(config.ImageWidth, config.ImageHeight)

	for lapNum := 0; lapNum < len(laps); lapNum++ {
		addLapPathToContext(laps, lapNum, measures, ctx)
		if config.PlotLapsSeparately {
			outFileLap := fmt.Sprintf("%s_lap_%d.png", outputFile, lapNum)
			if lapNum == 0 {
				outFileLap = fmt.Sprintf("%s_outlap.png", outputFile)
			} else if lapNum == len(laps)-1 {
				outFileLap = fmt.Sprintf("%s_inlap.png", outputFile)
			}
			err := renderAndSave(ctx, outFileLap)
			if err != nil {
				return err
			}
			ctx = sm.NewContext()
			ctx.SetSize(config.ImageWidth, config.ImageHeight)
		}
	}

	if !config.PlotLapsSeparately {
		err := renderAndSave(ctx, outputFile)
		if err != nil {
			return err
		}
	}
	return nil
}

func renderAndSave(ctx *sm.Context, outputFile string) error {
	img, err := ctx.Render()
	if err != nil {
		return err
	}

	if !strings.HasSuffix(strings.ToLower(outputFile), ".png") {
		outputFile = outputFile + ".png"
	}

	if err := gg.SavePNG(outputFile, img); err != nil {
		return err
	}

	fmt.Printf("Saved %s\n", outputFile)
	return nil
}

func addLapPathToContext(laps []Lap, lapNum int, measures []GPSMeasurement, ctx *sm.Context) {
	lapSet := MeasuresForLap(laps[lapNum], measures)
	positions := make([]s2.LatLng, len(lapSet))
	for i := 0; i < len(lapSet); i++ {
		// fmt.Printf("[%f, %f]\n", lapSet[i].latLng[0], lapSet[i].latLng[1])
		positions[i] = s2.LatLngFromDegrees(lapSet[i].latLng[0], lapSet[i].latLng[1])
	}
	lapPath := sm.NewPath(positions, color.RGBA{R: uint8(rand.Intn(255)), G: uint8(rand.Intn(255)), B: uint8(rand.Intn(255)), A: 0xff}, 2.0)
	ctx.AddPath(lapPath)
}

func filterFastestLap(laps []Lap) []Lap {
	fastestIndex := 0
	min := laps[fastestIndex].timeSeconds
	for i, lap := range laps {
		if lap.timeSeconds < min {
			fastestIndex = i
			min = lap.timeSeconds
		}
	}
	laps = []Lap{laps[fastestIndex]}
	return laps
}
