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

var (
	Red         = color.RGBA{R: uint8(255), G: uint8(0), B: uint8(0), A: 0xff}
	Black       = color.RGBA{R: uint8(0), G: uint8(0), B: uint8(0), A: 0xff}
	Transparent = color.RGBA{R: uint8(0), G: uint8(0), B: uint8(0), A: 0}
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

	gpsErrorStdDevMeters := stddev(measures,
		func(measurement GPSMeasurement) float64 {
			return measurement.accuracyMeter
		})

	outputFile := config.OutputFile
	pathColor := Black
	ctx := newPlotContext(config, "")
	for lapNum := 0; lapNum < len(laps); lapNum++ {
		if config.PlotLapsSeparately {
			ctx = newPlotContext(config, fmt.Sprintf("Lap Time: %s", getLapDuration(laps[lapNum]).String()))
		} else {
			pathColor = color.RGBA{R: uint8(rand.Intn(255)), G: uint8(rand.Intn(255)), B: uint8(rand.Intn(255)), A: 0xff}
		}

		addStartEndZone(ctx, data.TrackInformation.startLatLng, gpsErrorStdDevMeters)
		addLapPathToContext(laps[lapNum], measures, ctx, pathColor)

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

func addLapPathToContext(lap Lap, measures []GPSMeasurement, ctx *sm.Context, color color.RGBA) {
	lapSet := MeasuresForLap(lap, measures)
	positions := make([]s2.LatLng, len(lapSet))
	for i := 0; i < len(lapSet); i++ {
		// fmt.Printf("[%f, %f]\n", lapSet[i].latLng[0], lapSet[i].latLng[1])
		positions[i] = s2LatLngFromSlice(lapSet[i].latLng)
	}
	lapPath := sm.NewPath(positions, color, 2.0)
	ctx.AddPath(lapPath)
}

func addStartEndZone(ctx *sm.Context, startLatLng []float64, radius float64) {
	ctx.AddCircle(&sm.Circle{
		Position: s2LatLngFromSlice(startLatLng),
		Radius:   radius,
		Color:    Red,
		Fill:     Transparent,
		Weight:   2})
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

func s2LatLngFromSlice(slice []float64) s2.LatLng {
	return s2.LatLngFromDegrees(slice[0], slice[1])
}

func newPlotContext(config PlotConfig, attributionHackString string) *sm.Context {
	ctx := sm.NewContext()
	ctx.SetSize(config.ImageWidth, config.ImageHeight)
	provider := sm.NewTileProviderOpenStreetMaps()
	provider.Attribution = fmt.Sprintf("%s | %s", provider.Attribution, attributionHackString)
	ctx.SetTileProvider(provider)
	return ctx
}
