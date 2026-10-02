// Copyright 2026 Democratized Data Foundation
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
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/identity"
)

func MakeDocumentSaveCommand(ctx context.Context) *cobra.Command {
	var argDocID string
	var file string
	var shouldEncryptDoc bool
	var encryptedFields []string
	var enableSigning bool
	var cmd = &cobra.Command{
		Use:   "save [<document>]",
		Short: "Save a document (create or update).",
		Long: `Save a document. If a document with the given ID exists it will be updated,
otherwise a new document will be created. This operation is atomic.

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

			saveOpt := options.WithIdentity(
				options.SaveDocument().
					SetEncryptDoc(shouldEncryptDoc).
					SetEncryptedFields(encryptedFields),
				identity.FromContext(ctx),
			)
			// Bool flags are tri-state here: unset means use node config, false means disable.
			if cmd.Flags().Changed("enable-signing") {
				saveOpt.SetEnableSigning(enableSigning)
			}

			var doc *client.Document
			var err error
			if argDocID != "" {
				docID, err := client.NewDocIDFromString(argDocID)
				if err != nil {
					return NewErrParsingArgument("docID", err)
				}
				doc, err = client.NewDocWithID(ctx, docID, col.Version())
				if err != nil {
					return err
				}
				if len(docData) > 0 {
					if err := doc.SetWithJSON(ctx, docData); err != nil {
						return NewErrParsingArgument("document", err)
					}
				}
			} else {
				var rawMap map[string]any
				if err := json.Unmarshal(docData, &rawMap); err == nil && rawMap != nil {
					if _, hasDocID := rawMap[request.DocIDFieldName]; hasDocID {
						doc, err = client.NewDocFromMap(ctx, rawMap, col.Version())
						if err != nil {
							return NewErrParsingArgument("document", err)
						}
					}
				}
				if doc == nil {
					doc, err = client.NewDocFromJSON(ctx, docData, col.Version())
					if err != nil {
						return NewErrParsingArgument("document", err)
					}
				}
			}
			if err := col.SaveDocument(ctx, doc, saveOpt); err != nil {
				return err
			}
			return writeJSON(cmd, client.DocumentIDs([]*client.Document{doc}))
		},
	}

	EmbedCLIExample(ctx, cmd, "Save from string",
		`defradb client document save --collection-name User '{ "name": "Bob" }'`)

	EmbedCLIExample(ctx, cmd, "Save from string, with identity",
		`defradb client document save --collection-name User '{ "name": "Bob" }' \
  	-i 028d53f37a19afb9a0dbc5b4be30c65731479ee8cfa0c9bc8f8bf198cc3c075f`)

	EmbedCLIExample(ctx, cmd, "Save from file",
		`defradb client document save --collection-name User -f document.json`)

	EmbedCLIExample(ctx, cmd, "Save from stdin",
		`cat document.json | defradb client document save --collection-name User -`)

	cmd.PersistentFlags().BoolVarP(&shouldEncryptDoc, "encrypt", "e", false,
		"Flag to enable encryption of the document")
	cmd.PersistentFlags().StringSliceVar(&encryptedFields, "encrypt-fields", nil,
		"Comma-separated list of fields to encrypt")
	cmd.Flags().StringVarP(&file, "file", "f", "", "File containing document(s)")
	cmd.Flags().StringVar(&argDocID, "docID", "", "Document ID")
	cmd.Flags().BoolVar(&enableSigning, "enable-signing", false, "Override signing for this operation")
	setCollectionSelectorFlags(cmd)
	return cmd
}
