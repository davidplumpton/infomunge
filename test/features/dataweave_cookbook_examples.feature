Feature: DataWeave Cookbook Examples Converted to InfoMunge
  In order to verify InfoMunge can handle DataWeave cookbook examples
  As a developer
  I want to test common transformation patterns from MuleSoft DataWeave

  # Example 1: Simple Field Access
  Scenario: Extract single field from object
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.name
      """
    When I run the application with this JSON input:
      """
      {"name": "somebody"}
      """
    Then the output should be:
      """
      "somebody"
      """

  # Example 2: Multiple Field Extraction
  Scenario: Extract multiple fields with renaming
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      {
        address1: payload.address,
        city: payload.city,
        country: payload.nationality,
        email: payload.email,
        name: payload.name,
        postalCode: payload.postCode,
        stateOrProvince: payload.state
      }
      """
    When I run the application with this JSON input:
      """
      {
        "email": "mike@hotmail.com",
        "name": "Michael",
        "address": "Koala Boulevard 314",
        "city": "San Diego",
        "state": "CA",
        "postCode": "1345",
        "nationality": "USA"
      }
      """
    Then the output should be:
      """
      {"address1":"Koala Boulevard 314","city":"San Diego","country":"USA","email":"mike@hotmail.com","name":"Michael","postalCode":"1345","stateOrProvince":"CA"}
      """

  # Example 3: Map Array Items
  Scenario: Map array to transform each item
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.books map (item, index) -> {
        title: item.title,
        author: item.author,
        id: index
      }
      """
    When I run the application with this JSON input:
      """
      {
        "books": [
          {
            "title": "Everyday Italian",
            "author": "Giada De Laurentiis",
            "year": "2005"
          },
          {
            "title": "Harry Potter",
            "author": "J K. Rowling",
            "year": "2005"
          }
        ]
      }
      """
    Then the output should be:
      """
      [{"author":"Giada De Laurentiis","id":0,"title":"Everyday Italian"},{"author":"J K. Rowling","id":1,"title":"Harry Potter"}]
      """

  # Example 4: FlatMap (Map and Flatten)
  Scenario: FlatMap to flatten nested arrays
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[1, 2], [3, 4], [5]] flatMap (x) -> x
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3,4,5]
      """

  # Example 5: Type Coercion with Map
  Scenario: Map with type coercion using as operator
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload map (item) -> {
        title: item.title,
        price: item.price as Number,
        year: item.year as Number
      }
      """
    When I run the application with this JSON input:
      """
      [
        {
          "title": "Book One",
          "price": "30.00",
          "year": "2005"
        },
        {
          "title": "Book Two",
          "price": "29.99",
          "year": "2006"
        }
      ]
      """
    Then the output should be:
      """
      [{"price":30,"title":"Book One","year":2005},{"price":29.99,"title":"Book Two","year":2006}]
      """

  # Example 6: Filter Array
  Scenario: Filter array with condition
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload filter (item) -> item > 2
      """
    When I run the application with this JSON input:
      """
      [1, 2, 3, 4, 5]
      """
    Then the output should be:
      """
      [3,4,5]
      """

  # Example 7: GroupBy
  Scenario: Group array items by field value
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload groupBy (item) -> item.status
      """
    When I run the application with this JSON input:
      """
      [
        {"name": "Alice", "status": "active"},
        {"name": "Bob", "status": "inactive"},
        {"name": "Charlie", "status": "active"}
      ]
      """
    Then the output should be:
      """
      {"active":[{"name":"Alice","status":"active"},{"name":"Charlie","status":"active"}],"inactive":[{"name":"Bob","status":"inactive"}]}
      """

  # Example 8: Pluck (Extract Field)
  Scenario: Pluck extracts single field from array of objects
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload pluck "name"
      """
    When I run the application with this JSON input:
      """
      [
        {"name": "Alice", "age": 30},
        {"name": "Bob", "age": 25},
        {"name": "Charlie", "age": 35}
      ]
      """
    Then the output should be:
      """
      ["Alice","Bob","Charlie"]
      """

  # Example 9: Distinct By
  Scenario: DistinctBy removes duplicate values based on key expression
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload distinctBy (item) -> item.type
      """
    When I run the application with this JSON input:
      """
      [
        {"id": 1, "type": "A"},
        {"id": 2, "type": "A"},
        {"id": 3, "type": "B"},
        {"id": 4, "type": "B"}
      ]
      """
    Then the output should be valid JSON with array length of 2

  # Example 10: MapObject to transform keys
  Scenario: MapObject transforms object keys to uppercase
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload mapObject (key, value) -> [upper(key), value]
      """
    When I run the application with this JSON input:
      """
      {
        "name": "Alice",
        "age": 30,
        "email": "alice@example.com"
      }
      """
    Then the output should be:
      """
      {"AGE":30,"EMAIL":"alice@example.com","NAME":"Alice"}
      """

  # Example 11: FilterObject
  Scenario: FilterObject filters out entries not matching predicate
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload filterObject (key, value) -> value > 2
      """
    When I run the application with this JSON input:
      """
      {"a": 1, "b": 2, "c": 3, "d": 4}
      """
    Then the output should be:
      """
      {"c":3,"d":4}
      """

  # Example 12: KeysOf
  Scenario: KeysOf extracts all keys from object
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      keysOf(payload)
      """
    When I run the application with this JSON input:
      """
      {"name": "Alice", "age": 30, "email": "alice@example.com"}
      """
    Then the output should be:
      """
      ["age","email","name"]
      """

  # Example 13: ValuesOf
  Scenario: ValuesOf extracts all values from object
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      valuesOf(payload)
      """
    When I run the application with this JSON input:
      """
      {"name": "Alice", "age": 30}
      """
    Then the output should contain "\"Alice\""
    And the output should contain "30"

  # Example 14: EntriesOf
  Scenario: EntriesOf extracts key-value pairs as array of objects
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      entriesOf(payload)
      """
    When I run the application with this JSON input:
      """
      {"name": "Alice", "age": 30}
      """
    Then the output should contain "\"key\":\"age\""
    And the output should contain "\"value\":30"
    And the output should contain "\"key\":\"name\""
    And the output should contain "\"value\":\"Alice\""

  # Example 15: Conditional If-Else
  Scenario: Map with conditional logic
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload map (item) -> if (item > 2) item * 10 else item
      """
    When I run the application with this JSON input:
      """
      [1, 2, 3, 4, 5]
      """
    Then the output should be:
      """
      [1,2,30,40,50]
      """

  # Example 16: Nested Object Access
  Scenario: Access deeply nested fields
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      {
        name: payload.user.profile.name,
        city: payload.user.address.city
      }
      """
    When I run the application with this JSON input:
      """
      {
        "user": {
          "profile": {
            "name": "Alice",
            "age": 30
          },
          "address": {
            "street": "Main St",
            "city": "New York"
          }
        }
      }
      """
    Then the output should be:
      """
      {"city":"New York","name":"Alice"}
      """

  # Example 17: String Concatenation
  Scenario: Concatenate strings with ++ operator
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload map (item) -> item.firstName ++ " " ++ item.lastName
      """
    When I run the application with this JSON input:
      """
      [
        {"firstName": "John", "lastName": "Doe"},
        {"firstName": "Jane", "lastName": "Smith"}
      ]
      """
    Then the output should be:
      """
      ["John Doe","Jane Smith"]
      """

  # Example 18: Reduce for aggregation
  Scenario: Reduce sums all values
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      (payload map (item) -> item.value) reduce (acc, val) -> acc + val
      """
    When I run the application with this JSON input:
      """
      [
        {"value": 10},
        {"value": 20},
        {"value": 30}
      ]
      """
    Then the output should be:
      """
      60
      """

  # Example 19: Map with object construction
  Scenario: Map items into new object structure
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload map (item) -> {
        id: item.id,
        label: (item.name ++ " (" ++ (item.age as String) ++ ")")
      }
      """
    When I run the application with this JSON input:
      """
      [
        {"id": 1, "name": "Alice", "age": 30},
        {"id": 2, "name": "Bob", "age": 25}
      ]
      """
    Then the output should be:
      """
      [{"id":1,"label":"Alice (30)"},{"id":2,"label":"Bob (25)"}]
      """

  # Example 20: FlatMap with filtering
  Scenario: FlatMap with conditional logic
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload flatMap (item) -> if (item > 2) [item, item * 2] else [item]
      """
    When I run the application with this JSON input:
      """
      [1, 2, 3, 4]
      """
    Then the output should be:
      """
      [1,2,3,6,4,8]
      """
