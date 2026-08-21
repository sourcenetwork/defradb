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
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeCollectionTruncateCommand(ctx context.Context) *cobra.Command {
	var filter string
	var pruneHistory bool
	var cmd = &cobra.Command{
		Use:   "truncate",
		Short: "Truncate the given collection",
		Long: `Truncate the given collection, removing document data from the local node.
Without a filter all documents are removed. With a filter only matching documents are removed,
and --prune-history also removes their unshared history. Changes do not propagate to other nodes.`,
		Args: cobra.ExactArgs(0),
		RunE: func(cmd *cobra.Command, args []string) error {
			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return client.ErrCollectionNotFound
			}

			opt := options.WithIdentity(options.TruncateCollection(), identity.FromContext(cmd.Context()))
			if pruneHistory && filter == "" {
				return fmt.Errorf("--prune-history requires --filter")
			}
			if filter != "" {
				filterValue, err := parseTruncateFilter(filter)
				if err != nil {
					return NewErrParsingArgument("filter", err)
				}
				opt.SetFilter(filterValue).SetPruneHistory(pruneHistory)
			}
			return col.Truncate(cmd.Context(), opt)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Document filter")
	cmd.Flags().BoolVar(&pruneHistory, "prune-history", false, "Remove unshared history for matching documents")
	setCollectionSelectorFlags(cmd)
	return cmd
}

func parseTruncateFilter(value string) (any, error) {
	var filter any
	if err := json.Unmarshal([]byte(value), &filter); err != nil {
		return nil, err
	}
	if filter == nil {
		return nil, errors.New("filter cannot be null")
	}
	return filter, nil
}
