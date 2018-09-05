package cmd

import (
	"github.com/spf13/cobra"
)

var extensions []string

var fmtReplaceCmd = &cobra.Command{
	Use:   "replace [file or folder path]",
	Short: "Format and replace a file or a pool of files in folder",
	Run:   setupCmdFunc(formatAndReplace),
}

func formatAndReplace(msgHandler messageHandler, cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		msgHandler.errorFatalStr("you must provide a filename or a folder as argument")
	}

	if errs := getFileManager().TransformAndReplace(args[0], extensions); len(errs) > 0 {
		for _, e := range errs {
			msgHandler.error(e)
		}

		msgHandler.exit(1)
	}

	msgHandler.success(`"%s" formatted`, args[0])
}

func init() {
	fmtReplaceCmd.Flags().StringSliceVarP(&extensions, "extensions", "e", []string{"feature"}, "Define file extensions to use to find feature files, each separated with a comma")
	fmtCmd.AddCommand(fmtReplaceCmd)
}
