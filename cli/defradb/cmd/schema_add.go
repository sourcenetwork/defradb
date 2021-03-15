// Copyright 2020 Source Inc.
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.
/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	schemaFile string
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new schema type to a DefraDB instance",
	Long: `Example Usage:
> defradb client schema add -f user.sdl`,
	Run: func(cmd *cobra.Command, args []string) {
		var schema []byte
		if len(args) > 0 {
			schema = []byte(strings.Join(args, "\n"))
		} else if schemaFile != "" {
			buf, err := ioutil.ReadFile(schemaFile)
			cobra.CheckErr(err)
			schema = buf
		} else {
			log.Fatal("Missing schema")
		}

		dbaddr := viper.GetString("database.address")
		if dbaddr == "" {
			log.Error("No database url provided")
		}
		if !strings.HasPrefix(dbaddr, "http") {
			dbaddr = "http://" + dbaddr
		}
		endpointStr := fmt.Sprintf("%s/schema/load", dbaddr)
		endpoint, err := url.Parse(endpointStr)
		cobra.CheckErr(err)

		r, err := http.Post(endpoint.String(), "text", bytes.NewBuffer(schema))
		cobra.CheckErr(err)
		defer r.Body.Close()
		result, err := ioutil.ReadAll(r.Body)
		cobra.CheckErr(err)
		fmt.Println(string(result))
	},
}

func init() {
	schemaCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	addCmd.Flags().StringVarP(&schemaFile, "file", "f", "", "File to load a schema from")
}
