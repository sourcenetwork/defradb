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
	"github.com/spf13/cobra"
)

func MakeNodeACPRelationshipAddCommand() *cobra.Command {
	var (
		relationArg    string
		targetActorArg string
	)

	var cmd = &cobra.Command{
		Use:   "add [-r --relation] [-a --actor] [-i --identity]",
		Short: "Add new relationship",
		Long: `Add new relationship

To share node access (or grant a more restricted access) with another actor, we must add the type of
relationship for that actor. In order to make the relationship we require all of the following:
1) Relation Name: The type of relation (name must be defined within the nac policy).
2) Target Identity: The identity of the actor the relationship is being made with.
3) Requesting Identity: The identity of the actor that is making the request.

Notes:
  - The requesting identity MUST either be the owner OR the manager (manages the relation) of the resource.
  - Currently the only relation supported is the 'admin' relation.
  - The Target Identity format is a public key format.
  - The Requesting Identity is a secp256k1 private key in hex format.

Example: Make another actor an admin user:
  defradb client acp node relationship add \
	--relation admin \
	--actor did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn \
	--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cliClient := mustGetContextCLIClient(cmd)
			addNACActorRelationshipResult, err := cliClient.AddNACActorRelationship(
				cmd.Context(),
				relationArg,
				targetActorArg,
			)

			if err != nil {
				return err
			}

			return writeJSON(cmd, addNACActorRelationshipResult)
		},
	}

	cmd.Flags().StringVarP(
		&relationArg,
		"relation",
		"r",
		"",
		"Relation that needs to be set for the relationship",
	)
	_ = cmd.MarkFlagRequired("relation")

	cmd.Flags().StringVarP(
		&targetActorArg,
		"actor",
		"a",
		"",
		"Actor to add relationship with",
	)
	_ = cmd.MarkFlagRequired("actor")

	return cmd
}
