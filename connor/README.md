# Connor
**Flexible condition DSL for Go**

Connor is a simple condition DSL and evaluator for Go inspired by MongoDB's
query language. It aims to provide a simple and straightforward means to
express conditions against `map[string]interface{}` objects for everything
from rules engines to query tools.

Connor only implements a subset of MongoDB's query language, the most commonly
used methods, however it has been designed to make adding new operators simple
and straightforward should you require them.

## Example

```go
package main

import (
    "github.com/SierraSoftworks/connor"
)

func parse(d string) map[string]interface{} {
    var v map[string]interface{}
    if err := json.NewDecoder(bytes.NewBufferString(d)).Decode(&v); err != nil {
        fmt.Fatal(err)
    }

    return v
}

func main() {
    conds := parse(`{
        "x": 1,
        "y": { "$in": [1, 2, 3] },
        "z": { "$ne": 5 }
    }`)

    data := parse(`{
        "x": 1,
        "y": 2,
        "z": 3
    }`)

    if match, err := connor.Match(conds, data); err != nil {
        fmt.Fatal("failed to run match:", err)
    } else if match {
        fmt.Println("Matched")
    } else {
        fmt.Println("No Match")
    }
}
```

## Operators
Connor has a number of built in operators which enable you to quickly compare a number
of common data structures to one another. The following are supported operators for use
in your conditions.

### Equality `$eq`
```json
{ "$eq": "value" }
```

### Inequality `$ne`
```json
{ "$ne": "value" }
```

### Greater Than `$gt`
```json
{ "$gt": 5.3 }
```

### Greater Than or Equal `$ge`
```json
{ "$ge": 5 }
```

### Less Than `$lt`
```json
{ "$lt": 42 }
```

### Less Than or Equal `$le`
```json
{ "$le": 42 }
```

### Set Contains `$in`
```json
{ "$in": [1, 2, 3] }
```

### Set Excludes `$nin`
```json
{ "$nin": [1, 2, 3] }
```

### String Contains `$contains`
```json
{ "$contains": "test" }
```

### Logical And `$and`
```json
{ "$and": [{ "$gt": 5 }, { "$lt": 10 }]}
```

### Logical Or `$or`
```json
{ "$or": [{ "$gt": 10 }, { "$eq": 0 }]}
```

## Custom Operators
Connor supports registering your own custom operators for any additional condition
evaluation you wish to perform. These operators are registered using the `Register()`
method and are expected to match the following interface:

```go
type Operator interface {
    Name() string
    Evaluate(condition, data interface{}) (bool, error)
}
```

The following is an example of an operator which determines whether the data is nil.

```go
func init() {
    connor.Register(&NilOperator{})
}

type NilOperator struct {}

func (o *NilOperator) Name() string {
    return "nil"
}

func (o *NilOperator) Evaluate(condition, data interface{}) (bool, error) {
    if c, ok := condition.(bool); ok {
        return data != nil ^ c, nil
    } else {
        return data == nil, nil
    }
}
```

You can then use this operator as the following example shows, or specify `false`
to check for non-nil values.

```json
{
    "x": { "$nil": true }
}
```
