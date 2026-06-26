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
	"context"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeDocumentAddCommand(ctx context.Context) *cobra.Command {
	var file string
	var shouldEncryptDoc bool
	var encryptedFields []string
	var returnIDs bool
	var enableSigning bool
	var cmd = &cobra.Command{
		Use:   "add [<document>]",
		Short: "Add a new document.",
		Long: `Add a new document.

Options:
	-i, --identity
		Marks the document as private and set the identity as the owner. The access to the document
		and permissions are controlled by ACP (Access Control Policy).

	-e, --encrypt
		Encrypt flag specified if the document needs to be encrypted. If set, DefraDB will generate a
		symmetric key for encryption using AES-GCM.

	--encrypt-fields
		Comma-separated list of fields to encrypt. If set, DefraDB will encrypt only the specified fields
		and for every field in the list it will generate a symmetric key for encryption using AES-GCM.
		If combined with '--encrypt' flag, all the fields in the document not listed in '--encrypt-fields'
		will be encrypted with the same key.
		`,
		Args: cobra.RangeArgs(0, 1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var docData []byte
			switch {
			case file != "":
				data, err := os.ReadFile(file)
				if err != nil {
					return NewErrReadingArgument("file", err)
				}
				docData = data
			case len(args) == 1 && args[0] == "-":
				data, err := io.ReadAll(cmd.InOrStdin())
				if err != nil {
					return NewErrReadingArgument("stdin", err)
				}
				docData = data
			case len(args) == 1:
				docData = []byte(args[0])
			default:
				return ErrNoDocOrFile
			}

			col, ok := tryGetContextCollection(cmd)
			if !ok {
				return client.ErrCollectionNotFound
			}

			ctx := cmd.Context()

			addOpt := options.WithIdentity(
				options.AddDocument().
					SetEncryptDoc(shouldEncryptDoc).
					SetEncryptedFields(encryptedFields),
				identity.FromContext(ctx),
			)
			// The flag is optional; false only means "disable" when explicitly provided.
			if cmd.Flags().Changed("enable-signing") {
				addOpt.SetEnableSigning(enableSigning)
			}

			if client.IsJSONArray(docData) {
				docs, err := client.NewDocsFromJSON(ctx, docData, col.Version())
				if err != nil {
					return NewErrParsingArgument("document", err)
				}
				if err := col.AddManyDocuments(ctx, docs, addOpt); err != nil {
					return err
				}
				if returnIDs {
					return writeJSON(cmd, client.DocumentIDs(docs))
				}
				return nil
			}

			doc, err := client.NewDocFromJSON(ctx, docData, col.Version())
			if err != nil {
				return NewErrParsingArgument("document", err)
			}
			if err := col.AddDocument(cmd.Context(), doc, addOpt); err != nil {
				return err
			}
			if returnIDs {
				return writeJSON(cmd, client.DocumentIDs([]*client.Document{doc}))
			}
			return nil
		},
	}

	EmbedCLIExample(ctx, cmd, "Add from string1",
		`defradb client document add --collection-name User '{ "name": "Bob" }'`)

	EmbedCLIExample(ctx, cmd, "Add from string, with identity",
		`defradb client document add --collection-name User '{ "name": "Bob" }' \
  	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f`)

	EmbedCLIExample(ctx, cmd, "Add multiple from string",
		`defradb client document add --collection-name User '[{ "name": "Alice" }, { "name": "Bob" }]'`)

	EmbedCLIExample(ctx, cmd, "Add from file",
		`defradb client document add --collection-name User -f document.json`)

	EmbedCLIExample(ctx, cmd, "Add from stdin",
		`cat document.json | defradb client document add --collection-name User -`)

	cmd.PersistentFlags().BoolVarP(&shouldEncryptDoc, "encrypt", "e", false,
		"Flag to enable encryption of the document")
	cmd.PersistentFlags().StringSliceVar(&encryptedFields, "encrypt-fields", nil,
		"Comma-separated list of fields to encrypt")
	cmd.Flags().StringVarP(&file, "file", "f", "", "File containing document(s)")
	// Used by the CLI client adapter to copy save-assigned IDs back onto caller docs.
	cmd.Flags().BoolVar(&returnIDs, "return-ids", false, "Return generated document IDs")
	cmd.Flags().Lookup("return-ids").Hidden = true
	cmd.Flags().BoolVar(&enableSigning, "enable-signing", false, "Override signing for this operation")
	setCollectionSelectorFlags(cmd)
	return cmd
}
