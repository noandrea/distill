package cmd

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
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
	urlstore.NewSession(settings)
	defer urlstore.CloseSession()
	abp, err := filepath.Abs(csvFile)
	if err != nil {
		log.Fatalf("Invalid path %s: %v", csvFile, err)
	}
	if _, err = os.Stat(abp); os.IsNotExist(err) {
		log.Fatalf("Invalid path %s: %v", csvFile, err)
	}
	if rows, err := urlstore.ImportCSV(abp); err != nil {
		log.Fatalf("Error create backup at %s: %v", csvFile, err)
	} else {
		log.Info("Import complete, ", rows, " url record loaded")
	}
}
