// Package cmd for the cli commands
package cmd

// Copyright © 2018 Andrea Giacobino <no.andrea@gmail.com>
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

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/noandrea/distill/urlstore"
	"github.com/noandrea/distill/web"

	"github.com/jbrodriguez/mlog"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start distill",
	Long:  ``,
	Run:   start,
}

var restoreFile string

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().StringVarP(&restoreFile, "restore", "r", "", "Restore data from file before starting")
}

func start(cmd *cobra.Command, args []string) {
	mlog.Info("      _ _     _   _ _ _ ")
	mlog.Info("     | (_)   | | (_) | |")
	mlog.Info("   __| |_ ___| |_ _| | |")
	mlog.Info("  / _` | / __| __| | | |")
	mlog.Info(" | (_| | \\__ \\ |_| | | |")
	mlog.Info("  \\__,_|_|___/\\__|_|_|_|  v.%v", version)
	mlog.Info("")
	mlog.Info("Listening to %v:%v", urlstore.Config.Server.Host, urlstore.Config.Server.Port)

	urlstore.NewSession()
	if len(strings.TrimSpace(restoreFile)) > 0 {
		count, err := urlstore.Restore(restoreFile)
		if err != nil {
			mlog.Fatalf("Error restoring URLs from %s: %v ", restoreFile, err)
		}
		mlog.Info("Restored %d URLs from %s ", count, restoreFile)
	}
	r := web.RegisterEndpoints()
	http.ListenAndServe(fmt.Sprintf("%s:%d", urlstore.Config.Server.Host, urlstore.Config.Server.Port), r)
}
