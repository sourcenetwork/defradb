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
	"context"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sourcenetwork/defradb/logging"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	httpapi "github.com/sourcenetwork/defradb/api/http"
)

// requestCmd represents the query request command.
var requestCmd = &cobra.Command{
	Use:   "request",
	Short: "Send a GraphQL query request",
	Long: `Use this command if you wish to send a formatted GraphQL
query request to the database. It's advised to use a proper GraphQL client
to interact with the database, the reccomended approach is with a
local GraphiQL application (https://github.com/graphql/graphiql).

To learn more about the DefraDB GraphQL Query Language, you may use
the additional documentation found at: https://hackmd.io/@source/BksQY6Qfw.
		`,
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		logging.SetConfig(config.Logging.toLogConfig())

		dbaddr := viper.GetString("database.address")
		if dbaddr == "" {
			log.Error(ctx, "No database URL provided")
		}
		if !strings.HasPrefix(dbaddr, "http") {
			dbaddr = "http://" + dbaddr
		}

		if len(args) != 1 {
			log.Fatal(ctx, "needs a single request argument")
		}
		request := args[0]
		if request == "" {
			log.Error(ctx, "missing request")
			return
		}

		endpoint, err := httpapi.JoinPaths(dbaddr, httpapi.GraphQLPath)
		if err != nil {
			log.ErrorE(ctx, "join paths failed", err)
			return
		}

		p := url.Values{}
		p.Add("request", request)
		endpoint.RawQuery = p.Encode()

		res, err := http.Get(endpoint.String())
		if err != nil {
			log.ErrorE(ctx, "request failed", err)
			return
		}

		defer func() {
			err = res.Body.Close()
			if err != nil {
				log.ErrorE(ctx, "response body closing failed: ", err)
			}
		}()

		buf, err := io.ReadAll(res.Body)
		if err != nil {
			log.ErrorE(ctx, "request failed", err)
			return
		}

		log.Info(ctx, "", logging.NewKV("Response", string(buf)))
	},
}

func init() {
	clientCmd.AddCommand(requestCmd)
}
