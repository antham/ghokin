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
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := homedir.Dir()

		if err != nil {
			newMessageHandler().errorFatal(err)
		}

		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName(".ghokin-fmt")
	}

	viper.SetEnvPrefix("ghokin")
	for _, err := range []error{
		viper.BindEnv("indent.backgroundAndScenario"),
		viper.BindEnv("indent.step"),
		viper.BindEnv("indent.tableAndDocString"),
	} {
		if err != nil {
			newMessageHandler().errorFatal(err)
		}
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()

	viper.SetDefault("indent.backgroundAndScenario", 2)
	viper.SetDefault("indent.step", 4)
	viper.SetDefault("indent.tableAndDocString", 6)

	commands := map[string]string{}
	if err := json.Unmarshal([]byte(viper.GetString("commands")), &commands); err == nil {
		viper.SetDefault("commands", commands)
	}

	if err := viper.ReadInConfig(); err != nil {
		return
	}
}
