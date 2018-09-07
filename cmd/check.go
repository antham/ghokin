package cmd

import (
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check [file or folder path]",
	Short: "Check a file/folder is well formatted",
	Long:  "Check a file/folder is well formatted, otherwise it exit with an error code and the list of file badly formatted",
	Run:   setupCmdFunc(check),
}

func check(msgHandler messageHandler, cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		msgHandler.errorFatalStr("you must provide a filename or a folder as argument")
	}

	if errs := getFileManager().Check(args[0], extensions); len(errs) > 0 {
		for _, e := range errs {
			msgHandler.error(e)
		}

		msgHandler.exit(1)
	}

	msgHandler.success(`"%s" is well formatted`, args[0])
}

func init() {
	checkCmd.Flags().StringSliceVarP(&extensions, "extensions", "e", []string{"feature"}, "Define file extensions to use to find feature files, each separated with a comma")
	rootCmd.AddCommand(checkCmd)
}
