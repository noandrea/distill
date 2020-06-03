package cmd

import (
	"fmt"

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
	fmt.Printf("not yet implemented")
}
