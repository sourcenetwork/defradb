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

	acpIdentity "github.com/sourcenetwork/defradb/acp/identity"
	"github.com/sourcenetwork/defradb/client"
	"github.com/sourcenetwork/defradb/client/request"
)

func (db *db) basicImport(ctx context.Context, filepath string) (err error) {
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

			// check if self referencing and remove from docMap for key creation
			resetMap := map[string]any{}
			for _, field := range col.Schema().Fields {
				if field.Kind.IsObject() && !field.Kind.IsArray() {
					if val, ok := docMap[field.Name+request.RelatedObjectID]; ok {
						if docMap[request.NewDocIDFieldName] == val {
							resetMap[field.Name+request.RelatedObjectID] = val
							delete(docMap, field.Name+request.RelatedObjectID)
						}
					}
				}
			}

			delete(docMap, request.DocIDFieldName)
			delete(docMap, request.NewDocIDFieldName)

			doc, err := client.NewDocFromMap(docMap, col.Schema())
			if err != nil {
				return NewErrDocFromMap(err)
			}

			// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2430 - Add identity ability to backup
			err = col.Create(ctx, acpIdentity.NoIdentity, doc)
			if err != nil {
				return NewErrDocCreate(err)
			}

			// add back the self referencing fields and update doc.
			for k, v := range resetMap {
				err := doc.Set(k, v)
				if err != nil {
					return NewErrDocUpdate(err)
				}
				// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2430 - Add identity ability to backup
				err = col.Update(ctx, acpIdentity.NoIdentity, doc)
				if err != nil {
					return NewErrDocUpdate(err)
				}
			}
		}
		_, err = d.Token()
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) basicExport(ctx context.Context, config *client.BackupConfig) (err error) {
	// old key -> new Key
	keyChangeCache := map[string]string{}

	cols := []client.Collection{}
	if len(config.Collections) == 0 {
		cols, err = db.getCollections(ctx, client.CollectionFetchOptions{})
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
	colNameCache := map[string]struct{}{}
	for _, col := range cols {
		colNameCache[col.Name().Value()] = struct{}{}
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
			fmt.Sprintf("\"%s\":[", col.Name().Value()),
			fmt.Sprintf("  \"%s\": [\n", col.Name().Value()),
			config.Pretty,
		)
		if err != nil {
			return err
		}
		// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2430 - Add identity ability to export
		docIDsCh, err := col.GetAllDocIDs(ctx, acpIdentity.NoIdentity)
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
			// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2430 - Add identity ability to export
			doc, err := col.Get(ctx, acpIdentity.NoIdentity, docResultWithID.ID, false)
			if err != nil {
				return err
			}

			isSelfReference := false
			refFieldName := ""
			// replace any foreign key if it needs to be changed
			for _, field := range col.Schema().Fields {
				if field.Kind.IsObject() && !field.Kind.IsArray() {
					if _, ok := colNameCache[field.Kind.Underlying()]; !ok {
						continue
					}
					if foreignKey, err := doc.Get(field.Name + request.RelatedObjectID); err == nil {
						if newKey, ok := keyChangeCache[foreignKey.(string)]; ok {
							err := doc.Set(field.Name+request.RelatedObjectID, newKey)
							if err != nil {
								return err
							}
							if foreignKey.(string) == doc.ID().String() {
								isSelfReference = true
								refFieldName = field.Name + request.RelatedObjectID
							}
						} else {
							foreignCol, err := db.getCollectionByName(ctx, field.Kind.Underlying())
							if err != nil {
								return NewErrFailedToGetCollection(field.Kind.Underlying(), err)
							}
							foreignDocID, err := client.NewDocIDFromString(foreignKey.(string))
							if err != nil {
								return err
							}
							// TODO-ACP: https://github.com/sourcenetwork/defradb/issues/2430
							foreignDoc, err := foreignCol.Get(ctx, acpIdentity.NoIdentity, foreignDocID, false)
							if err != nil {
								err := doc.Set(field.Name+request.RelatedObjectID, nil)
								if err != nil {
									return err
								}
							} else {
								oldForeignDoc, err := foreignDoc.ToMap()
								if err != nil {
									return err
								}

								delete(oldForeignDoc, request.DocIDFieldName)
								if foreignDoc.ID().String() == foreignDocID.String() {
									delete(oldForeignDoc, field.Name+request.RelatedObjectID)
								}

								if foreignDoc.ID().String() == doc.ID().String() {
									isSelfReference = true
									refFieldName = field.Name + request.RelatedObjectID
								}

								newForeignDoc, err := client.NewDocFromMap(oldForeignDoc, foreignCol.Schema())
								if err != nil {
									return err
								}

								if foreignDoc.ID().String() != doc.ID().String() {
									err = doc.Set(field.Name+request.RelatedObjectID, newForeignDoc.ID().String())
									if err != nil {
										return err
									}
								}

								if newForeignDoc.ID().String() != foreignDoc.ID().String() {
									keyChangeCache[foreignDoc.ID().String()] = newForeignDoc.ID().String()
								}
							}
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

			newDoc, err := client.NewDocFromMap(docM, col.Schema())
			if err != nil {
				return err
			}
			// a new docID is needed to let the user know what will be the docID of the imported document.
			docM[request.NewDocIDFieldName] = newDoc.ID().String()
			// NewDocFromMap removes the "_docID" map item so we add it back.
			docM[request.DocIDFieldName] = doc.ID().String()

			if isSelfReference {
				docM[refFieldName] = newDoc.ID().String()
			}

			if newDoc.ID().String() != doc.ID().String() {
				keyChangeCache[doc.ID().String()] = newDoc.ID().String()
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
