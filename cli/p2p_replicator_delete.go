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
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/config"
	"github.com/sourcenetwork/defradb/errors"
)

func MakeP2PReplicatorDeleteCommand(cfg *config.Config, db client.DB) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete <peer>",
		Short: "Delete a replicator. It will stop synchronizing",
		Long:  `Delete a replicator. It will stop synchronizing.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return errors.New("must specify one argument: PeerID")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			addr, err := peer.AddrInfoFromString(args[0])
			if err != nil {
				return err
			}
			return db.DeleteReplicator(cmd.Context(), client.Replicator{Info: *addr})
		},
	}
	return cmd
}
