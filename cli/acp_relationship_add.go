// Copyright 2024 Democratized Data Foundation
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

func MakeACPRelationshipAddCommand() *cobra.Command {
	const (
		collectionFlagLong  string = "collection"
		collectionFlagShort string = "c"

		relationFlagLong  string = "relation"
		relationFlagShort string = "r"

		targetActorFlagLong  string = "actor"
		targetActorFlagShort string = "a"

		docIDFlag string = "docID"
	)

	var (
		collectionArg  string
		relationArg    string
		targetActorArg string
		docIDArg       string
	)

	var cmd = &cobra.Command{
		Use:   "add [-i --identity] [policy]",
		Short: "Add new relationship",
		Long: `Add new relationship

Notes:
  - ACP must be available (i.e. ACP can not be disabled).
  - The target document must be registered with ACP already (policy & resource specified).
  - The requesting identity MUST either be the owner OR the manager (manages the relation) of the resource.
  - If the specified relation was not granted the miminum DPI permissions (read or write) within the policy,
  and a relationship is formed, the subject/actor will still not be able to access (read or write) the resource.
  - Learn more about [ACP & DPI Rules](/acp/README.md)

Consider the following policy:
'
description: A Policy

actor:
  name: actor

resources:
  users:
    permissions:
      read:
        expr: owner + reader + writer
      write:
        expr: owner + writer
      nothing:
        expr: dummy

    relations:
      owner:
        types:
          - actor
      reader:
        types:
          - actor
      writer:
        types:
          - actor
      admin:
        manages:
          - reader
        types:
          - actor
      dummy:
        types:
          - actor
'

defradb client ... --identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac


Example: Let another actor read my private document:
  defradb client acp relationship add --collection User --docID bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab \
	--relation reader --actor did:key:z6MkkHsQbp3tXECqmUJoCJwyuxSKn1BDF1RHzwDGg9tHbXKw \
	--identity 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

Example: Create a dummy relation that doesn't do anything (from database prespective):
  defradb client acp relationship add -c User --docID bae-91171025-ed21-50e3-b0dc-e31bccdfa1ab -r dummy \
	-a did:key:z6MkkHsQbp3tXECqmUJoCJwyuxSKn1BDF1RHzwDGg9tHbXKw \
	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f

`,
		RunE: func(cmd *cobra.Command, args []string) error {
			db := mustGetContextDB(cmd)
			exists, err := db.AddDocActorRelationship(
				cmd.Context(),
				collectionArg,
				docIDArg,
				relationArg,
				targetActorArg,
			)

			if err != nil {
				return err
			}

			return writeJSON(cmd, exists)
		},
	}

	cmd.Flags().StringVarP(
		&collectionArg,
		collectionFlagLong,
		collectionFlagShort,
		"",
		"Collection that has the resource and policy for object",
	)
	_ = cmd.MarkFlagRequired(collectionFlagLong)

	cmd.Flags().StringVarP(
		&relationArg,
		relationFlagLong,
		relationFlagShort,
		"",
		"Relation that needs to be set for the relationship",
	)
	_ = cmd.MarkFlagRequired(relationFlagLong)

	cmd.Flags().StringVarP(
		&targetActorArg,
		targetActorFlagLong,
		targetActorFlagShort,
		"",
		"Actor to add relationship with",
	)
	_ = cmd.MarkFlagRequired(targetActorFlagLong)

	cmd.Flags().StringVarP(
		&docIDArg,
		docIDFlag,
		"",
		"",
		"Document Identifier (ObjectID) to make relationship for",
	)
	_ = cmd.MarkFlagRequired(docIDFlag)

	return cmd
}
