// Copyright 2022 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
package cmd

import (
	"bytes"
	"os"
	"strings"

	logging "github.com/ipfs/go-log/v2"
	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var (
	// root flag vars
	cfgFile string
	dbURL   string
	logLvl  string

	log = logging.Logger("defra.cli")

	config Config
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "defradb",
	Short: "DefraDB Edge Database",
	Long: `DefraDB is the edge database to power the user-centric future.
This CLI is the main reference implementation of DefraDB. Use it to start
a new database process, query a local or remote instance, and much more.
For example:

# Start a new database instance
> defradb start `,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// expose root as public
var RootCmd = rootCmd

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	cobra.CheckErr(rootCmd.Execute())
}

func init() {
	rootCmd.PersistentFlags().StringVar(&logLvl, "log", "info", "Log level to use, options are info, debug, error")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.defradb.yaml)")
	rootCmd.PersistentFlags().StringVar(&dbURL, "url", "http://localhost:9181", "url of the target database")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// cobra.OnInitialize()
	cobra.OnInitialize(initConfig, initLogger)
}

func initLogger() {
	lvl, err := logging.LevelFromString(logLvl)
	if err != nil {
		panic(err)
	}
	logging.SetAllLoggers(lvl)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	var home string
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		var err error
		home, err = homedir.Dir()
		cobra.CheckErr(err)

		// Search config in home directory with name ".defradb" (without extension).
		viper.AddConfigPath(home + "/.defradb")
		viper.SetConfigName("config")
		viper.SetConfigType("yaml")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		// fmt.Fprintln(os.Stderr, "Using config file:", viper.ConfigFileUsed())
		log.Debug("Loading config file:", viper.ConfigFileUsed())
	} else {
		if err := os.Mkdir(home+"/.defradb", os.ModePerm); err != nil {
			cobra.CheckErr(err)
		}
		// if err != nil {
		// 	cobra.CheckErr(err)
		// }
		// fmt.Fprintln(os.Stdout, "Generating default config file")
		defaultConfig.Database.Badger.Path = strings.Replace(defaultConfig.Database.Badger.Path, "$HOME", home, -1)
		bs, err := yaml.Marshal(defaultConfig)
		cobra.CheckErr(err)

		err = viper.ReadConfig(bytes.NewBuffer(bs))
		cobra.CheckErr(err)

		err = viper.WriteConfigAs(home + "/.defradb/" + "config.yaml")
		cobra.CheckErr(err)
	}

	err := viper.BindPFlag("database.address", rootCmd.Flags().Lookup("url"))
	cobra.CheckErr(err)

	err = viper.BindPFlag("database.store", startCmd.Flags().Lookup("store"))
	cobra.CheckErr(err)

	err = viper.Unmarshal(&config)
	cobra.CheckErr(err)
}
