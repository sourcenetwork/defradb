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
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a block by its CID from the blockstore",
	Run: func(cmd *cobra.Command, args []string) {
		dbaddr := viper.GetString("database.address")
		if dbaddr == "" {
			log.Error("No database url provided")
		}
		if !strings.HasPrefix(dbaddr, "http") {
			dbaddr = "http://" + dbaddr
		}

		if len(args) != 1 {
			log.Fatal("Needs a single CID")
		}
		cid := args[0]

		res, err := http.Get(fmt.Sprintf("%s/blocks/get/%s", dbaddr, cid))
		if err != nil {
			log.Error("request failed: ", err)
			return
		}

		defer func() {
			err = res.Body.Close()
			if err != nil {
				// Should this be `log.Fatal` ??
				log.Error("response body closing failed: ", err)
			}
		}()

		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Error("request failed: ", err)
			return
		}
		fmt.Println("Block:")
		fmt.Println(string(buf))
	},
}

func init() {
	blocksCmd.AddCommand(getCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
