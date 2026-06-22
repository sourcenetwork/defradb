// Copyright 2025 Democratized Data Foundation
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
	"context"
	"encoding/json"
	"io"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeDocumentAddManyCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add-many [<json-array>]",
		Short: "Save multiple documents to a collection in a single transaction",
		Long: `Save multiple documents to a collection in a single transaction.

Input must be a JSON array of document objects. Documents that already exist are
updated; new documents are created. Pass '-' to read from stdin.`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return client.ErrCollectionNotFound
			}

			var raw []byte
			switch {
			case len(args) == 1 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return NewErrReadingArgument("stdin", err)
				}
				raw = data
			case len(args) == 1:
				raw = []byte(args[0])
			default:
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return NewErrReadingArgument("stdin", err)
				}
				raw = data
			}

			var rawDocs []json.RawMessage
			if err := json.Unmarshal(raw, &rawDocs); err != nil {
				return NewErrParsingArgument("json array", err)
			}

			cliCtx := cmd.Context()

			docs := make([]*client.Document, 0, len(rawDocs))
			for _, rawDoc := range rawDocs {
				doc, err := client.NewDocFromJSON(cliCtx, rawDoc, col.Version())
				if err != nil {
					return NewErrParsingArgument("document", err)
				}
				docs = append(docs, doc)
			}

			opt := options.WithIdentity(options.SaveDocument(), identity.FromContext(cliCtx))
			return col.SaveManyDocuments(cliCtx, docs, opt)
		},
	}
	setCollectionSelectorFlags(cmd)
	return cmd
}
