package cmd

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// restoreCmd represents the restore command
var restoreCmd = &cobra.Command{
	Use:   "restore",
	Short: "Restore a backup",
	Long:  `The backup has to have been created with the backup command`,
	Run:   restore,
}

func init() {
	rootCmd.AddCommand(restoreCmd)
	restoreCmd.Flags().StringVarP(&backupFile, "backup-file", "f", "ilij.backup.bin", "Input for restore")
}

func restore(cmd *cobra.Command, args []string) {
	urlstore.NewSession(settings)
	defer urlstore.CloseSession()
	abp, err := filepath.Abs(backupFile)
	if err != nil {
		log.Fatalf("Invalid path %s: %v", backupFile, err)
	}
	if _, err := os.Stat(abp); os.IsNotExist(err) {
		log.Fatalf("Invalid path %s: %v", csvFile, err)
	}
	if count, err := urlstore.Restore(abp); err != nil {
		log.Fatalf("Error restoring backup from %s: %v", backupFile, err)
	} else {
		log.Infof("Restored %d URLs from %s ", count, backupFile)
	}

}
