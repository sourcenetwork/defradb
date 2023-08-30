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

	"github.com/spf13/cobra"
)

// func newHttpClient(cfg *config.Config) (client.Store, error) {
// 	db, err := http.NewClient(cfg.API.Address)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if txId != 0 {
// 		return db.WithTxnID(txId), nil
// 	}
// 	return db, nil
// }

func writeJSON(cmd *cobra.Command, out any) error {
	enc := json.NewEncoder(cmd.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}
