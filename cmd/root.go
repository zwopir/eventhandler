package cmd

import (
	"fmt"
	"os"

	"eventhandler/config"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/prometheus/common/log"
	"flag"
	"github.com/spf13/pflag"
)

var (
	cfgFile string
	cfg config.Config
)

// RootCmd represents the base command when called without any subcommands
var RootCmd = &cobra.Command{
	Use:   "eventhandler",
	Short: "A brief description of your application",
	Long: `A longer description that spans multiple lines and likely contains
examples and usage of using your application. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
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

	// Here you will define your flags and configuration settings.
	// Cobra supports Persistent Flags, which, if defined here,
	// will be global for your application.
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
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Fatal(err.Error())
	}
}
