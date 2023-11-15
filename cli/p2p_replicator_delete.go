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

	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
)

func MakeP2PReplicatorDeleteCommand() *cobra.Command {
	var collections []string
	var cmd = &cobra.Command{
		Use:   "delete [-c, --collection] <peer>",
		Short: "Delete replicator(s) and stop synchronization",
		Long: `Delete replicator(s) and stop synchronization.
A replicator synchronizes one or all collection(s) from this node to another.
		
Example:		
  defradb client p2p replicator delete -c Users '{"ID": "12D3", "Addrs": ["/ip4/0.0.0.0/tcp/9171"]}'
		`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			p2p := mustGetP2PContext(cmd)

			var info peer.AddrInfo
			if err := json.Unmarshal([]byte(args[0]), &info); err != nil {
				return err
			}
			rep := client.Replicator{
				Info:    info,
				Schemas: collections,
			}
			return p2p.DeleteReplicator(cmd.Context(), rep)
		},
	}
	cmd.Flags().StringSliceVarP(&collections, "collection", "c",
		[]string{}, "Collection(s) to stop replicating")
	return cmd
}
