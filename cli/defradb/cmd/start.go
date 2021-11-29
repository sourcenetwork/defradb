// Copyright 2020 Source Inc.
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
	"context"
	"os"
	"os/signal"

	"github.com/sourcenetwork/defradb/db"

	ds "github.com/ipfs/go-datastore"
	badgerds "github.com/ipfs/go-ds-badger"
	"github.com/spf13/cobra"
)

// startCmd represents the start command
var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a DefraDB server ",
	Long:  `Start a new instance of DefraDB server:`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Info("Starting DefraDB process...")
		ctx := context.Background()

		// setup signal handlers
		signalCh := make(chan os.Signal, 1)
		signal.Notify(signalCh, os.Interrupt)

		var rootstore ds.Batching
		var options interface{}
		var err error
		if config.Database.Store == "badger" {
			log.Info("opening badger store: ", config.Database.Badger.Path)
			rootstore, err = badgerds.NewDatastore(config.Database.Badger.Path, config.Database.Badger.Options)
			options = config.Database.Badger
			if err != nil {
				log.Error("Failed to initiate database:", err)
				os.Exit(1)
			}
		} else if config.Database.Store == "memory" {
			log.Info("building new memory store")
			rootstore = ds.NewMapDatastore()
			options = config.Database.Memory
		}

		db, err := db.NewDB(rootstore, options)
		if err != nil {
			log.Error("Failed to initiate database:", err)
			os.Exit(1)
		}
		if err := db.Start(ctx); err != nil {
			log.Error("Failed to start the database: ", err)
			db.Close()
			os.Exit(1)
		}

		// run the server listener in a seperate goroutine
		go func() {
			db.Listen(config.Database.Address)
		}()

		/*
			Commenting out this block of code because the linter crys with this:
			S1000: should use a simple channel send/receive instead of `select` with a single case
			Note: Removed the the `signalCh` out of `select { case }` in the following lines.

					// capture the interrupt signal, and gracefully exit
					// @todo: Handle hard interuppt
					select {
					case <-signalCh:
						log.Info("Recieved interrupt; closing db")
						db.Close()
						os.Exit(0)
					}
		*/

		<-signalCh
		log.Info("Recieved interrupt; closing db")
		db.Close()
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(startCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// startCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	startCmd.Flags().String("store", "badger", "Specify the data store to use (supported: badger, memory)")

}
