// Copyright 2023 Democratized Data Foundation
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
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/options"
	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/internal/datastore"
	"github.com/sourcenetwork/defradb/internal/db/description"
	"github.com/sourcenetwork/defradb/internal/db/id"
	"github.com/sourcenetwork/defradb/internal/utils"
)

func (db *DB) basicImport(ctx context.Context, filepath string) (err error) {
	f, err := os.Open(filepath)
	if err != nil {
		return NewErrOpenFile(err, filepath)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			err = NewErrCloseFile(closeErr, err)
		}
	}()

	d := json.NewDecoder(bufio.NewReader(f))

	t, err := d.Token()
	if err != nil {
		return err
	}
	if t != json.Delim('{') {
		return ErrExpectedJSONObject
	}
	importedDocIDs := map[string]string{}
	pendingRelations := []pendingImportRelation{}
	for d.More() {
		t, err := d.Token()
		if err != nil {
			return err
		}
		colName := t.(string)
		col, err := db.getCollectionByName(ctx, colName)
		if err != nil {
			return NewErrFailedToGetCollection(colName, err)
		}

		t, err = d.Token()
		if err != nil {
			return err
		}
		if t != json.Delim('[') {
			return ErrExpectedJSONArray
		}

		for d.More() {
			docMap := map[string]any{}
			err = d.Decode(&docMap)
			if err != nil {
				return NewErrJSONDecode(err)
			}

			docIDAliases := importDocIDAliases(docMap)
			deferredRelations := rewriteImportRelations(col, docMap, importedDocIDs)

			doc, err := client.NewDocFromMap(ctx, docMap, col.Version())
			if err != nil {
				return NewErrDocFromMap(err)
			}

			err = col.AddDocument(ctx, doc)
			if err != nil {
				return NewErrDocAdd(err)
			}
			for docID := range docIDAliases {
				err = db.setImportedDocIDAlias(ctx, col, doc, docID)
				if err != nil {
					return err
				}
			}

			docID := doc.ID().String()
			for docIDAlias := range docIDAliases {
				importedDocIDs[docIDAlias] = docID
			}
			for _, relation := range deferredRelations {
				relation.docID = docID
				if docID, ok := importedDocIDs[relation.importedDocID]; ok {
					err := updateImportedRelation(ctx, col, doc, relation.fieldName, docID)
					if err != nil {
						return err
					}
					continue
				}
				pendingRelations = append(pendingRelations, relation)
			}
		}
		_, err = d.Token()
		if err != nil {
			return err
		}
	}

	for _, relation := range pendingRelations {
		docID, ok := importedDocIDs[relation.importedDocID]
		if !ok {
			docID = relation.importedDocID
		}
		err := updatePendingImportRelation(ctx, relation, docID)
		if err != nil {
			return err
		}
	}

	return nil
}

type pendingImportRelation struct {
	col           client.Collection
	docID         string
	fieldName     string
	importedDocID string
}

func importDocIDAliases(docMap map[string]any) map[string]struct{} {
	aliases := map[string]struct{}{}
	if docID, ok := docMap[request.DocIDFieldName].(string); ok {
		aliases[docID] = struct{}{}
	}

	newDocID, hasNewDocID := docMap[request.NewDocIDFieldName]
	delete(docMap, request.NewDocIDFieldName)
	if hasNewDocID {
		if docID, ok := newDocID.(string); ok {
			aliases[docID] = struct{}{}
		}
		docMap[request.DocIDFieldName] = newDocID
	}
	return aliases
}

func rewriteImportRelations(
	col client.Collection,
	docMap map[string]any,
	importedDocIDs map[string]string,
) []pendingImportRelation {
	var deferred []pendingImportRelation
	for _, field := range col.Version().Fields {
		if !field.Kind.IsObject() || field.Kind.IsArray() {
			continue
		}
		for _, fieldName := range []string{field.Name, request.ToFieldID(field.Name)} {
			importedDocID, ok := docMap[fieldName].(string)
			if !ok {
				continue
			}
			if docID, ok := importedDocIDs[importedDocID]; ok {
				docMap[fieldName] = docID
				continue
			}
			delete(docMap, fieldName)
			deferred = append(deferred, pendingImportRelation{
				col:           col,
				fieldName:     fieldName,
				importedDocID: importedDocID,
			})
		}
	}
	return deferred
}

func updateImportedRelation(
	ctx context.Context,
	col client.Collection,
	doc *client.Document,
	fieldName string,
	docID string,
) error {
	err := doc.Set(ctx, fieldName, docID)
	if err != nil {
		return NewErrDocUpdate(err)
	}
	err = col.UpdateDocument(ctx, doc)
	if err != nil {
		return NewErrDocUpdate(err)
	}
	return nil
}

func updatePendingImportRelation(ctx context.Context, relation pendingImportRelation, docID string) error {
	relationDocID, err := client.NewDocIDFromString(relation.docID)
	if err != nil {
		return err
	}
	doc, err := relation.col.GetDocument(ctx, relationDocID)
	if err != nil {
		return err
	}
	return updateImportedRelation(ctx, relation.col, doc, relation.fieldName, docID)
}

func (db *DB) setImportedDocIDAlias(
	ctx context.Context,
	col client.Collection,
	doc *client.Document,
	importedDocID string,
) error {
	if importedDocID == "" || importedDocID == doc.ID().String() {
		return nil
	}

	ctx, txn, err := ensureContextTxn(ctx, db, false)
	if err != nil {
		return err
	}
	defer txn.Discard()

	colShortID, err := id.GetCollectionShortID(ctx, col.CollectionID())
	if err != nil {
		return err
	}
	docShortID, found, err := id.GetDocShortID(ctx, colShortID, doc.ID().String())
	if err != nil {
		return err
	}
	if !found {
		return NewErrDocIDNotFound(doc.ID().String())
	}
	if err := id.SetDocIDToDocRefMapping(ctx, colShortID, docShortID, importedDocID); err != nil {
		return err
	}
	return txn.Commit()
}

func (db *DB) basicExport(ctx context.Context, config *client.BackupConfig) (err error) {
	cols := []client.Collection{}
	if len(config.Collections) == 0 {
		cols, err = db.getCollections(ctx, utils.NewOptions(options.GetCollections()), true)
		if err != nil {
			return NewErrFailedToGetAllCollections(err)
		}
	} else {
		for _, colName := range config.Collections {
			col, err := db.getCollectionByName(ctx, colName)
			if err != nil {
				return NewErrFailedToGetCollection(colName, err)
			}
			cols = append(cols, col)
		}
	}

	tempFile := config.Filepath + ".temp"
	f, err := os.Create(tempFile)
	if err != nil {
		return NewErrCreateFile(err, tempFile)
	}
	defer func() {
		closeErr := f.Close()
		if closeErr != nil {
			err = NewErrCloseFile(closeErr, err)
		} else if err != nil {
			// ensure we cleanup if there was an error
			removeErr := os.Remove(tempFile)
			if removeErr != nil {
				err = NewErrRemoveFile(removeErr, err, tempFile)
			}
		} else {
			_ = os.Rename(tempFile, config.Filepath)
		}
	}()

	// open the object
	err = writeString(f, "{", "{\n", config.Pretty)
	if err != nil {
		return err
	}

	firstCol := true
	for _, col := range cols {
		if firstCol {
			firstCol = false
		} else {
			// add collection separator
			err = writeString(f, ",", ",\n", config.Pretty)
			if err != nil {
				return err
			}
		}

		// set collection
		err = writeString(
			f,
			fmt.Sprintf("\"%s\":[", col.Name()),
			fmt.Sprintf("  \"%s\": [\n", col.Name()),
			config.Pretty,
		)
		if err != nil {
			return err
		}
		docIDsCh, err := col.(*collection).getAllDocIDsChan(ctx)
		if err != nil {
			return err
		}

		firstDoc := true
		for docResultWithID := range docIDsCh {
			if firstDoc {
				firstDoc = false
			} else {
				// add document separator
				err = writeString(f, ",", ",\n", config.Pretty)
				if err != nil {
					return err
				}
			}
			doc, err := col.GetDocument(ctx, docResultWithID.ID)
			if err != nil {
				return err
			}

			isSelfReference := false
			refFieldName := ""
			// Clear references to missing documents and hold self-references until after add.
			for _, field := range col.Version().Fields {
				if field.Kind.IsObject() && !field.Kind.IsArray() {
					fieldID := request.ToFieldID(field.Name)
					if foreignKey, err := doc.Get(fieldID); err == nil {
						foreignDef, _, err := description.GetRelatedCollection(ctx, db.collectionRepository, col.Version(), field.Kind)
						if err != nil {
							return err
						}

						txnOpt := datastore.CtxTryGetTxnOption(ctx)
						foreignCol, err := db.newCollection(ctx, foreignDef, txnOpt)
						if err != nil {
							return err
						}

						foreignDocID, err := client.NewDocIDFromString(foreignKey.(string))
						if err != nil {
							return err
						}
						foreignDoc, err := foreignCol.GetDocument(ctx, foreignDocID)
						if err != nil {
							err := doc.Set(ctx, request.ToFieldID(field.Name), nil)
							if err != nil {
								return err
							}
						} else if foreignDoc.ID().String() == doc.ID().String() {
							isSelfReference = true
							refFieldName = fieldID
						}
					}
				}
			}

			docM, err := doc.ToMap()
			if err != nil {
				return err
			}

			delete(docM, request.DocIDFieldName)
			if isSelfReference {
				delete(docM, refFieldName)
			}

			docM[request.NewDocIDFieldName] = doc.ID().String()
			docM[request.DocIDFieldName] = doc.ID().String()

			if isSelfReference {
				docM[refFieldName] = doc.ID().String()
			}

			var b []byte
			if config.Pretty {
				_, err = f.WriteString("    ")
				if err != nil {
					return NewErrFailedToWriteString(err)
				}
				b, err = json.MarshalIndent(docM, "    ", "  ")
				if err != nil {
					return NewErrFailedToWriteString(err)
				}
			} else {
				b, err = json.Marshal(docM)
				if err != nil {
					return err
				}
			}

			// write document
			_, err = f.Write(b)
			if err != nil {
				return err
			}
		}

		// close collection
		err = writeString(f, "]", "\n  ]", config.Pretty)
		if err != nil {
			return err
		}
	}

	// close object
	err = writeString(f, "}", "\n}", config.Pretty)
	if err != nil {
		return err
	}

	err = f.Sync()
	if err != nil {
		return err
	}

	return nil
}

func writeString(f *os.File, normal, pretty string, isPretty bool) error {
	if isPretty {
		_, err := f.WriteString(pretty)
		if err != nil {
			return NewErrFailedToWriteString(err)
		}
		return nil
	}

	_, err := f.WriteString(normal)
	if err != nil {
		return NewErrFailedToWriteString(err)
	}
	return nil
}
