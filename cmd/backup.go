package cmd

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

var backupFile string

// backupCmd represents the backup command
var backupCmd = &cobra.Command{
	Use:   "backup",
	Short: "backup the database urls",
	Long: `Create a backup of the urls database.
  The backup format can be binary or csv, the format will be 
  selected by the extension of the backup file (.bin for binary and .csv for csv).
  The backup command will try to create the output file and all the intermediate folders.
  
  The backup cannot be executed in a live service`,
	Run: backup,
}

func init() {
	rootCmd.AddCommand(backupCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// backupCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	backupCmd.Flags().StringVarP(&backupFile, "backup-file", "f", "ilij.backu.bin", "Output file for backup")

}

func backup(cmd *cobra.Command, args []string) {
	urlstore.NewSession(settings)
	defer urlstore.CloseSession()
	abp, err := filepath.Abs(backupFile)
	if err != nil {
		log.Fatalf("Invalid path %s: %v", backupFile, err)
	}
	err = os.MkdirAll(filepath.Dir(abp), os.ModeDir)
	if err != nil {
		log.Fatalf("Error create backup path to %s: %v", backupFile, err)
	}
	if err = urlstore.Backup(abp); err != nil {
		log.Fatalf("Error create backup at %s: %v", backupFile, err)
	}
}
