// Package cmd for the cli commands
package cmd

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/noandrea/distill/urlstore"
	"github.com/noandrea/distill/web"

	log "github.com/sirupsen/logrus"
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
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().StringVarP(&restoreFile, "restore", "r", "", "Restore data from file before starting")
}

func start(cmd *cobra.Command, args []string) {
	log.Info("      _ _     _   _ _ _ ")
	log.Info("     | (_)   | | (_) | |")
	log.Info("   __| |_ ___| |_ _| | |")
	log.Info("  / _` | / __| __| | | |")
	log.Info(" | (_| | \\__ \\ |_| | | |")
	log.Infof("  \\__,_|_|___/\\__|_|_|_|  v.%v", settings.RuntimeVersion)
	log.Info("")
	log.Infof("Listening to %v:%v", settings.Server.Host, settings.Server.Port)

	urlstore.NewSession(settings)
	if len(strings.TrimSpace(restoreFile)) > 0 {
		count, err := urlstore.Restore(restoreFile)
		if err != nil {
			log.Fatalf("Error restoring URLs from %s: %v ", restoreFile, err)
		}
		log.Infof("Restored %d URLs from %s ", count, restoreFile)
	}
	r := web.RegisterEndpoints(settings)
	http.ListenAndServe(fmt.Sprintf("%s:%d", settings.Server.Host, settings.Server.Port), r)
}
