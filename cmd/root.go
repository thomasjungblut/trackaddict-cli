package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/thomasjungblut/trackaddict-cli/pkg"
	"os"
)

var (
	InputFile string
	PlotFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "trackaddict-cli",
	Short: "This commandline tool will try to fix your laps from noisy GPS and print the laptimes.",
}

var createCmd = &cobra.Command{
	Use:   "fix",
	Short: "fixes the given file",
	Run: func(cmd *cobra.Command, args []string) {
		err := pkg.Fix(InputFile, PlotFlag)
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
	createCmd.Flags().StringVarP(&InputFile, "inputFile", "f", "", "Input File (required)")
	createCmd.MarkFlagRequired("inputFile")
	createCmd.Flags().BoolVarP(&PlotFlag, "plot", "p", false, "Plot your track")

	rootCmd.AddCommand(createCmd)
	rootCmd.AddCommand(versionCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
