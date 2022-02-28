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
	"context"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	yaml "gopkg.in/yaml.v2"
)

var (
	// root flag vars
	cfgFile           string
	dbURL             string
	logLvl            string
	stackTraceEnabled bool

	log = logging.MustNewLogger("defra.cli")

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
	rootCmd.PersistentFlags().StringVar(&logLvl, "log", "", "Log level to use, options are info, debug, error")
	rootCmd.PersistentFlags().BoolVar(&stackTraceEnabled, "stacktrace", false, "Include stacktrace in Error and Fatal logs")

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.defradb/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&dbURL, "url", "http://localhost:9181", "url of the target database")
	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	cobra.OnInitialize(initConfig)
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
		ctx := context.Background()
		applyCliParamsToConfig(ctx)
		logging.SetConfig(config.Logging.toLogConfig())
		log.Debug(ctx, "Loaded config file", logging.NewKV("ConfigFile", viper.ConfigFileUsed()))
	} else {
		dir := home + "/.defradb"
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			if err := os.Mkdir(dir, os.ModePerm); err != nil {
				cobra.CheckErr(err)
			}
		}

		defaultConfig.Database.Badger.Path = strings.Replace(defaultConfig.Database.Badger.Path, "$HOME", home, -1)
		bs, err := yaml.Marshal(defaultConfig)
		cobra.CheckErr(err)

		err = viper.ReadConfig(bytes.NewBuffer(bs))
		cobra.CheckErr(err)

		err = viper.WriteConfigAs(home + "/.defradb/" + "config.yaml")
		cobra.CheckErr(err)

		// we should do this *after* writing the default config to file
		ctx := context.Background()
		applyCliParamsToConfig(ctx)
	}

	err := viper.BindPFlag("database.address", rootCmd.Flags().Lookup("url"))
	cobra.CheckErr(err)

	err = viper.BindPFlag("database.store", startCmd.Flags().Lookup("store"))
	cobra.CheckErr(err)

	err = viper.BindPFlag("database.badger.path", startCmd.Flags().Lookup("data"))
	cobra.CheckErr(err)

	err = viper.BindPFlag("net.p2paddress", startCmd.Flags().Lookup("p2paddr"))
	cobra.CheckErr(err)

	err = viper.BindPFlag("net.tcpaddress", startCmd.Flags().Lookup("tcpaddr"))
	cobra.CheckErr(err)

	err = viper.BindPFlag("net.p2pdisabled", startCmd.Flags().Lookup("no-p2p"))
	cobra.CheckErr(err)

	err = viper.Unmarshal(&config)
	cobra.CheckErr(err)
}

func applyCliParamsToConfig(ctx context.Context) {
	lvls := strings.Split(logLvl, ",")
	if lvls[0] == "" {
		config.Logging.Level = &lvls[0]

		if len(lvls) > 1 {
			namedOptions := []NamedLoggerOptions{}

			for _, l := range lvls[1:] {
				namedLevel := strings.Split(l, "=")
				if len(namedLevel) != 2 {
					log.Fatal(ctx, "Invalid format for log level", logging.NewKV("Value", l))
				}
				namedOptions = append(namedOptions, NamedLoggerOptions{
					Name:  namedLevel[0],
					Level: &namedLevel[1],
				})
			}

			config.Logging.NamedOptions = &namedOptions
		}
	}

	if stackTraceEnabled {
		// should be nil by default, not false
		config.Logging.EnableStackTrace = &stackTraceEnabled
	}
}
