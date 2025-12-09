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
	"io"
	"os"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/request/graphql/schema"

	gql "github.com/sourcenetwork/graphql-go"
	"github.com/spf13/cobra"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/astprinter"
	"github.com/wundergraph/graphql-go-tools/v2/pkg/introspection"
)

var (
	defaultOutputPath = "schema.gen.graphql"
	fileLineSeperator = []byte("\n\n")
)

func MakeSDLGenerateCommand(ctx context.Context) *cobra.Command {
	var outputFile string
	var cmd = &cobra.Command{
		Use:   "generate --output schema.graphql <input schema files...>",
		Short: "Generate full GraphQL formatted schema.",
		Long: `Generates the fully formatted GraphQL schema from a given user type definition.

		Accepts multiple input files as well as "-" to use stdin.
		`,
		Args: cobra.ArbitraryArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			var inputBuf io.Reader
			fileInputBuf := bytes.NewBuffer(nil)

			// Either we use stdin or we concat all the file
			// arguments
			if len(args) == 1 && args[0] == "-" {
				inputBuf = os.Stdin
			} else if len(args) > 0 {
				for i, arg := range args {
					if arg == "-" {
						return errors.New("stdin only allowed as single input ")
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
			} else {
				return errors.New("input can't be empty")
			}

			schemaManager, err := schema.NewSchemaManager(false)
			if err != nil {
				return err
			}

			sdlBuf, err := io.ReadAll(inputBuf)
			if err != nil {
				return err
			}

			cols, err := schemaManager.ParseSDL(string(sdlBuf))
			if err != nil {
				return err
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
				return err
			}

			params := gql.Params{Schema: *schemaManager.Schema(), RequestString: string(introspectionQueryRequest)}
			r := gql.Do(params)
			if len(r.Errors) != 0 {
				return errors.New(r.Errors[0].Error())
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
				f, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, 0644)
				if err != nil {
					return err
				}
				defer f.Close()
				outWriter = f
			}

			err = astprinter.PrintIndent(doc, []byte("    "), outWriter)
			if err != nil {
				return err
			}

			return nil
		},
	}

	cmd.PersistentFlags().StringVarP(&outputFile, "output", "o", defaultOutputPath,
		"The output file to write the generated schema. Accepts '-' to write to stdout")

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
