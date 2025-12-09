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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astprinter"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"

	gql "github.com/sourcenetwork/graphql-go"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"
)

var (
	defaultOutputPath = "schema.gen.graphql"
	fileLineSeperator = []byte("\n\n")
)

func MakeSDLGenerateCommand(ctx context.Context) *cobra.Command {
	var outputFile string
	var yesOverwrite bool
	var searchableEncryption bool
	var cmd = &cobra.Command{
		Use:   "generate --output schema.graphql <input schema files...>",
		Short: "Generate full GraphQL formatted schema.",
		Long: `Generates the fully formatted GraphQL schema from a given user type definition(s).

		Accepts multiple input files as well as "-" to use stdin.
		`,
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var inputBuf io.Reader

			// Either we use stdin or we concat all the file
			// arguments
			if len(args) == 1 && args[0] == "-" {
				inputBuf = os.Stdin
			} else {
				fileInputBuf := bytes.NewBuffer(nil)
				for i, arg := range args {
					if arg == "-" {
						return ErrStdinSingleInputOnly
					}
					fileBuf, err := os.ReadFile(arg)
					if err != nil {
						return err
					}

					if i != 0 {
						fileInputBuf.Write(fileLineSeperator)
					}
					fileInputBuf.Write(fileBuf)
				}
				inputBuf = fileInputBuf
			}

			schemaManager, err := schema.NewSchemaManager(searchableEncryption)
			if err != nil {
				return err
			}

			sdlBuf, err := io.ReadAll(inputBuf)
			if err != nil {
				return errors.Join(ErrReadingInput, err)
			}

			cols, err := schemaManager.ParseSDL(string(sdlBuf))
			if err != nil {
				return errors.Join(ErrParsingSDL, err)
			}

			collections := make([]client.CollectionVersion, len(cols))
			for i, c := range cols {
				collections[i] = c.Definition
			}

			cache := description.NewCollectionCache()
			cache.AddAll(collections)
			ctx := description.WithCollectionCache(ctx, cache)

			_, err = schemaManager.Generator.Generate(ctx, collections)
			if err != nil {
				return errors.Join(ErrGeneratingSDL, err)
			}

			params := gql.Params{Schema: *schemaManager.Schema(), RequestString: introspectionQueryRequest}
			r := gql.Do(params)
			if len(r.Errors) != 0 {
				// for simplicity we're just going to return the
				// first error, if there are more, they'll be caught on
				// follow up invocations.
				return errors.Join(ErrGeneratingSDL, r.Errors[0])
			}

			respJson, err := json.Marshal(r.Data)
			if err != nil {
				return err
			}
			respBuf := bytes.NewBuffer(respJson)

			converter := introspection.JsonConverter{}
			doc, err := converter.GraphQLDocument(respBuf)
			if err != nil {
				return err
			}

			var outWriter io.Writer
			if outputFile == "-" {
				outWriter = os.Stdout
			} else {
				// check if the file exists, if so check for the overwrite
				// flag
				ofinfo, err := os.Stat(outputFile)
				if err != nil && !errors.Is(err, os.ErrNotExist) {
					return err
				}
				if ofinfo != nil && !yesOverwrite {
					fmt.Fprintln(os.Stderr, "output file path already exists. If you want to overwrite use -y")
					os.Exit(1)
				}

				f, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil {
					return err
				}
				defer f.Close() //nolint:errcheck
				outWriter = f
			}

			err = astprinter.PrintIndent(doc, []byte("    "), outWriter)
			if err != nil {
				return errors.Join(ErrWritingOutput, err)
			}

			return nil
		},
	}

	EmbedCLIExample(ctx, cmd, "Generate SDL",
		`defradb sdl generate foo.graphql`)

	EmbedCLIExample(ctx, cmd, "Generate Multiple SDLs",
		`defradb sdl generate foo.graphql bar.graphql`)

	EmbedCLIExample(ctx, cmd, "Generate SDL and overwrite output",
		`defradb sdl generate foo.graphql bar.graphql --output schema.graphql -y`)

	cmd.PersistentFlags().StringVarP(&outputFile, "output", "o", defaultOutputPath,
		"The output file to write the generated schema. Accepts '-' to write to stdout")

	EmbedCLIExample(ctx, cmd, "Generate SDL with Searchable Encryption type definitions",
		`defradb sdl generate foo.graphql -s`)

	cmd.PersistentFlags().BoolVarP(&yesOverwrite, "overwrite", "y", false,
		"Overwrite any existing matching output file paths")

	cmd.PersistentFlags().BoolVarP(&searchableEncryption, "include-searchable-encryption", "s",
		false, "Include the schema type definitions to support Searchable Encryption")

	return cmd
}

var introspectionQueryRequest = `

    query IntrospectionQuery {
      __schema {
        
        queryType { name }
        mutationType { name }
        subscriptionType { name }
        types {
          ...FullType
        }
        directives {
          name
          description
          
          locations
          args {
            ...InputValue
          }
        }
      }
    }

    fragment FullType on __Type {
      kind
      name
      description
      
      
      fields(includeDeprecated: true) {
        name
        description
        args {
          ...InputValue
        }
        type {
          ...TypeRef
        }
        isDeprecated
        deprecationReason
      }
      inputFields {
        ...InputValue
      }
      interfaces {
        ...TypeRef
      }
      enumValues(includeDeprecated: true) {
        name
        description
        isDeprecated
        deprecationReason
      }
      possibleTypes {
        ...TypeRef
      }
    }

    fragment InputValue on __InputValue {
      name
      description
      type { ...TypeRef }
      defaultValue
      
      
    }

    fragment TypeRef on __Type {
      kind
      name
      ofType {
        kind
        name
        ofType {
          kind
          name
          ofType {
            kind
            name
            ofType {
              kind
              name
              ofType {
                kind
                name
                ofType {
                  kind
                  name
                  ofType {
                    kind
                    name
                    ofType {
                      kind
                      name
                      ofType {
                        kind
                        name
                      }
                    }
                  }
                }
              }
            }
          }
        }
      }
    }
  
`
