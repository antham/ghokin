package cmd

import (
	"github.com/spf13/cobra"
)

var fmtStdoutCmd = &cobra.Command{
	Use:   "stdout [file path]",
	Short: "Format stdin or a file and dump the result on stdout",
	Run:   setupCmdFunc(formatOnStdout),
}

func formatOnStdout(msgHandler messageHandler, cmd *cobra.Command, args []string) {
	var output []byte
	var err error
	if len(args) == 0 {
		output, err = getStdinManager().Transform(cmd.InOrStdin())
	} else {
		output, err = getFileManager().Transform(args[0])
	}

	if err != nil {
		msgHandler.errorFatal(err)
	}
	msgHandler.print("%s", output)
}

func init() {
	fmtCmd.AddCommand(fmtStdoutCmd)
}
