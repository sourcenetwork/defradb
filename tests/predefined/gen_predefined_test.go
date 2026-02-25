// Copyright 2023 Democratized Data Foundation
//
// Use of this software is governed by the Business Source License
// included in the file licenses/BSL.txt.
//
// As of the Change Date specified in that file, in accordance with
// the Business Source License, use of this software will be governed
// by the Apache License, Version 2.0, included in the file
// licenses/APL.txt.

package predefined

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sourcenetwork/defradb/client/request"
	"github.com/sourcenetwork/defradb/tests/gen"
)

func TestGeneratePredefinedFromSchema_Simple(t *testing.T) {
	ctx := context.Background()
	schema := `
		type User {
			name: String
			age: Int
		}`

	docsList := DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{"name": "John", "age": 30},
			{"name": "Fred", "age": 25},
		},
	}
	docs, err := AddFromSDL(ctx, schema, docsList)
	assert.NoError(t, err)

	colDefMap, err := gen.ParseSDL(schema)
	require.NoError(t, err)

	errorMsg := assertDocs(mustAddDocIDsToDocs(ctx, docsList.Docs, colDefMap["User"]), docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_StripExcessiveFields(t *testing.T) {
	ctx := context.Background()
	schema := `
		type User {
			name: String
		}`

	docs, err := AddFromSDL(ctx, schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{"name": "John", "age": 30},
			{"name": "Fred", "age": 25},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := gen.ParseSDL(schema)
	require.NoError(t, err)

	errorMsg := assertDocs(mustAddDocIDsToDocs(ctx, []map[string]any{
		{"name": "John"},
		{"name": "Fred"},
	}, colDefMap["User"]), docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_OneToOne(t *testing.T) {
	ctx := context.Background()
	schema := `
		type User {
			name: String
			device: Device
		}
		type Device {
			model: String
			owner: User @primary
		}`

	docs, err := AddFromSDL(ctx, schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
				},
			},
			{
				"name": "Fred",
				"device": map[string]any{
					"model": "MacBook",
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := gen.ParseSDL(schema)
	require.NoError(t, err)

	userDocs := mustAddDocIDsToDocs(ctx, []map[string]any{
		{"name": "John"},
		{"name": "Fred"},
	}, colDefMap["User"])

	deviceDocs := mustAddDocIDsToDocs(ctx, []map[string]any{
		{
			"model":    "iPhone",
			"_ownerID": mustGetDocIDFromDocMap(ctx, map[string]any{"name": "John"}, colDefMap["User"]),
		},
		{
			"model":    "MacBook",
			"_ownerID": mustGetDocIDFromDocMap(ctx, map[string]any{"name": "Fred"}, colDefMap["User"]),
		},
	}, colDefMap["Device"])

	errorMsg := assertDocs(append(userDocs, deviceDocs...), docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_OneToOnePrimary(t *testing.T) {
	ctx := context.Background()
	schema := `
		type User {
			name: String
			device: Device @primary
		}
		type Device {
			model: String
			owner: User
		}`

	docs, err := AddFromSDL(ctx, schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
				},
			},
			{
				"name": "Fred",
				"device": map[string]any{
					"model": "MacBook",
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := gen.ParseSDL(schema)
	require.NoError(t, err)

	userDocs := mustAddDocIDsToDocs(ctx, []map[string]any{
		{
			"name":      "John",
			"_deviceID": mustGetDocIDFromDocMap(ctx, map[string]any{"model": "iPhone"}, colDefMap["Device"]),
		},
		{
			"name":      "Fred",
			"_deviceID": mustGetDocIDFromDocMap(ctx, map[string]any{"model": "MacBook"}, colDefMap["Device"]),
		},
	}, colDefMap["User"])
	deviceDocs := mustAddDocIDsToDocs(ctx, []map[string]any{
		{"model": "iPhone"},
		{"model": "MacBook"},
	}, colDefMap["Device"])

	errorMsg := assertDocs(append(userDocs, deviceDocs...), docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_OneToOneToOnePrimary(t *testing.T) {
	ctx := context.Background()
	schema := `
		type User {
			name: String
			device: Device @primary
		}
		type Device {
			model: String
			owner: User
			specs: Specs @primary
		}
		type Specs {
			OS: String
			device: Device
		}`

	docs, err := AddFromSDL(ctx, schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
					"specs": map[string]any{
						"OS": "iOS",
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := gen.ParseSDL(schema)
	require.NoError(t, err)

	specsDoc := mustAddDocIDToDoc(ctx, map[string]any{"OS": "iOS"}, colDefMap["Specs"])
	deviceDoc := mustAddDocIDToDoc(ctx, map[string]any{
		"model":    "iPhone",
		"_specsID": specsDoc[request.DocIDFieldName],
	}, colDefMap["Device"])
	userDoc := mustAddDocIDToDoc(ctx, map[string]any{
		"name":      "John",
		"_deviceID": deviceDoc[request.DocIDFieldName],
	}, colDefMap["User"])

	errorMsg := assertDocs([]map[string]any{userDoc, deviceDoc, specsDoc}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_OneToTwoPrimary(t *testing.T) {
	ctx := context.Background()
	schema := `
		type User {
			name: String
			device: Device @primary
		}
		type Device {
			model: String
			owner: User
			specs: Specs
		}
		type Specs {
			OS: String
			device: Device @primary
		}`

	docs, err := AddFromSDL(ctx, schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
					"specs": map[string]any{
						"OS": "iOS",
					},
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := gen.ParseSDL(schema)
	require.NoError(t, err)

	deviceDoc := mustAddDocIDToDoc(ctx, map[string]any{"model": "iPhone"}, colDefMap["Device"])
	specsDoc := mustAddDocIDToDoc(ctx, map[string]any{
		"OS":        "iOS",
		"_deviceID": deviceDoc[request.DocIDFieldName],
	}, colDefMap["Specs"])
	userDoc := mustAddDocIDToDoc(ctx, map[string]any{
		"name":      "John",
		"_deviceID": deviceDoc[request.DocIDFieldName],
	}, colDefMap["User"])

	errorMsg := assertDocs([]map[string]any{userDoc, deviceDoc, specsDoc}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

func TestGeneratePredefinedFromSchema_TwoPrimaryToOneRoot(t *testing.T) {
	ctx := context.Background()
	schema := `
		type User {
			name: String
			device: Device @primary
			address: Address @primary
		}
		type Device {
			model: String
			owner: User
		}
		type Address {
			street: String
			user: User 
		}`

	docs, err := AddFromSDL(ctx, schema, DocsList{
		ColName: "User",
		Docs: []map[string]any{
			{
				"name": "John",
				"device": map[string]any{
					"model": "iPhone",
				},
				"address": map[string]any{
					"street": "Backer",
				},
			},
		},
	})
	assert.NoError(t, err)

	colDefMap, err := gen.ParseSDL(schema)
	require.NoError(t, err)

	deviceDoc := mustAddDocIDToDoc(ctx, map[string]any{"model": "iPhone"}, colDefMap["Device"])
	addressDoc := mustAddDocIDToDoc(ctx, map[string]any{"street": "Backer"}, colDefMap["Address"])
	userDoc := mustAddDocIDToDoc(ctx, map[string]any{
		"name":       "John",
		"_deviceID":  deviceDoc[request.DocIDFieldName],
		"_addressID": addressDoc[request.DocIDFieldName],
	}, colDefMap["User"])

	errorMsg := assertDocs([]map[string]any{userDoc, deviceDoc, addressDoc}, docs)
	if errorMsg != "" {
		t.Error(errorMsg)
	}
}

// func TestGeneratePredefinedFromSchema_OneToMany(t *testing.T) {
// 	schema := `
// 		type User {
// 			name: String
// 			devices: [Device]
// 		}
// 		type Device {
// 			model: String
// 			owner: User
// 		}`

// 	docs, err := AddFromSDL(schema, DocsList{
// 		ColName: "User",
// 		Docs: []map[string]any{
// 			{
// 				"name": "John",
// 				"devices": []map[string]any{
// 					{"model": "iPhone"},
// 					{"model": "PlayStation"},
// 				},
// 			},
// 			{
// 				"name": "Fred",
// 				"devices": []map[string]any{
// 					{"model": "Surface"},
// 					{"model": "Pixel"},
// 				},
// 			},
// 		},
// 	})
// 	assert.NoError(t, err)

// 	colDefMap, err := parseSDL(schema)
// 	require.NoError(t, err)

// 	johnDocID := mustGetDocIDFromDocMap(map[string]any{"name": "John"}, colDefMap["User"].Schema)
// 	fredDocID := mustGetDocIDFromDocMap(map[string]any{"name": "Fred"}, colDefMap["User"].Schema)
// 	errorMsg := assertDocs(mustAddDocIDsToDocs([]map[string]any{
// 		{"name": "John"},
// 		{"name": "Fred"},
// 		{"model": "iPhone", "_ownerID": johnDocID},
// 		{"model": "PlayStation", "_ownerID": johnDocID},
// 		{"model": "Surface", "_ownerID": fredDocID},
// 		{"model": "Pixel", "_ownerID": fredDocID},
// 	}, col), docs)
// 	if errorMsg != "" {
// 		t.Error(errorMsg)
// 	}
// }

// func TestGeneratePredefinedFromSchema_OneToManyToOne(t *testing.T) {
// 	schema := `
// 		type User {
// 			name: String
// 			devices: [Device]
// 		}
// 		type Device {
// 			model: String
// 			owner: User
// 			specs: Specs
// 		}
// 		type Specs {
// 			CPU: String
// 			device: Device @primary
// 		}`

// 	docs, err := AddFromSDL(schema, DocsList{
// 		ColName: "User",
// 		Docs: []map[string]any{
// 			{
// 				"name": "John",
// 				"devices": []map[string]any{
// 					{
// 						"model": "iPhone",
// 						"specs": map[string]any{
// 							"CPU": "A13",
// 						},
// 					},
// 					{
// 						"model": "MacBook",
// 						"specs": map[string]any{
// 							"CPU": "M2",
// 						},
// 					},
// 				},
// 			},
// 		},
// 	})
// 	assert.NoError(t, err)

// 	colDefMap, err := parseSDL(schema)
// 	require.NoError(t, err)

// 	johnDocID := mustGetDocIDFromDocMap(map[string]any{"name": "John"}, colDefMap["User"].Schema)
// 	errorMsg := assertDocs(mustAddDocIDsToDocs([]map[string]any{
// 		{"name": "John"},
// 		{"model": "iPhone", "_ownerID": johnDocID},
// 		{"model": "MacBook", "_ownerID": johnDocID},
// 		{
// 			"CPU": "A13",
// 			"_deviceID": mustGetDocIDFromDocMap(map[string]any{
// 				"model":    "iPhone",
// 				"_ownerID": johnDocID,
// 			}, colDefMap["Device"].Schema),
// 		},
// 		{
// 			"CPU": "M2",
// 			"_deviceID": mustGetDocIDFromDocMap(map[string]any{
// 				"model":    "MacBook",
// 				"_ownerID": johnDocID,
// 			}, colDefMap["Device"].Schema),
// 		},
// 	}), docs)
// 	if errorMsg != "" {
// 		t.Error(errorMsg)
// 	}
// }

// func TestGeneratePredefined_OneToMany(t *testing.T) {
// 	defs := []client.CollectionDefinition{
// 		{
// 			Description: client.CollectionVersion{
// 				Name: "User",
// 				ID:   0,
// 			},
// 			Schema: client.SchemaDescription{
// 				Name: "User",
// 				Fields: []client.SchemaFieldDescription{
// 					{
// 						Name: "name",
// 						Kind: client.FieldKind_STRING,
// 					},
// 					{
// 						Name:         "devices",
// 						Kind:         client.FieldKind_FOREIGN_OBJECT_ARRAY,
// 						Schema:       "Device",
// 					},
// 				},
// 			},
// 		},
// 		{
// 			Description: client.CollectionVersion{
// 				Name: "Device",
// 				ID:   1,
// 			},
// 			Schema: client.SchemaDescription{
// 				Name: "Device",
// 				Fields: []client.SchemaFieldDescription{
// 					{
// 						Name: "model",
// 						Kind: client.FieldKind_STRING,
// 					},
// 					{
// 						Name:   "owner",
// 						Kind:   client.FieldKind_FOREIGN_OBJECT,
// 						Schema: "User",
// 						IsPrimary: true,
// 					},
// 				},
// 			},
// 		},
// 	}
// 	docs, err := Add(defs, DocsList{
// 		ColName: "User",
// 		Docs: []map[string]any{
// 			{
// 				"name": "John",
// 				"devices": []map[string]any{
// 					{"model": "iPhone"},
// 					{"model": "PlayStation"},
// 				},
// 			},
// 			{
// 				"name": "Fred",
// 				"devices": []map[string]any{
// 					{"model": "Surface"},
// 					{"model": "Pixel"},
// 				},
// 			},
// 		},
// 	})
// 	assert.NoError(t, err)

// 	johnDocID := mustGetDocIDFromDocMap(map[string]any{"name": "John"}, defs[0].Schema)
// 	fredDocID := mustGetDocIDFromDocMap(map[string]any{"name": "Fred"}, defs[0].Schema)
// 	errorMsg := assertDocs(mustAddDocIDsToDocs([]map[string]any{
// 		{"name": "John"},
// 		{"name": "Fred"},
// 		{"model": "iPhone", "_ownerID": johnDocID},
// 		{"model": "PlayStation", "_ownerID": johnDocID},
// 		{"model": "Surface", "_ownerID": fredDocID},
// 		{"model": "Pixel", "_ownerID": fredDocID},
// 	}), docs)
// 	if errorMsg != "" {
// 		t.Error(errorMsg)
// 	}
// }
