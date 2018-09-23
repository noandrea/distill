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

	"gitlab.com/welance/oss/distill/urlstore"

	"github.com/jbrodriguez/mlog"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gitlab.com/welance/oss/distill/pkg/common"
)

var cfgFile, logFile, version string
var profile, debug, generateConfigOnly bool

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "distill",
	Short: "A practical url shortener",
	Long:  ``,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute(v string) {
	version = v
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
	RootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/distill/settings.yaml)")
	RootCmd.PersistentFlags().StringVar(&logFile, "log", "", "set a logging file, default stdout")
	RootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug")
	RootCmd.PersistentFlags().BoolVar(&generateConfigOnly, "generate-config", false, "print the default configuration file and exit")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// enable debug logging if required
	loglevel := mlog.LevelInfo
	if debug {
		loglevel = mlog.LevelTrace
	}
	// start logging
	mlog.Start(loglevel, logFile)
	mlog.DefaultFlags = log.Ldate | log.Ltime | log.Lmicroseconds | log.Lshortfile

	// if only generate config
	if generateConfigOnly {

	}

	// set configuration paramteres
	viper.SetConfigName("settings")     // name of config file (without extension)
	viper.AddConfigPath("/etc/distill") // adding home directory as first search path
	viper.AddConfigPath("./configs")
	viper.AddConfigPath(".")
	viper.AutomaticEnv() // read in environment variables that match
	// if there is the config file read it
	if len(cfgFile) > 0 { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		mlog.Info("Using config file: %v", viper.ConfigFileUsed())
		viper.Unmarshal(&urlstore.Config)
		urlstore.Config.Defaults()
		urlstore.Config.Validate()
	} else {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			if do := common.AskYes("A configuration file was not found, would you like to generate one?", true); do {
				urlstore.GenerateDefaultConfig("settings.yaml", version)
				fmt.Println("Configuration settings.yaml created")
				return
			}
		}
		fmt.Println("Configuration file not found!!")
		os.Exit(1)
	}
}
