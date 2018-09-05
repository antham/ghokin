package cmd

import (
	"github.com/spf13/cobra"
)

var fmtStdoutCmd = &cobra.Command{
	Use:   "stdout [file path]",
	Short: "Format a file and dump the result on stdout",
	Run:   setupCmdFunc(formatOnStdout),
}

func formatOnStdout(msgHandler messageHandler, cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		msgHandler.errorFatalStr("you must provide a filename as argument")
	}

	output, err := getFileManager().Transform(args[0])

	if err != nil {
		msgHandler.errorFatal(err)
	}

	msgHandler.print("%s", output.String())
}

func init() {
	fmtCmd.AddCommand(fmtStdoutCmd)
}
