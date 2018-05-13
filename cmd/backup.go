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

	"github.com/jbrodriguez/mlog"
	"gitlab.com/welance/distill/internal/distill"

	"github.com/spf13/cobra"
)

var backupFile string

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup the database urls",
	Long: `Create a backup of the urls database.
  The backup format can be binary or csv, the format will be 
  selected by the extension of the backup file (.bin for binary and .csv for csv).
  The backup command will try to create the output file and all the intermediate folders.
  
  The backup cannot be executed in a live service`,
	Run: backup,
}

func init() {
	RootCmd.AddCommand(backupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	backupCmd.Flags().StringVarP(&backupFile, "backup-file", "f", "ilij.backu.bin", "Output file for backup")

}

func backup(cmd *cobra.Command, args []string) {
	distill.NewSession()
	defer distill.CloseSession()
	abp, err := filepath.Abs(backupFile)
	if err != nil {
		mlog.Fatalf("Invalid path %s: %v", backupFile, err)
	}
	err = os.MkdirAll(filepath.Dir(abp), os.ModeDir)
	if err != nil {
		mlog.Fatalf("Error create backup path to %s: %v", backupFile, err)
	}
	if err = distill.Backup(abp); err != nil {
		mlog.Fatalf("Error create backup at %s: %v", backupFile, err)
	}
}
