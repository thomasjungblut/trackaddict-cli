package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thomasjungblut/trackaddict-cli/pkg"
	"log"
	"os"
)

var (
	InputFile          string
	OutputFile         string
	PlotImageWidth     int
	PlotImageHeight    int
	PlotFastestLapOnly bool
	PlotLapsSeparately bool
	FilteringEnabled   bool
	RecalculateLaps    bool
)

var rootCmd = &cobra.Command{
	Use:   "trackaddict-cli",
	Short: "This commandline tool will try to fix your laps from noisy GPS and print the lap times.",
}

var lapCmd = &cobra.Command{
	Use:   "laps",
	Short: "Prints your lap times",
	Run: func(cmd *cobra.Command, args []string) {
		dataConfig := pkg.DataConfig{InputFile: InputFile, UseSmoothedGPSData: FilteringEnabled, RecalculateLaps: RecalculateLaps}
		data, err := pkg.ReadData(dataConfig)
		if err != nil {
			log.Fatalf("encountered an error: %v", err)
		}
		
		pkg.PrettyPrintLaps(data.Laps)
	},
}

var plotCmd = &cobra.Command{
	Use:   "plot",
	Short: "Plots a small map of your GPS coordinates",
	Run: func(cmd *cobra.Command, args []string) {
		dataConfig := pkg.DataConfig{InputFile: InputFile, UseSmoothedGPSData: FilteringEnabled, RecalculateLaps: RecalculateLaps}
		data, err := pkg.ReadData(dataConfig)
		if err != nil {
			log.Fatalf("encountered an error: %v", err)
		}

		config := pkg.PlotConfig{
			DataConfig:         dataConfig,
			OutputFile:         OutputFile,
			ImageWidth:         PlotImageWidth,
			ImageHeight:        PlotImageHeight,
			PlotLapsSeparately: PlotLapsSeparately,
			FastestLapOnly:     PlotFastestLapOnly,
		}

		err = pkg.Plot(data, config)
		if err != nil {
			log.Fatalf("encountered an error: %v", err)
		}
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of trackaddict-cli",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("0.0.1")
	},
}

func init() {
	lapCmd.Flags().StringVarP(&InputFile, "inputFile", "i", "", "Input File (required)")
	_ = lapCmd.MarkFlagRequired("inputFile")
	lapCmd.Flags().BoolVarP(&RecalculateLaps, "fix-laps", "", false, "If set, it will heuristically recalculate the laps")
	lapCmd.Flags().BoolVarP(&FilteringEnabled, "smooth", "", false, "If set, it will try to smooth the GPS data with accelerometer information")

	plotCmd.Flags().StringVarP(&InputFile, "inputFile", "i", "", "Input File (required)")
	_ = plotCmd.MarkFlagRequired("inputFile")
	plotCmd.Flags().StringVarP(&OutputFile, "outputFile", "o", "", "Output File (png)")
	_ = plotCmd.MarkFlagRequired("outputFile")

	plotCmd.Flags().IntVarP(&PlotImageWidth, "width", "", 2000, "Width of the output image, 2000px default")
	plotCmd.Flags().IntVarP(&PlotImageHeight, "height", "", 2000, "Height of the output image, 2000px default")
	plotCmd.Flags().BoolVarP(&PlotFastestLapOnly, "fastest-lap-only", "", false, "If set, it plots only the fastest lap")
	plotCmd.Flags().BoolVarP(&PlotLapsSeparately, "plot-each-lap", "", false, "If set, it will plot each lap in its own file by appending the lap number to the given outputfile name.")
	plotCmd.Flags().BoolVarP(&RecalculateLaps, "fix-laps", "", false, "If set, it will heuristically recalculate the laps")
	plotCmd.Flags().BoolVarP(&FilteringEnabled, "smooth", "", false, "If set, it will try to smooth the GPS location by kalman filtering using accelerometer data")

	rootCmd.AddCommand(lapCmd)
	rootCmd.AddCommand(plotCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
