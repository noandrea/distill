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
	"github.com/spf13/cobra"
	"gitlab.com/welance/oss/distill/internal/distill"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup",
	Long:  `The backup has to have been created with the backup command`,
	Run:   restore,
}

func init() {
	RootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVarP(&backupFile, "backup-file", "f", "ilij.backup.bin", "Input for restore")
}

func restore(cmd *cobra.Command, args []string) {
	distill.NewSession()
	defer distill.CloseSession()
	abp, err := filepath.Abs(backupFile)
	if err != nil {
		mlog.Fatalf("Invalid path %s: %v", backupFile, err)
	}
	if _, err := os.Stat(abp); os.IsNotExist(err) {
		mlog.Fatalf("Invalid path %s: %v", csvFile, err)
	}
	if count, err := distill.Restore(abp); err != nil {
		mlog.Fatalf("Error restoring backup from %s: %v", backupFile, err)
	} else {
		mlog.Info("Restored %d URLs from %s ", count, backupFile)
	}

}
