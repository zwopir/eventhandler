package cmd

import (
	"fmt"
	"os"

	"eventhandler/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"flag"
	"github.com/prometheus/common/log"
	"github.com/spf13/pflag"
)

var (
	cfgFile string
	cfg     config.Config
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
	cobra.OnInitialize(initConfig)

	RootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/config.yaml)")
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath("$HOME")  // adding home directory as first search path
	viper.AddConfigPath("/etc/eventhandler")
	viper.AutomaticEnv() // read in environment variables that match

	if cfgFile != "" { // enable ability to specify config file via flag
		viper.SetConfigFile(cfgFile)
		log.Infof("using config file %s", cfgFile)
	}

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("can't read config from config file %s: %s", viper.ConfigFileUsed(), err)
	}
	// unmarshal viper to config/Config struct
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
}
