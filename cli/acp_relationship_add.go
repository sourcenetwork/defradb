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
		Use:   "add [--docID] [-c --collection] [-r --relation] [-a --actor] [-i --identity]",
		Short: "Add new relationship",
		Long: `Add new relationship

To share a document (or grant a more restricted access) with another actor, we must add a relationship between the
actor and the document. Inorder to make the relationship we require all of the following:
1) Target DocID: The docID of the document we want to make a relationship for.
2) Collection Name: The name of the collection that has the Target DocID.
3) Relation Name: The type of relation (name must be defined within the linked policy on collection).
4) Target Identity: The identity of the actor the relationship is being made with.
5) Requesting Identity: The identity of the actor that is making the request.

Notes:
  - ACP must be available (i.e. ACP can not be disabled).
  - The target document must be registered with ACP already (policy & resource specified).
  - The requesting identity MUST either be the owner OR the manager (manages the relation) of the resource.
  - If the specified relation was not granted the miminum DPI permissions (read or write) within the policy,
  and a relationship is formed, the subject/actor will still not be able to access (read or write) the resource.
  - Learn more about [ACP & DPI Rules](/acp/README.md)

Example: Let another actor (4d092126012ebaf56161716018a71630d99443d9d5217e9d8502bb5c5456f2c5) read a private document:
  defradb client acp relationship add \
	--collection Users \
	--docID bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c \
	--relation reader \
	--actor did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn \
	--identity e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac

Example: Creating a dummy relationship does nothing (from database prespective):
  defradb client acp relationship add \
	-c Users \
	--docID bae-ff3ceb1c-b5c0-5e86-a024-dd1b16a4261c \
	-r dummy \
	-a did:key:z7r8os2G88XXBNBTLj3kFR5rzUJ4VAesbX7PgsA68ak9B5RYcXF5EZEmjRzzinZndPSSwujXb4XKHG6vmKEFG6ZfsfcQn \
	-i e3b722906ee4e56368f581cd8b18ab0f48af1ea53e635e3f7b8acd076676f6ac
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
