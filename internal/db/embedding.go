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
	"strings"

	"github.com/philippgille/chromem-go"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/errors"
	"github.com/sourcenetwork/defradb/internal/keys"
)

const (
	ollama string = "ollama"
	openai string = "openai"
)

var supportedEmbeddingProviders = map[string]struct{}{
	ollama: {},
	openai: {},
}

func getEmbeddingFunc(provider, model, url string) chromem.EmbeddingFunc {
	var embedFunc chromem.EmbeddingFunc
	// No need to defend against unknown providers as it will have been
	// validated when creating/updating the embedding config.
	switch provider {
	case ollama:
		embedFunc = chromem.NewEmbeddingFuncOllama(model, url)
	case openai:
		normalized := true
		apiURL := url
		if apiURL == "" {
			apiURL = chromem.BaseURLOpenAI
		}
		embedFunc = chromem.NewEmbeddingFuncOpenAICompat(
			apiURL,
			os.Getenv("OPENAI_API_KEY"),
			model,
			&normalized,
		)
	}
	return embedFunc
}

// setEmbedding sets the embedding fields on the document if the related fields are dirty.
// However, if the vector field itself has been set, it will not be overwritten by a new embedding generation.
func (c *collection) setEmbedding(ctx context.Context, doc *client.Document, isCreate bool) error {
	embeddingGenerated := false
	for _, embedding := range c.Description().VectorEmbeddings {
		vecValue, err := doc.GetValue(embedding.FieldName)
		if err != nil && !errors.Is(err, client.ErrFieldNotExist) {
			return NewErrGetEmbeddingField(err)
		}
		if vecValue != nil && vecValue.IsDirty() {
			// If the vector has been explicitly set, no need to generate the embedding.
			continue
		}
		fieldsVal := make(map[string]client.NormalValue)
		needsGeneration := false
		missingFieldsForGeneration := []client.FieldDefinition{}

		// Get the new values of the fields used for embedding generation. We keep track
		// of the fields that aren't defined to lookup their previous values later.
		for _, embedField := range embedding.Fields {
			if docField, ok := doc.Fields()[embedField]; ok {
				if doc.Values()[docField].IsDirty() {
					needsGeneration = true
					fieldsVal[embedField] = doc.Values()[docField].NormalValue()
				} else {
					fieldDef, ok := c.def.GetFieldByName(embedField)
					if !ok {
						return NewErrEmbeddingFieldNotFound(embedField)
					}
					missingFieldsForGeneration = append(missingFieldsForGeneration, fieldDef)
				}
			}
		}

		if !needsGeneration {
			continue
		}

		// If we are updating the document and we don't have all the fields used for vector embedding
		// generation, we get the document to see if the fields have previously been set.
		if !isCreate && len(missingFieldsForGeneration) > 0 {
			oldDoc, err := c.get(
				ctx,
				keys.DataStoreKeyFromDocID(doc.ID()).ToPrimaryDataStoreKey(),
				missingFieldsForGeneration,
				false,
			)
			if err != nil {
				return NewErrGetDocForEmbedding(err)
			}
			for _, embedField := range missingFieldsForGeneration {
				if docField, ok := oldDoc.Fields()[embedField.Name]; ok {
					fieldsVal[embedField.Name] = oldDoc.Values()[docField].NormalValue()
				}
			}
		}

		embeddingFunc := getEmbeddingFunc(
			embedding.Provider,
			embedding.Model,
			embedding.URL,
		)

		var text strings.Builder
		for _, fieldName := range embedding.Fields {
			if val, ok := fieldsVal[fieldName]; ok {
				text.WriteString(fmt.Sprintf("%v\n", val.Unwrap()))
			}
		}
		embeddingVec, err := embeddingFunc(ctx, text.String())
		if err != nil {
			return err
		}
		err = doc.Set(embedding.FieldName, embeddingVec)
		if err != nil {
			return err
		}
		embeddingGenerated = true
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
