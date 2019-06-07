package pkg

import (
	sm "github.com/flopp/go-staticmaps"
	"github.com/fogleman/gg"
	"github.com/golang/geo/s2"
	"image/color"
)

func Plot(inputFile string, outputFile string) error {
	_, measures, err := readTrackMeasures(inputFile)
	if err != nil {
		return err
	}

	plot(measures, outputFile)
	return nil
}

func plot(measures []GPSMeasurement, outputFile string) error {
	println("Plotting your map...")
	ctx := sm.NewContext()
	ctx.SetSize(2000, 2000)

	positions := make([]s2.LatLng, len(measures))
	for i := 0; i < len(measures); i++ {
		positions[i] = s2.LatLngFromDegrees(measures[i].latLng[0], measures[i].latLng[1])
	}
	path := sm.NewPath(positions, color.RGBA{0xff, 0, 0, 0xff}, 2.0)
	ctx.AddPath(path)
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
