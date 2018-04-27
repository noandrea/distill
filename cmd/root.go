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
	"log"
	"os"

	"github.com/jbrodriguez/mlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/lowgroundandbigshoes/iljl/internal"
)

var cfgFile string

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "iljl",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.

	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/iljl/shortener.conf.yaml)")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	//RootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {

	mlog.Start(mlog.LevelInfo, "")
	mlog.DefaultFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	viper.SetConfigName("shortener.conf")       // name of config file (without extension)
	viper.AddConfigPath(os.Getenv("/etc/iljl")) // adding home directory as first search path
	viper.AddConfigPath(os.Getenv("."))
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		mlog.Info("Using config file: %s", viper.ConfigFileUsed())
		viper.Unmarshal(&internal.Config)
	}

}
