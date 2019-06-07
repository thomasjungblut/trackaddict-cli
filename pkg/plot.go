package pkg

import (
	"fmt"
	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"image/color"
	"math/rand"
)

func Plot(measures []GPSMeasurement, laps []Lap, outputFile string, imageWidth int, imageHeight int, fastestOnly bool) error {
	fmt.Printf("Plotting your map in [Width/Height] [%d, %d]\n", imageWidth, imageHeight)
	ctx := sm.NewContext()
	ctx.SetSize(imageWidth, imageHeight)

	if fastestOnly {
		fastestIndex := 0
		min := laps[fastestIndex].timeSeconds
		for i, lap := range laps {
			if lap.timeSeconds < min {
				fastestIndex = i
				min = lap.timeSeconds
			}
		}
		laps = []Lap{laps[fastestIndex]}
		fmt.Printf("Plotting the fastest Lap [%s]\n", getLapDuration(laps[0]).String())
	}

	for lapNum := 0; lapNum < len(laps); lapNum++ {
		lapSet := MeasuresForLap(laps[lapNum], measures)
		positions := make([]s2.LatLng, len(lapSet))

		for i := 0; i < len(lapSet); i++ {
			// fmt.Printf("[%f, %f]\n", lapSet[i].latLng[0], lapSet[i].latLng[1])
			positions[i] = s2.LatLngFromDegrees(lapSet[i].latLng[0], lapSet[i].latLng[1])
		}
		lapPath := sm.NewPath(positions, color.RGBA{R: uint8(rand.Intn(255)), G: uint8(rand.Intn(255)), B: uint8(rand.Intn(255)), A: 0xff}, 2.0)
		ctx.AddPath(lapPath)
	}

	img, err := ctx.Render()
	if err != nil {
		return err
	}

	if err := gg.SavePNG(outputFile, img); err != nil {
		return err
	}

	fmt.Printf("Map saved in %s\n", outputFile)
	return nil
}
