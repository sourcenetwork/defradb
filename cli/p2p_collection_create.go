// Copyright 2022 Democratized Data Foundation
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
	"strings"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeP2PCollectionCreateCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "create [collectionNames]",
		Short: "Create P2P collections",
		Long: `Create P2P collections to the synchronized pubsub topics.
The collections are synchronized between nodes of a pubsub network.`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)

			var collectionNames []string
			for id := range strings.SplitSeq(args[0], ",") {
				id = strings.TrimSpace(id)
				if id == "" {
					continue
				}
				collectionNames = append(collectionNames, id)
			}

			opt := options.WithIdentity(options.CreateP2PCollections(), identity.FromContext(cmd.Context()))
			return cliClient.CreateP2PCollections(cmd.Context(), collectionNames, opt)
		},
	}

	EmbedCLIExample(ctx, cmd, "create single collection",
		`defradb client p2p collection create User`)

	EmbedCLIExample(ctx, cmd, "create multiple collections",
		`defradb client p2p collection create User,Address`)

	return cmd
}
