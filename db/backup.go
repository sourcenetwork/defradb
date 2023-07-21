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
	"github.com/sourcenetwork/defradb/datastore"
)

func (db *db) basicImport(ctx context.Context, txn datastore.Txn, filepath string) (err error) {
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
		col, err := db.getCollectionByName(ctx, txn, colName)
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

			delete(docMap, "_key")
			delete(docMap, "_newKey")

			doc, err := client.NewDocFromMap(docMap)
			if err != nil {
				return NewErrDocFromMap(err)
			}

			err = col.WithTxn(txn).Create(ctx, doc)
			if err != nil {
				return NewErrDocCreate(err)
			}
		}
		_, err = d.Token()
		if err != nil {
			return err
		}
	}

	return nil
}

func (db *db) basicExport(ctx context.Context, txn datastore.Txn, config *client.BackupConfig) (err error) {
	cols := []client.Collection{}
	if len(config.Collections) == 0 {
		cols, err = db.getAllCollections(ctx, txn)
		if err != nil {
			return NewErrFailedToGetAllCollections(err)
		}
	} else {
		for _, colName := range config.Collections {
			col, err := db.getCollectionByName(ctx, txn, colName)
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
			// add collection seperator
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
		colTxn := col.WithTxn(txn)
		keysCh, err := colTxn.GetAllDocKeys(ctx)
		if err != nil {
			return err
		}

		firstDoc := true
		for key := range keysCh {
			if firstDoc {
				firstDoc = false
			} else {
				// add document seperator
				err = writeString(f, ",", ",\n", config.Pretty)
				if err != nil {
					return err
				}
			}
			doc, err := colTxn.Get(ctx, key.Key, false)
			if err != nil {
				return err
			}

			docM, err := doc.ToMap()
			if err != nil {
				return err
			}

			// Temporary until https://github.com/sourcenetwork/defradb/issues/1681 is resolved.
			ensureIntIsInt(col.Schema().Fields, docM)

			delete(docM, "_key")
			newDoc, err := client.NewDocFromMap(docM)
			if err != nil {
				return err
			}
			// newKey is needed to let the user know what will be the key of the imported document.
			docM["_newKey"] = newDoc.Key().String()
			// NewDocFromMap removes the "_key" map item so we add it back.
			docM["_key"] = doc.Key().String()

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

// Temporary until https://github.com/sourcenetwork/defradb/issues/1681 is resolved.
func ensureIntIsInt(fields []client.FieldDescription, docMap map[string]any) {
	for _, field := range fields {
		if field.Kind == client.FieldKind_INT {
			if val, ok := docMap[field.Name]; ok {
				switch v := val.(type) {
				case uint64:
					docMap[field.Name] = int(v)
				case int64:
					docMap[field.Name] = int(v)
				}
			}
		}
	}
}
