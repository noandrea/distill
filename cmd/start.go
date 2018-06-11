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
	"fmt"
	"net/http"

	"github.com/jbrodriguez/mlog"
	"github.com/spf13/cobra"
	"gitlab.com/welance/oss/distill/internal"
	"gitlab.com/welance/oss/distill/internal/distill"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start distill",
	Long:  ``,
	Run:   start,
}

func init() {
	RootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// startCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

}

func start(cmd *cobra.Command, args []string) {
	mlog.Info("      _ _     _   _ _ _ ")
	mlog.Info("     | (_)   | | (_) | |")
	mlog.Info("   __| |_ ___| |_ _| | |")
	mlog.Info("  / _` | / __| __| | | |")
	mlog.Info(" | (_| | \\__ \\ |_| | | |")
	mlog.Info("  \\__,_|_|___/\\__|_|_|_|  v.%v", version)
	mlog.Info("")
	mlog.Info("Listening to %v:%v", internal.Config.Server.Host, internal.Config.Server.Port)

	distill.NewSession()
	r := distill.RegisterEndpoints()
	http.ListenAndServe(fmt.Sprintf("%s:%d", internal.Config.Server.Host, internal.Config.Server.Port), r)
}
