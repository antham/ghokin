package cmd

import (
	"github.com/antham/ghokin/ghokin"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func setupCmdFunc(f func(messageHandler, *cobra.Command, []string)) func(*cobra.Command, []string) {
	return func(cmd *cobra.Command, args []string) {
		msgHandler := newMessageHandler()
		f(msgHandler, cmd, args)
	}
}

func getFileManager() ghokin.FileManager {
	return ghokin.NewFileManager(
		viper.GetInt("indent"),
		viper.GetStringMapString("aliases"),
	)
}

func getStdinManager() ghokin.StdinManager {
	return ghokin.NewStdinManager(
		viper.GetInt("indent"),
		viper.GetStringMapString("aliases"),
	)
}
