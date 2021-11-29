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
	"net/url"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
/* Commented because it is deadcode, for linter.
queryStr string
*/
)

// queryCmd represents the query command
var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Send a GraphQL query",
	Long: `Use this command if you wish to send a formatted GraphQL
query to the database. It's advised to use a proper GraphQL client
to interact with the database, the reccomended approach is with a
local GraphiQL application (https://github.com/graphql/graphiql).

To learn more about the DefraDB GraphQL Query Language, you may use
the additional documenation found at: https://hackmd.io/@source/BksQY6Qfw.
		`,
	Run: func(cmd *cobra.Command, args []string) {
		dbaddr := viper.GetString("database.address")
		if dbaddr == "" {
			log.Error("No database url provided")
		}
		if !strings.HasPrefix(dbaddr, "http") {
			dbaddr = "http://" + dbaddr
		}

		if len(args) != 1 {
			log.Fatal("needs a single query argument")
		}
		query := args[0]
		if query == "" {
			log.Error("missing query")
			return
		}
		endpointStr := fmt.Sprintf("%s/graphql", dbaddr)
		endpoint, err := url.Parse(endpointStr)
		if err != nil {
			log.Fatal(err)
		}

		p := url.Values{}
		p.Add("query", query)
		endpoint.RawQuery = p.Encode()

		res, err := http.Get(endpoint.String())
		if err != nil {
			log.Error("request failed: ", err)
			return
		}
		defer res.Body.Close()
		buf, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Error("request failed: ", err)
			return
		}

		fmt.Printf("Response: %s", string(buf))
	},
}

func init() {
	clientCmd.AddCommand(queryCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// queryCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// queryCmd.Flags().StringVar(&queryStr, "query", "", "Query to run on the database")
}
