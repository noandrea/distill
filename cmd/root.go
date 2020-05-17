package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/noandrea/distill/urlstore"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile, logFile string
var profile, debug bool

// Config system configuration
var settings urlstore.ConfigSchema

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
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
	rootCmd.Version = v
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is /etc/distill/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&logFile, "log", "", "set a logging file, default stdout")
	rootCmd.PersistentFlags().BoolVar(&debug, "debug", false, "enable debug")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if debug {
		// Only log the warning severity or above.
		log.SetLevel(log.DebugLevel)
	}

	// set configuration paramteres
	viper.SetConfigName("config")       // name of config file (without extension)
	viper.AddConfigPath("/etc/distill") // adding home directory as first search path
	viper.SetEnvPrefix("DISTILL")
	viper.AutomaticEnv() // read in environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	urlstore.Defaults() // load defaults for configuration
	// if there is the config file read it
	if len(cfgFile) > 0 { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Infof("Using config file: %v", viper.ConfigFileUsed())
		viper.Unmarshal(&settings)
		settings.Validate()
	} else {
		switch err.(type) {
		case viper.ConfigFileNotFoundError:
			viper.Unmarshal(&settings)
		}
	}
	// make the version available via settings
	settings.RuntimeVersion = rootCmd.Version
	log.Debugf("config %#v", settings)
}
