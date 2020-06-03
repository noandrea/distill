package cmd

import (
	"fmt"

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
	fmt.Println("not yet implemented")

}
