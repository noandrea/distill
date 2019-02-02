// Copyright © 2018 NAME HERE <EMAIL ADDRESS>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cmd

import (
	"os"
	"path/filepath"

	"github.com/noandrea/distill/urlstore"

	"github.com/jbrodriguez/mlog"
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
	RootCmd.AddCommand(importCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// importCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	importCmd.Flags().StringVarP(&csvFile, "csv-file", "f", "urls.csv", "Path to the csv to import")

}

func importCsv(cmd *cobra.Command, args []string) {
	urlstore.NewSession()
	defer urlstore.CloseSession()
	abp, err := filepath.Abs(csvFile)
	if err != nil {
		mlog.Fatalf("Invalid path %s: %v", csvFile, err)
	}
	if _, err = os.Stat(abp); os.IsNotExist(err) {
		mlog.Fatalf("Invalid path %s: %v", csvFile, err)
	}
	if rows, err := urlstore.ImportCSV(abp); err != nil {
		mlog.Fatalf("Error create backup at %s: %v", csvFile, err)
	} else {
		mlog.Info("Import complete, %d url record loaded", rows)
	}
}
