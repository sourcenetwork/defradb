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

	"github.com/spf13/cobra"
)

func MakeSDLCommand(ctx context.Context) *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "sdl",
		Short: "Utilities to interact with the DefraDB SDL",
		Long:  `Utilities to interact with the DefraDB Schema Definition Language.`,
	}

	return cmd
}
