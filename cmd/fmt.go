package cmd

import (
	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format a feature file/folder",
	Run:   setupCmdFunc(format),
}

func format(msgHandler messageHandler, cmd *cobra.Command, args []string) {
	if err := cmd.Help(); err != nil {
		msgHandler.errorFatal(err)
	}
}

func init() {
	rootCmd.AddCommand(fmtCmd)
}
