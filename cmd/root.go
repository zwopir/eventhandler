package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"flag"
	"github.com/prometheus/common/log"
	"github.com/spf13/pflag"
)

var (
	cfgFile  string
	logLevel string
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "eventhandler",
	Short: "A nats queue based remote execution/trigger framework",
	Long:  `A nats queue based remote execution/trigger framework for monitoring event handler`,
}

// Execute adds all child commands to the root command sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func init() {
	// add go flags to pflag
	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	cobra.OnInitialize(InitConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/config.yaml)")
	RootCmd.PersistentFlags().StringVar(&logLevel, "log.level", "info", "log level")
}

// InitConfig reads in config file and ENV variables if set.
func InitConfig() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AddConfigPath("/etc/eventhandler")
	viper.AutomaticEnv() // read in environment variables that match

	log.Base().SetLevel(logLevel)

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
	}

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("can't read config from config file %s: %s", viper.ConfigFileUsed(), err)
	}

	log.Debugf("read config from %s", viper.ConfigFileUsed())
	log.Debugf("effective config: %v", viper.AllSettings())
}
