package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import data from csv file",
	Long:  `The csv file can be with or without header`,
	Run:   importCsv,
}

var csvFile string

func init() {
	rootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	importCmd.Flags().StringVarP(&csvFile, "csv-file", "f", "urls.csv", "Path to the csv to import")

}

func importCsv(cmd *cobra.Command, args []string) {
	fmt.Printf("not yet implemented")
}
