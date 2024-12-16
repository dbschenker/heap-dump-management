package functions

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string

	rootCmd = &cobra.Command{

		Use:   "heap-dump-companion",
		Short: "heap-dump-companion - decrypt heap dumps with vault",
		Long: `heap-dump-companion is a super fancy CLI (kidding) to decrypt heap dumps with vault
	   
you can easily decrypt encrypted heap dumps with the transit engine from hashicorp vault.
Just make sure that you are signed in before usage, else this will always result in a permission denied`,
	}
)

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
	viper.WriteConfig()
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.heap-dump-companion.yaml)")
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)

		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".heap-dump-companion")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
