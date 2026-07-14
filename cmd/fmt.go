package cmd

import (
	"github.com/spf13/cobra"
)

var fmtCmd = &cobra.Command{
	Use:   "fmt",
	Short: "Format stdin or feature files/folders",
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
