package pkg

import (
	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"image/color"
	"math/rand"
)

func Plot(inputFile string, outputFile string) error {
	trackInfo, measures, err := readTrackMeasures(inputFile)
	if err != nil {
		return err
	}

	laps := extractLaps(measures, trackInfo)

	return plot(measures, laps, outputFile)
}

func plot(measures []GPSMeasurement, laps []Lap, outputFile string) error {
	println("Plotting your map...")
	ctx := sm.NewContext()
	ctx.SetSize(2000, 2000)

		for lapNum := 0; lapNum < len(laps); lapNum++ {
			lapSet := measures[laps[lapNum].measureStartIndex : laps[lapNum].measureEndIndexExclusive]
			positions := make([]s2.LatLng, len(lapSet))

			for i := 0; i < len(lapSet); i++ {
				positions[i] = s2.LatLngFromDegrees(lapSet[i].latLng[0], lapSet[i].latLng[1])
			}
			lapPath := sm.NewPath(positions, color.RGBA{uint8(rand.Intn(255)), uint8(rand.Intn(255)), uint8(rand.Intn(255)), 0xff}, 2.0)
			ctx.AddPath(lapPath)
	}


	img, err := ctx.Render()
	if err != nil {
		return err
	}

	if err := gg.SavePNG(outputFile, img); err != nil {
		return err
	}

	println("Map plotted to: " + outputFile)
	return nil
}
