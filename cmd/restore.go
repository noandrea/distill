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
	"gitlab.com/lowgroundandbigshoes/iljl/internal/iljl"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: restore,
}

func init() {
	RootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVarP(&backupFile, "backup-file", "f", "ilij.backu.bin", "Input for restore")
}

func restore(cmd *cobra.Command, args []string) {
	iljl.NewSession()
	defer iljl.CloseSession()
	abp, err := filepath.Abs(backupFile)
	if err != nil {
		mlog.Fatalf("Invalid path %s: %v", backupFile, err)
	}
	err = os.MkdirAll(filepath.Dir(abp), os.ModeDir)
	if err != nil {
		mlog.Fatalf("Error create backup path to %s: %v", backupFile, err)
	}
	err = iljl.Restore(abp)
	if err != nil {
		mlog.Fatalf("Error create backup at %s: %v", backupFile, err)
	}
}
