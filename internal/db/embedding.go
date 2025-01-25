// Copyright 2025 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package db

import (
	"context"
	"fmt"
	"os"

	"github.com/philippgille/chromem-go"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/keys"
)

func getEmbeddingFunc(provider, model, url string) (chromem.EmbeddingFunc, error) {
	switch provider {
	case "ollama":
		return chromem.NewEmbeddingFuncOllama(model, url), nil
	case "openai":
		normalized := true
		apiURL := url
		if apiURL == "" {
			apiURL = chromem.BaseURLOpenAI
		}
		return chromem.NewEmbeddingFuncOpenAICompat(
			apiURL,
			os.Getenv("OPENAI_API_KEY"),
			model,
			&normalized,
		), nil
	default:
		return nil, NewErrUnknownEmbeddingProvider(provider)
	}
}

// setEmbedding sets the embedding fields on the document if the related fields are dirty.
func (c *collection) setEmbedding(ctx context.Context, doc *client.Document, isCreate bool) error {
	embeddingGenerated := false
	for _, colField := range c.Definition().GetFields() {
		if colField.Embedding != nil {
			fieldsVal := make(map[string]client.NormalValue)
			needsGeneration := false
			missingFieldsForGeneration := []client.FieldDefinition{}
			for _, embedField := range colField.Embedding.Fields {
				if docField, ok := doc.Fields()[embedField]; ok {
					if doc.Values()[docField].IsDirty() {
						needsGeneration = true
						fieldsVal[embedField] = doc.Values()[docField].NormalValue()
					} else {
						fieldDef, ok := c.def.GetFieldByName(embedField)
						if !ok {
							return errors.New("field not found", errors.NewKV("field", embedField))
						}
						missingFieldsForGeneration = append(missingFieldsForGeneration, fieldDef)
					}
				}
			}
			if needsGeneration && len(missingFieldsForGeneration) > 0 {
				oldDoc, err := c.get(
					ctx,
					keys.DataStoreKeyFromDocID(doc.ID()).ToPrimaryDataStoreKey(),
					missingFieldsForGeneration,
					false,
				)
				if err != nil {
					return err
				}
				for _, embedField := range missingFieldsForGeneration {
					if docField, ok := oldDoc.Fields()[embedField.Name]; ok {
						fieldsVal[embedField.Name] = oldDoc.Values()[docField].NormalValue()
					}
				}
			}
			if needsGeneration {
				embeddingFunc, err := getEmbeddingFunc(
					colField.Embedding.Provider,
					colField.Embedding.Model,
					colField.Embedding.URL,
				)
				if err != nil {
					return err
				}

				text := ""
				for _, fieldName := range colField.Embedding.Fields {
					if val, ok := fieldsVal[fieldName]; ok {
						text += fmt.Sprintf("%v\n", val.Unwrap())
					}
				}
				embedding, err := embeddingFunc(ctx, text)
				if err != nil {
					return err
				}
				err = doc.Set(colField.Name, embedding)
				if err != nil {
					return err
				}
				embeddingGenerated = true
			}
		}
	}

	// If an embedding was generated on create, we need to update the document ID.
	if isCreate && embeddingGenerated {
		err := doc.GenerateAndSetDocID()
		if err != nil {
			return err
		}
	}

	return nil
}
