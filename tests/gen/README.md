# Automatic Documents Generation 

`AutoGenerateFromSchema` is a highly versatile function designed for dynamic document generation, perfect for testing and simulation purposes. 
It creates documents based on a specified schema, allowing for extensive customization of data generation. 

The function generates documents adhering to a defined schema and it's configuration.
It interprets the types and relationships within the schema to create realistic, interconnected data structures.

### Demand Calculation:

The function calculates the 'demand' or the number of documents to generate based on the configuration provided.
For related types within the schema, it intelligently adjusts the number of generated documents to maintain consistency in relationships (one-to-one, one-to-many, etc.).

In the absence of explicit demands, it deduces demands from the maximum required by related types or uses a default value if no relation-based demands are present.

The error will be returned if the demand for documents can not be satisfied. 
For example, a document expects at least 10 secondary documents, but the demand for secondary documents is 5.

## Configuration

There are two ways to configure the function:
1. Directly within the schema using annotations
2. Via options passed to the function

Options take precedence over in-schema configurations.

### In-schema Configuration:

Field values can be configured directly within the schema using annotations after "#" (e.g., `# min: 1, max: 120` for an integer field).

At the moment, the following value configurations are supported:
- `min` and `max` for integer, float and relation fields. For relation fields, the values define the minimum and maximum number of related documents.
- `len` for string fields

Default value ranges are used when not explicitly set in the schema or via options.

### Customization with Options:

- `WithTypeDemand` and `WithTypeDemandRange` allow setting the specific number (or range) of documents for a given type.
- `WithFieldRange` and `WithFieldLen` override in-schema configurations for field ranges and lengths.
- `WithFieldGenerator` provides custom value generation logic for specific fields.
- `WithRandomSeed` ensures deterministic output, useful for repeatable tests.

## Examples

### Basic Document Generation:

```go
schema := `
type User {
  name: String # len: 10
  age: Int # min: 18, max: 50
  verified: Boolean
  rating: Float # min: 0.0, max: 5.0
}`
docs, _ := AutoGenerateFromSchema(schema, WithTypeDemand("User", 100))
```

### Custom Field Range:

Overrides the age range specified in the schema.

```go
docs, _ := AutoGenerateFromSchema(schema, WithTypeDemand("User", 50), WithFieldRange("User", "age", 25, 30))
```

### One-to-Many Relationship:

Generates User documents each related to multiple Device documents.

```go
schema := `
type User { 
  name: String 
  devices: [Device] # min: 1, max: 3
}
type Device {
  model: String
  owner: User
}`
docs, _ := AutoGenerateFromSchema(schema, WithTypeDemand("User", 10))
```

### Custom Value Generation:

Custom generation for age field.

```go
nameWithPrefix := func(i int, next func() any) any {
  return "user_" + next().(string)
}
docs, _ := AutoGenerateFromSchema(schema, WithTypeDemand("User", 10), WithFieldGenerator("User", "name", nameWithPrefix))
```

## Conclusion

`AutoGenerateFromSchema` is a powerful tool for generating structured, relational data on the fly. Its flexibility in configuration and intelligent demand calculation makes it ideal for testing complex data models and scenarios.

# Generation of Predefined Documents

`GeneratePredefinedFromSchema` can be used to generate predefined documents.

It accepts the predefined list of documents `DocList` that in turn might include nested documents.

The fields in `DocList` might be a superset of the fields in the schema. 
In that case, only the fields in the schema will be considered.


For example, for the following schema:
```graphql
type User {
  name: String 
  devices: [Device] 
} 

type Device {
  model: String 
  owner: User
} 
```
if the `DocList` is as follows:
```go
gen.DocsList{
  ColName: "User",
  Docs: []map[string]any{
    {
      "name":     "Shahzad",
      "age":      20,
      "verified": false,
      "email":    "shahzad@gmail.com",
      "devices": []map[string]any{
        {
          "model": "iPhone Xs",
          "year":  2022,
          "type":  "phone",
        }},
    }},
}
```
only the following doc will be considered:
```go
gen.DocsList{
  ColName: "User",
  Docs: []map[string]any{
    {
      "name":     "Shahzad",
      "devices": []map[string]any{
        {
          "model": "iPhone Xs",
        }},
    }},
}
```
This allows having a predefined large list of documents (and sub-documents) and only use a subset of field for a particular test case.