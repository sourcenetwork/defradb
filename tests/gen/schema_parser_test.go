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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &schemaParser{}
			got := p.Parse(tt.schema)
			assert.Equal(t, tt.want, got)
		})
	}
}
