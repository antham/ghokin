package cmd

import (
	"github.com/spf13/cobra"
)

var extensions []string

var fmtReplaceCmd = &cobra.Command{
	Use:   "replace [files or folders path]",
	Short: "Format and replace files",
	Run:   setupCmdFunc(formatAndReplace),
}

func formatAndReplace(msgHandler messageHandler, cmd *cobra.Command, args []string) {
	if len(args) == 0 {
		msgHandler.errorFatalStr("you must provide filenames or folders as argument")
	}

	if errs := getFileManager().TransformAndReplace(args, extensions); len(errs) > 0 {
		for _, e := range errs {
			msgHandler.error(e)
		}

		msgHandler.exit(1)
	}

	for _, a := range args {
		msgHandler.success(`"%s" formatted`, a)
	}
}

func init() {
	fmtReplaceCmd.Flags().StringSliceVarP(&extensions, "extensions", "e", []string{"feature"}, "Define file extensions to use to find feature files, each separated with a comma")
	fmtCmd.AddCommand(fmtReplaceCmd)
}
