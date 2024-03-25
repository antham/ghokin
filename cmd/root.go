package cmd

import (
	"encoding/json"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "ghokin",
	Short: "Clean and/or apply transformation on gherkin files",
}

// Execute runs root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		newMessageHandler().errorFatal(err)
	}
}

func init() {
	cobra.OnInitialize(initConfig(newMessageHandler()))
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func initConfig(msgHandler messageHandler) func() {
	return func() {
		if cfgFile != "" {
			viper.SetConfigFile(cfgFile)
		} else {
			home, err := homedir.Dir()
			if err != nil {
				msgHandler.errorFatal(err)
			}

			viper.AddConfigPath(home)
			viper.AddConfigPath(".")
			viper.SetConfigName(".ghokin")
		}

		viper.SetEnvPrefix("ghokin")
		for _, err := range []error{
			viper.BindEnv("indent"),
		} {
			if err != nil {
				msgHandler.errorFatal(err)
			}
		}

		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
		viper.AutomaticEnv()
		viper.SetDefault("indent", 2)

		aliases := map[string]string{}
		if err := json.Unmarshal([]byte(viper.GetString("aliases")), &aliases); viper.IsSet("aliases") && err != nil {
			msgHandler.errorFatalStr("check aliases is a well-formed JSON : " + err.Error())
		}
		viper.SetDefault("aliases", aliases)

		if err := viper.ReadInConfig(); err != nil {
			switch err.(type) {
			case viper.ConfigParseError:
				msgHandler.errorFatalStr("check your yaml config file is well-formed : " + err.Error())
			}
		}
	}
}
