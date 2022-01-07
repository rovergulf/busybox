/*
Copyright © 2022 Rovergulf Engineers <support@rovergulf.net>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"github.com/rovergulf/busybox/handler"
	"github.com/spf13/cobra"
	"os"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "busybox",
	Short: "REST Server debug tool",
	Long: `Rovergulf Engineers Busybox - is a simple REST server can be used
as a incoming HTTP request debug tool`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	RunE: func(cmd *cobra.Command, args []string) error {
		h := new(handler.Handler)
		return h.Run()
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.busybox.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().String("jaeger_trace", os.Getenv("JAEGER_TRACE"), "Jaeger tracing collector address")
	rootCmd.Flags().String("env", "dev", "App environment")
	rootCmd.Flags().Bool("log_json", false, "Enable JSON logging")
	rootCmd.Flags().Bool("log_stacktrace", true, "Enable logger stacktrace")
	rootCmd.Flags().String("listen-addr", ":8080", "TCP address listen to")

	viper.BindPFlag("log_json", rootCmd.Flags().Lookup("log_json"))
	viper.BindPFlag("log_stacktrace", rootCmd.Flags().Lookup("log_stacktrace"))
	viper.BindPFlag("jaeger_addr", rootCmd.Flags().Lookup("jaeger_addr"))
	viper.BindPFlag("env", rootCmd.Flags().Lookup("env"))
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".busybox" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".busybox")
	}

	viper.AutomaticEnv() // read in environment variables that match

	viper.SetDefault("listen_addr", ":8081")

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}