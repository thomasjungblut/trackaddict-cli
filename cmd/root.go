package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thomasjungblut/trackaddict-cli/pkg"
	"os"
)

var (
	InputFile          string
	OutputFile         string
	PlotImageWidth     int
	PlotImageHeight    int
	PlotFastestLapOnly bool
)

var rootCmd = &cobra.Command{
	Use:   "trackaddict-cli",
	Short: "This commandline tool will try to fix your laps from noisy GPS and print the lap times.",
}

var createCmd = &cobra.Command{
	Use:   "laps",
	Short: "Prints your lap times",
	Run: func(cmd *cobra.Command, args []string) {
		_, _, laps, err := pkg.ReadData(InputFile)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		pkg.PrettyPrintLaps(laps)
	},
}

var plotCmd = &cobra.Command{
	Use:   "plot",
	Short: "Plots a small map of your GPS coordinates",
	Run: func(cmd *cobra.Command, args []string) {
		_, measures, laps, err := pkg.ReadData(InputFile)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}

		err = pkg.Plot(measures, laps, OutputFile, PlotImageWidth, PlotImageHeight, PlotFastestLapOnly)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
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
	createCmd.Flags().StringVarP(&InputFile, "inputFile", "i", "", "Input File (required)")
	_ = createCmd.MarkFlagRequired("inputFile")

	plotCmd.Flags().StringVarP(&InputFile, "inputFile", "i", "", "Input File (required)")
	_ = plotCmd.MarkFlagRequired("inputFile")
	plotCmd.Flags().StringVarP(&OutputFile, "outputFile", "o", "", "Output File (with .png)")
	_ = plotCmd.MarkFlagRequired("outputFile")

	plotCmd.Flags().IntVarP(&PlotImageWidth, "width", "", 2000, "width of the output image, 2000px default")
	plotCmd.Flags().IntVarP(&PlotImageHeight, "height", "", 2000, "height of the output image, 2000px default")
	plotCmd.Flags().BoolVarP(&PlotFastestLapOnly, "fastest-lap-only", "", false, "if set, it plots only the fastest lap")

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(plotCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
