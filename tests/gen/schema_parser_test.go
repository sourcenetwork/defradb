package gen

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSchemaParser_Parse(t *testing.T) {
	tests := []struct {
		name   string
		schema string
		want   map[string]typeDefinition
	}{
		{
			name: "basic types",
			schema: `
				type User {
					name: String
					age: Int
				}`,
			want: map[string]typeDefinition{
				"User": {
					name:  "User",
					index: 0,
					props: []propDefinition{
						{name: "name", typeStr: "String"},
						{name: "age", typeStr: "Int"},
					},
				},
			},
		},
		{
			name: "array and relations",
			schema: `
				type User {
					name: String
					devices: [Device]
				}
				type Device {
					model: String
					owner: User
				}`,
			want: map[string]typeDefinition{
				"User": {
					name:  "User",
					index: 0,
					props: []propDefinition{
						{name: "name", typeStr: "String"},
						{name: "devices", typeStr: "Device", isArray: true, isRelation: true},
					},
				},
				"Device": {
					name:  "Device",
					index: 1,
					props: []propDefinition{
						{name: "model", typeStr: "String"},
						{name: "owner", typeStr: "User", isRelation: true, isPrimary: true},
					},
				},
			},
		},
		{
			name: "primary annotation",
			schema: `
				type User {
					name: String
					device: Device @primary
				}
				type Device {
					model: String
					owner: User
				}`,
			want: map[string]typeDefinition{
				"User": {
					name:  "User",
					index: 0,
					props: []propDefinition{
						{name: "name", typeStr: "String"},
						{name: "device", typeStr: "Device", isRelation: true, isPrimary: true},
					},
				},
				"Device": {
					name:  "Device",
					index: 1,
					props: []propDefinition{
						{name: "model", typeStr: "String"},
						{name: "owner", typeStr: "User", isRelation: true},
					},
				},
			},
		},
		{
			name: "make first encountered type primary",
			schema: `
				type T1 {
					secondary: T2 
				}
				type T2 {
					primary: T1
				}
				type T3 {
					secondary: T4 
				}
				type T4 {
					primary: T3
                    secondary: T5
				}
				type T5 {
					primary: T4
				}`,
			want: map[string]typeDefinition{
				"T1": {
					name:  "T1",
					index: 0,
					props: []propDefinition{
						{name: "secondary", typeStr: "T2", isRelation: true},
					},
				},
				"T2": {
					name:  "T2",
					index: 1,
					props: []propDefinition{
						{name: "primary", typeStr: "T1", isRelation: true, isPrimary: true},
					},
				},
				"T3": {
					name:  "T3",
					index: 2,
					props: []propDefinition{
						{name: "secondary", typeStr: "T4", isRelation: true},
					},
				},
				"T4": {
					name:  "T4",
					index: 3,
					props: []propDefinition{
						{name: "primary", typeStr: "T3", isRelation: true, isPrimary: true},
						{name: "secondary", typeStr: "T5", isRelation: true},
					},
				},
				"T5": {
					name:  "T5",
					index: 4,
					props: []propDefinition{
						{name: "primary", typeStr: "T4", isRelation: true, isPrimary: true},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &schemaParser{}
			got := p.Parse(tt.schema)
			assert.Equal(t, tt.want, got)
		})
	}
}
