// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package cli

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/spf13/cobra"

	httpapi "github.com/sourcenetwork/defradb/api/http"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
)

type schemaListResponse struct {
	Data struct {
		Collections []struct {
			Name   string `json:"name"`
			ID     string `json:"id"`
			Fields []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Kind string `json:"kind"`
			} `json:"fields"`
		} `json:"collections"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func MakeSchemaListCommand(cfg *config.Config) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List schema types with their respective fields",
		RunE: func(cmd *cobra.Command, args []string) (err error) {
			endpoint, err := httpapi.JoinPaths(cfg.API.AddressToURL(), httpapi.SchemaPath)
			if err != nil {
				return NewErrFailedToJoinEndpoint(err)
			}

			res, err := http.Get(endpoint.String())
			if err != nil {
				return NewErrFailedToSendRequest(err)
			}
			defer res.Body.Close() //nolint:errcheck

			data, err := io.ReadAll(res.Body)
			if err != nil {
				return NewErrFailedToReadResponseBody(err)
			}

			var r schemaListResponse
			if err := json.Unmarshal(data, &r); err != nil {
				return NewErrFailedToUnmarshalResponse(err)
			}
			if len(r.Errors) > 0 {
				return errors.New("failed to list schemas", errors.NewKV("errors", r.Errors))
			}

			for _, c := range r.Data.Collections {
				cmd.Printf("type %s {\n", c.Name)
				for _, f := range c.Fields {
					cmd.Printf("\t%s: %s\n", f.Name, f.Kind)
				}
				cmd.Printf("}\n")
			}

			return nil
		},
	}
	return cmd
}
