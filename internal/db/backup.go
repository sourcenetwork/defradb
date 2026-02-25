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
	"github.com/sourcenetwork/defradb/internal/db/description"
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
			for _, field := range col.Version().Fields {
				if field.Kind.IsObject() && !field.Kind.IsArray() {
					fieldID := request.ToFieldID(field.Name)
					if val, ok := docMap[fieldID]; ok {
						if docMap[request.NewDocIDFieldName] == val {
							resetMap[fieldID] = val
							delete(docMap, fieldID)
						}
					}
				}
			}

			delete(docMap, request.DocIDFieldName)
			delete(docMap, request.NewDocIDFieldName)

			doc, err := client.NewDocFromMap(ctx, docMap, col.Version())
			if err != nil {
				return NewErrDocFromMap(err)
			}

			err = col.Add(ctx, doc)
			if err != nil {
				return NewErrDocAdd(err)
			}

			// add back the self referencing fields and update doc.
			for k, v := range resetMap {
				err := doc.Set(ctx, k, v)
				if err != nil {
					return NewErrDocUpdate(err)
				}
				err = col.Update(ctx, doc)
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

func (db *DB) basicExport(ctx context.Context, config *client.BackupConfig) (err error) {
	// old key -> new Key
	keyChangeCache := map[string]string{}

	cols := []client.Collection{}
	if len(config.Collections) == 0 {
		cols, err = db.getCollections(ctx, utils.NewOptions(options.GetCollections()))
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
			doc, err := col.Get(ctx, docResultWithID.ID)
			if err != nil {
				return err
			}

			isSelfReference := false
			refFieldName := ""
			// replace any foreign key if it needs to be changed
			for _, field := range col.Version().Fields {
				if field.Kind.IsObject() && !field.Kind.IsArray() {
					fieldID := request.ToFieldID(field.Name)
					if foreignKey, err := doc.Get(fieldID); err == nil {
						if newKey, ok := keyChangeCache[foreignKey.(string)]; ok {
							err := doc.Set(ctx, request.ToFieldID(field.Name), newKey)
							if err != nil {
								return err
							}
							if foreignKey.(string) == doc.ID().String() {
								isSelfReference = true
								refFieldName = fieldID
							}
						} else {
							foreignDef, _, err := description.GetRelatedCollection(ctx, col.Version(), field.Kind)
							if err != nil {
								return err
							}

							foreignCol, err := db.newCollection(foreignDef)
							if err != nil {
								return err
							}

							foreignDocID, err := client.NewDocIDFromString(foreignKey.(string))
							if err != nil {
								return err
							}
							foreignDoc, err := foreignCol.Get(ctx, foreignDocID)
							if err != nil {
								err := doc.Set(ctx, request.ToFieldID(field.Name), nil)
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
									delete(oldForeignDoc, fieldID)
								}

								if foreignDoc.ID().String() == doc.ID().String() {
									isSelfReference = true
									refFieldName = fieldID
								}

								newForeignDoc, err := client.NewDocFromMap(ctx, oldForeignDoc, foreignCol.Version())
								if err != nil {
									return err
								}

								if foreignDoc.ID().String() != doc.ID().String() {
									err = doc.Set(ctx, request.ToFieldID(field.Name), newForeignDoc.ID().String())
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

			newDoc, err := client.NewDocFromMap(ctx, docM, col.Version())
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
