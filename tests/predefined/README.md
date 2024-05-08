# Creation of Predefined Documents

`Create` and `CreateFromSDL` can be used to generate predefined documents.

They accepts the predefined list of documents `DocList` that in turn might include nested documents.

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
      "address": map[string]any{
        "city": "Munich",
      },
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