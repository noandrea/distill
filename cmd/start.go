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

	"gitlab.com/lowgroundandbigshoes/iljl/internal"
	"gitlab.com/lowgroundandbigshoes/iljl/internal/iljl"

	"github.com/jbrodriguez/mlog"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: start,
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
	// TODO: Work your own magic here

	mlog.Info("                                        ")
	mlog.Info("    iiii  lllllll  jjjj lllllll         ")
	mlog.Info("   i::::i l:::::l j::::jl:::::l         ")
	mlog.Info("    iiii  l:::::l  jjjj l:::::l         ")
	mlog.Info("          l:::::l       l:::::l         ")
	mlog.Info("  iiiiiii  l::::ljjjjjjj l::::l         ")
	mlog.Info("  i:::::i  l::::lj:::::j l::::l         ")
	mlog.Info("   i::::i  l::::l j::::j l::::l         ")
	mlog.Info("   i::::i  l::::l j::::j l::::l         ")
	mlog.Info("   i::::i  l::::l j::::j l::::l         ")
	mlog.Info("   i::::i  l::::l j::::j l::::l         ")
	mlog.Info("   i::::i  l::::l j::::j l::::l         ")
	mlog.Info("   i::::i  l::::l j::::j l::::l         ")
	mlog.Info("  i::::::il::::::lj::::jl::::::l        ")
	mlog.Info("  i::::::il::::::lj::::jl::::::l ...... ")
	mlog.Info("  i::::::il::::::lj::::jl::::::l .::::. ")
	mlog.Info("  iiiiiiiillllllllj::::jllllllll ...... ")
	mlog.Info("                  j::::j                ")
	mlog.Info("        jjjj      j::::j                ")
	mlog.Info("       j::::jj   j:::::j                ")
	mlog.Info("       j::::::jjj::::::j                ")
	mlog.Info("        jj::::::::::::j                 ")
	mlog.Info("          jjj::::::jjj                  ")
	mlog.Info("             jjjjjj                     ")
	mlog.Info("                                        ")
	mlog.Info("      !!    starting    !!              ")

	iljl.NewSession()
	r := iljl.RegisterEndpoints()
	http.ListenAndServe(fmt.Sprintf("%s:%d", internal.Config.Server.Host, internal.Config.Server.Port), r)
}
