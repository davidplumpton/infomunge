# DataWeave Cookbook Examples - InfoMunge Equivalents

This document shows how to accomplish the DataWeave cookbook examples using InfoMunge.

## Important Syntax Notes

Before reviewing the examples, be aware of these key InfoMunge syntax requirements:

### 1. **Parentheses Required for Complex Expressions in Object Literals**
```im
// ❌ WRONG
{ label: value1 ++ value2 }

// ✅ CORRECT
{ label: (value1 ++ value2) }
```

### 2. **Array Field Selection and Recursive Descent**
Like DataWeave, InfoMunge supports direct field selection on arrays:
```im
items.name
items["name"]
```

These selectors collect the field from immediate object elements, skip elements
without the field, and return `null` when no element matches. Use recursive
descent when matching objects may be nested:

```im
items..name
```

### 3. **mapObject Parameter Order is Reversed**
```im
// DataWeave: mapObject (value, key)
// InfoMunge: mapObject (key, value)
{ a: 1, b: 2 } mapObject (key, val) -> [upper(key), val]
```

### 4. **Parentheses Around Multi-line Operations**
When chaining operations, wrap in parentheses if needed:
```im
// ❌ May fail
payload map (x) -> x.value reduce (v, acc) -> acc + v

// ✅ Preferred
(payload map (x) -> x.value) reduce (v, acc) -> acc + v
```

### 5. **Input Headers Are Compatibility Metadata**
InfoMunge accepts `input application/json` and `input payload application/json`
header lines so DataWeave examples remain readable, but those lines do not parse,
reparse, rename, validate, or create inputs. The CLI `-i` flags, server `/run`
inputs, or embedding API context choose input names and formats before the runner
evaluates the script.

## 1. Extract Data - Simple Field Access

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload.name
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload.name
```

**Input JSON:**
```json
{"name": "somebody"}
```

**Output:**
```json
"somebody"
```

---

## 2. Extract Data - Multiple Fields

### DataWeave Example
```dataweave
%dw 2.0
var myObject = { "myKey" : "1234", "name" : "somebody" }
var myArray = [ { "myKey" : "1234" }, { "name" : "somebody" } ]
output application/json
---
{
    selectingValueUsingKeyInObject : myObject.name,
    selectingValueUsingKeyOfObjectInArray : myArray.name,
}
```

### InfoMunge Equivalent

```im
%im 0.1
var myObject = { "myKey": "1234", "name": "somebody" }
var myArray = [ { "myKey": "1234" }, { "name": "somebody" } ]
output application/json
---
{
  selectingValueUsingKeyInObject: myObject.name,
  selectingValueUsingKeyOfObjectInArray: myArray.name
}
```

**Output:**
```json
{
  "selectingValueUsingKeyInObject": "somebody",
  "selectingValueUsingKeyOfObjectInArray": ["somebody"]
}
```

---

## 3. Transform XML to JSON (Basic Transformation)

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
{
    address1: payload.order.buyer.address,
    city: payload.order.buyer.city,
    country: payload.order.buyer.nationality,
    email: payload.order.buyer.email,
    name: payload.order.buyer.name,
    postalCode: payload.order.buyer.postCode,
    stateOrProvince: payload.order.buyer.state
}
```

### InfoMunge Equivalent
```im
%im 0.1
input application/xml
output application/json
---
{
  address1: payload.order.buyer.address,
  city: payload.order.buyer.city,
  country: payload.order.buyer.nationality,
  email: payload.order.buyer.email,
  name: payload.order.buyer.name,
  postalCode: payload.order.buyer.postCode,
  stateOrProvince: payload.order.buyer.state
}
```

**Input XML:**
```xml
<?xml version='1.0' encoding='UTF-8'?>
<order>
  <buyer>
    <email>mike@hotmail.com</email>
    <name>Michael</name>
    <address>Koala Boulevard 314</address>
    <city>San Diego</city>
    <state>CA</state>
    <postCode>1345</postCode>
    <nationality>USA</nationality>
  </buyer>
</order>
```

**Output:**
```json
{
  "address1": "Koala Boulevard 314",
  "city": "San Diego",
  "country": "USA",
  "email": "mike@hotmail.com",
  "name": "Michael",
  "postalCode": "1345",
  "stateOrProvince": "CA"
}
```

---

## 4. Map Array - Transform Each Item

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
items: payload.books map (item, index) -> {
      book: item mapObject (value, key) -> {
      (upper(key)): value
      }
}
```

### InfoMunge Equivalent

**Note:** Nested mapObject within object literals has parsing limitations. Transform all keys to uppercase as a separate step:

```im
%im 0.1
input application/json
output application/json
---
payload.books map (item) -> (item mapObject (key, value) -> [upper(key), value])
```

**Alternative approach (simpler):**
```im
%im 0.1
input application/json
output application/json
---
{
  items: payload.books map (item, index) -> {
    book: item
  }
}
```

**Input JSON:**
```json
{
  "books": [
    {
      "title": "Everyday Italian",
      "author": "Giada De Laurentiis",
      "year": "2005",
      "price": "30.00"
    },
    {
      "title": "Harry Potter",
      "author": "J K. Rowling",
      "year": "2005",
      "price": "29.99"
    }
  ]
}
```

**Output (with mapObject approach):**
```json
[
  {
    "TITLE": "Everyday Italian",
    "AUTHOR": "Giada De Laurentiis",
    "YEAR": "2005",
    "PRICE": "30.00"
  },
  {
    "TITLE": "Harry Potter",
    "AUTHOR": "J K. Rowling",
    "YEAR": "2005",
    "PRICE": "29.99"
  }
]
```

---

## 5. Map and Flatten Array

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
[[1, 2], [3, 4], [5]] flatMap (x) -> x
```

### InfoMunge Equivalent
```im
%im 0.1
output application/json
---
[[1, 2], [3, 4], [5]] flatMap (x) -> x
```

**Output:**
```json
[1,2,3,4,5]
```

---

## 6. Map with Type Coercion

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
items: (payload.books map {
      category: "book",
      price: $.price as Number,
      id: $$,
      properties: {
        title: $.title,
        author: $.author,
        year: $.year as Number
      }
})
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
{
  items: payload.books map (item, index) -> {
    category: "book",
    price: item.price as Number,
    id: index,
    properties: {
      title: item.title,
      author: item.author,
      year: item.year as Number
    }
  }
}
```

**Note:** In InfoMunge, type coercion uses the `as` operator, and array index is accessed via the second parameter in lambda functions.

---

## 7. Filter Array

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload.users.*name[?($ == "Mariano")]
```

### InfoMunge Equivalent
```im
%im 0.1
output application/json
---
["Mariano", "Luis", "Mariano"] filter (name) -> name == "Mariano"
```

**Or with extracted names from objects:**
```im
%im 0.1
input application/json
output application/json
---
(payload map (item) -> item.name) filter (name) -> name == "Mariano"
```

**Input JSON:**
```json
[
  {"name": "Mariano"},
  {"name": "Luis"},
  {"name": "Mariano"}
]
```

**Output:**
```json
["Mariano", "Mariano"]
```

---

## 8. Group By

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload groupBy (x) -> x.status
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload groupBy (item) -> item.status
```

**Input JSON:**
```json
[
  {"name": "Alice", "status": "active"},
  {"name": "Bob", "status": "inactive"},
  {"name": "Charlie", "status": "active"}
]
```

**Output:**
```json
{
  "active": [
    {"name": "Alice", "status": "active"},
    {"name": "Charlie", "status": "active"}
  ],
  "inactive": [
    {"name": "Bob", "status": "inactive"}
  ]
}
```

---

## 9. Pluck (Extract Field from Array of Objects)

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload pluck "name"
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload pluck "name"
```

**Input JSON:**
```json
[
  {"name": "Alice", "age": 30},
  {"name": "Bob", "age": 25},
  {"name": "Charlie", "age": 35}
]
```

**Output:**
```json
["Alice","Bob","Charlie"]
```

---

## 10. Distinct (Remove Duplicates)

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload distinctBy (x) -> x.type
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload distinctBy (item) -> item.type
```

**Input JSON:**
```json
[
  {"id": 1, "type": "A"},
  {"id": 2, "type": "A"},
  {"id": 3, "type": "B"},
  {"id": 4, "type": "B"}
]
```

**Output:**
```json
[
  {"id": 1, "type": "A"},
  {"id": 3, "type": "B"}
]
```

---

## 11. Map Object Keys and Values

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload mapObject (value, key) -> {
  (upper(key)): value
}
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload mapObject (key, value) -> [upper(key), value]
```

**Input JSON:**
```json
{
  "name": "Alice",
  "age": 30,
  "email": "alice@example.com"
}
```

**Output:**
```json
{
  "NAME": "Alice",
  "AGE": 30,
  "EMAIL": "alice@example.com"
}
```

---

## 12. Filter Object

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload mapObject (value, key) -> {
  (key) : value
} filter (value) -> value > 2
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload filterObject (key, value) -> value > 2
```

**Input JSON:**
```json
{
  "a": 1,
  "b": 2,
  "c": 3,
  "d": 4
}
```

**Output:**
```json
{
  "c": 3,
  "d": 4
}
```

---

## 13. Extract Object Keys

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
keysOf(payload)
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
keysOf(payload)
```

**Input JSON:**
```json
{
  "name": "Alice",
  "age": 30
}
```

**Output:**
```json
["age", "name"]
```

---

## 14. Extract Object Values

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
valuesOf(payload)
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
valuesOf(payload)
```

**Input JSON:**
```json
{
  "name": "Alice",
  "age": 30
}
```

**Output:**
```json
["Alice", 30]
```

---

## 15. Object Entries (Key-Value Pairs)

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
entriesOf(payload)
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
entriesOf(payload)
```

**Input JSON:**
```json
{
  "name": "Alice",
  "age": 30
}
```

**Output:**
```json
[
  {"key": "age", "value": 30},
  {"key": "name", "value": "Alice"}
]
```

---

## 16. Conditional Logic with If-Else

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload map (x) -> if (x > 2) x * 10 else x
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload map (item) -> if (item > 2) item * 10 else item
```

**Input JSON:**
```json
[1, 2, 3, 4, 5]
```

**Output:**
```json
[1, 2, 30, 40, 50]
```

---

## 17. Reduce (Custom Aggregation)

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload reduce (item, acc = 0) -> acc + item.value
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload reduce (item, acc = 0) -> acc + item.value
```

**Input JSON:**
```json
[
  {"value": 10},
  {"value": 20},
  {"value": 30}
]
```

**Output:**
```json
60
```

---

## 18. Multiple Inputs

### DataWeave Example
```dataweave
%dw 2.0
input payload application/json
input users application/json
output application/json
---
{
  orders: payload,
  users: users
}
```

### InfoMunge Equivalent
```im
%im 0.1
input payload application/json
input users application/json
output application/json
---
{
  orders: payload,
  users: users
}
```

**Usage:**
The `payload` and `users` variables below are created by the `-i` flags. The
matching `input` header lines are accepted as compatibility documentation only.

```bash
./infomunge -i payload=orders.json -i users=users.json "%im 0.1
input payload application/json
input users application/json
output application/json
---
{
  orders: payload,
  users: users
}"
```

---

## 19. String Functions - Concatenation

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
payload map (item) -> item.firstName ++ " " ++ item.lastName
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
payload map (item) -> (item.firstName ++ " " ++ item.lastName)
```

**Note:** String concatenation in lambda bodies doesn't require parentheses, but in object literals it does: `{label: (val1 ++ val2)}`

**Input JSON:**
```json
[
  {"firstName": "John", "lastName": "Doe"},
  {"firstName": "Jane", "lastName": "Smith"}
]
```

**Output:**
```json
["John Doe", "Jane Smith"]
```

---

## 20. Type Coercion

### DataWeave Example
```dataweave
%dw 2.0
output application/json
---
{
  stringNum: payload.price as String,
  numberStr: payload.quantity as Number,
  boolVal: payload.active as Boolean
}
```

### InfoMunge Equivalent
```im
%im 0.1
input application/json
output application/json
---
{
  stringNum: payload.price as String,
  numberStr: payload.quantity as Number,
  boolVal: payload.active as Boolean
}
```

**Input JSON:**
```json
{
  "price": 100,
  "quantity": "50",
  "active": "true"
}
```

**Output:**
```json
{
  "stringNum": "100",
  "numberStr": 50,
  "boolVal": true
}
```

---

## Summary: Key Differences Between DataWeave and InfoMunge

| Feature | DataWeave | InfoMunge |
|---------|-----------|-----------|
| **Header** | `%dw 2.0` | `%im 0.1` |
| **Array Field Selection** | `.name` on array | Must use `..name` (recursive descent) or `map` |
| **Object Iteration** | `mapObject (value, key)` | `mapObject (key, value)` |
| **Object Iteration Result** | Single value (key is in parens) | Array `[key, value]` |
| **Default Parameters** | `$` (value), `$$` (index) | Named parameters required |
| **String Concat in Objects** | `{key: val1 ++ val2}` | Must use `{key: (val1 ++ val2)}` (requires parens) |
| **Complex Expressions in Objects** | Can use operators directly | Require parentheses: `{key: (expr1 op expr2)}` |
| **Multiple Inputs** | `input <name> <format>` | CLI/server input names create variables; `input <name> <format>` headers are compatibility metadata |
| **Null Literal** | `null` | `null` (preferred), `nil` alias also accepted |
| **Type System** | Complex type definitions | Basic types with `as` coercion |
| **Supported Formats** | JSON, XML, CSV, YAML, Properties | JSON, XML, CSV, YAML; Properties input only |

---

## Running These Examples

To test these examples, use:

```bash
# Simple expression
./infomunge -i payload=input.json "%im 0.1
input application/json
output application/json
---
payload.name"

# From file
./infomunge -i payload=input.json -f transformation.im

# Multiple inputs
./infomunge -i payload=orders.json -i users=users.json -f multi_input.im
```
