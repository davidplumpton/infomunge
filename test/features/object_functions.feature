Feature: Object Functions
  In order to manipulate and transform objects
  As a developer
  I want to use the object functions: entriesOf, objectToArray, arrayToObject, keysOf, valuesOf, namesOf, mapObject, filterObject

  # entriesOf tests
  Scenario: entriesOf basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      entriesOf({"name": "Alice", "age": 30})
      """
    When I run the application with this content
    Then the output should contain "\"key\":\"age\""
    And the output should contain "\"value\":30"
    And the output should contain "\"key\":\"name\""
    And the output should contain "\"value\":\"Alice\""

  Scenario: entriesOf with numeric keys
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      entriesOf({"id": 1, "count": 5, "active": true})
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 3

  Scenario: entriesOf with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      entriesOf({})
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: entriesOf on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      entriesOf([1, 2, 3])
      """
    When I run the application and it fails
    Then the output should contain "entriesOf: argument 1 expected object"

  Scenario: entriesOf with nested objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      entriesOf({"person": {"name": "Bob"}, "status": "active"})
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2

  # objectToArray tests
  Scenario: objectToArray basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      objectToArray({"a": 1, "b": 2})
      """
    When I run the application with this content
    Then the output should contain "\"key\":\"a\""
    And the output should contain "\"value\":1"
    And the output should contain "\"key\":\"b\""
    And the output should contain "\"value\":2"

  Scenario: objectToArray on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      objectToArray([1, 2, 3])
      """
    When I run the application and it fails
    Then the output should contain "objectToArray: argument 1 expected object"

  # arrayToObject tests
  Scenario: arrayToObject basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      arrayToObject([{key: "a", value: 1}, {key: "b", value: 2}])
      """
    When I run the application with this content
    Then the output should contain "\"a\":1"
    And the output should contain "\"b\":2"

  Scenario: arrayToObject on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      arrayToObject({"a": 1})
      """
    When I run the application and it fails
    Then the output should contain "arrayToObject: argument 1 expected array"

  Scenario: arrayToObject with missing key fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      arrayToObject([{value: 1}])
      """
    When I run the application and it fails
    Then the output should contain "arrayToObject: missing key at index 0"

  # keysOf tests
  Scenario: keysOf basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keysOf({"name": "Alice", "age": 30})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["age","name"]
      """

  Scenario: keysOf with single key
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keysOf({"id": 123})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["id"]
      """

  Scenario: keysOf with many keys
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keysOf({"z": 1, "a": 2, "m": 3, "b": 4})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","b","m","z"]
      """

  Scenario: keysOf with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keysOf({})
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: keysOf on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keysOf("not an object")
      """
    When I run the application and it fails
    Then the output should contain "keysOf: argument 1 expected object"

  # valuesOf tests
  Scenario: valuesOf basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      valuesOf({"name": "Alice", "age": 30})
      """
    When I run the application with this content
    Then the output should contain "30"
    And the output should contain "\"Alice\""

  Scenario: valuesOf maintains key order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      valuesOf({"z": "last", "a": "first", "m": "middle"})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["first","middle","last"]
      """

  Scenario: valuesOf with mixed types
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      valuesOf({"str": "text", "num": 42, "bool": true, "null": null})
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 4

  Scenario: valuesOf with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      valuesOf({})
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: valuesOf on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      valuesOf(123)
      """
    When I run the application and it fails
    Then the output should contain "valuesOf: argument 1 expected object"

  # namesOf tests (same as keysOf)
  Scenario: namesOf basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      namesOf({"first": "John", "last": "Doe"})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["first","last"]
      """

  Scenario: namesOf with numeric string keys
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      namesOf({"2": "b", "1": "a", "3": "c"})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["1","2","3"]
      """

  Scenario: namesOf on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      namesOf(null)
      """
    When I run the application and it fails
    Then the output should contain "keysOf: argument 1 expected object"

  # mapObject tests
  Scenario: mapObject with key transformation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"a": 1, "b": 2}, (k, v) -> [k ++ "_new", v * 2])
      """
    When I run the application with this content
    Then the output should contain "\"a_new\":2"
    And the output should contain "\"b_new\":4"

  Scenario: mapObject with value transformation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"x": 10, "y": 20}, (k, v) -> [k, v + 5])
      """
    When I run the application with this content
    Then the output should contain "\"x\":15"
    And the output should contain "\"y\":25"

  Scenario: mapObject with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({}, (k, v) -> [k, v])
      """
    When I run the application with this content
    Then the output should be:
      """
      {}
      """

  Scenario: mapObject ignores key parameter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"id": 1}, (k, v) -> ["constant", v])
      """
    When I run the application with this content
    Then the output should contain "\"constant\":1"

  Scenario: mapObject requires lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"a": 1}, "not a lambda")
      """
    When I run the application and it fails
    Then the output should contain "mapObject expects a lambda function"

  Scenario: mapObject on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject([1, 2, 3], (k, v) -> [k, v])
      """
    When I run the application and it fails
    Then the output should contain "mapObject expects an object"

  Scenario: mapObject with DataWeave parameter order (value, key)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"a": 1, "b": 2}, (value, key) -> [key ++ "_mapped", value * 10])
      """
    When I run the application with this content
    Then the output should contain "\"a_mapped\":10"
    And the output should contain "\"b_mapped\":20"

  Scenario: mapObject with DataWeave short names (v, k)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"x": 5, "y": 10}, (v, k) -> [k ++ "_val", v + 100])
      """
    When I run the application with this content
    Then the output should contain "\"x_val\":105"
    And the output should contain "\"y_val\":110"

  Scenario: mapObject with arbitrary parameter names uses DataWeave order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"a": 1, "b": 2}, (first, second) -> [second ++ "_dw", first + 1])
      """
    When I run the application with this content
    Then the output should contain "\"a_dw\":2"
    And the output should contain "\"b_dw\":3"

  Scenario: mapObject with key-value parameter names keeps legacy order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"a": 1, "b": 2}, (k, v) -> [k, v + 10])
      """
    When I run the application with this content
    Then the output should contain "\"a\":11"
    And the output should contain "\"b\":12"

  Scenario: mapObject returns object entries from conditional object literal
    Given the following JSON input:
      """
      {
        "flights": [
          {
            "availableSeats": 45,
            "airlineName": "Ryan Air",
            "aircraftBrand": "Boeing",
            "aircraftType": "737",
            "departureDate": "12/14/2017",
            "origin": "BCN",
            "destination": "FCO"
          },
          {
            "availableSeats": 15,
            "airlineName": "Ryan Air",
            "aircraftBrand": "Boeing",
            "aircraftType": "747",
            "departureDate": "08/03/2017",
            "origin": "FCO",
            "destination": "DFW"
          }
        ]
      }
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.flights map (flight) -> {
        (flight mapObject (value, key) -> {
          (emptySeats: value) if (key == "availableSeats"),
          (airline: value) if (key == "airlineName"),
          ((key): value) if (key != "availableSeats" and key != "airlineName")
        })
      }
      """
    When I run the script
    Then the output should contain "\"emptySeats\":45"
    And the output should contain "\"airline\":\"Ryan Air\""
    And the output should contain "\"destination\":\"FCO\""
    And the output should contain "\"emptySeats\":15"
    And the output should contain "\"destination\":\"DFW\""

  Scenario: mapObject conditional keys with key as String
    Given the following JSON input:
      """
      {
        "flights": [
          {
            "availableSeats": 45,
            "airlineName": "Ryan Air",
            "aircraftBrand": "Boeing",
            "aircraftType": "737",
            "departureDate": "12/14/2017",
            "origin": "BCN",
            "destination": "FCO"
          },
          {
            "availableSeats": 15,
            "airlineName": "Ryan Air",
            "aircraftBrand": "Boeing",
            "aircraftType": "747",
            "departureDate": "08/03/2017",
            "origin": "FCO",
            "destination": "DFW"
          }
        ]
      }
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.flights map (flight) -> {
        (flight mapObject (value, key) -> {
          (emptySeats: value) if (key as String == "availableSeats"),
          (airline: value) if (key as String == "airlineName"),
          ((key): value) if (key as String != "availableSeats" and key as String != "airlineName")
        })
      }
      """
    When I run the script
    Then the output should contain "\"emptySeats\":45"
    And the output should contain "\"airline\":\"Ryan Air\""
    And the output should contain "\"destination\":\"FCO\""
    And the output should contain "\"emptySeats\":15"
    And the output should contain "\"destination\":\"DFW\""

  Scenario: top-level object literal without braces
    Given the following JSON input:
      """
      {
        "books": [
          {
            "title": "Dune",
            "author": "Frank Herbert"
          }
        ]
      }
      """
    And the following script:
      """
      %dw 2.0
      output application/json
      ---
      items: payload.books map (item, index) -> {
        book: item mapObject (value, key) -> {
          (upper(key)): value
        }
      }
      """
    When I run the script
    Then the output should contain "\"items\""
    And the output should contain "\"TITLE\":\"Dune\""
    And the output should contain "\"AUTHOR\":\"Frank Herbert\""

  # filterObject tests
  Scenario: filterObject basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2, "c": 3}, (k, v) -> v > 1)
      """
    When I run the application with this content
    Then the output should contain "\"b\":2"
    And the output should contain "\"c\":3"
    And the output should not contain "\"a\":1"

  Scenario: filterObject by key name
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"active": true, "archived": false, "admin": true}, (k, v) -> v)
      """
    When I run the application with this content
    Then the output should contain "\"active\":true"
    And the output should contain "\"admin\":true"
    And the output should not contain "\"archived\""

  Scenario: filterObject with arbitrary parameter names uses DataWeave order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2, "c": 3}, (left, right) -> left > 1)
      """
    When I run the application with this content
    Then the output should contain "\"b\":2"
    And the output should contain "\"c\":3"
    And the output should not contain "\"a\":1"

  Scenario: filterObject with arbitrary parameter names and index uses DataWeave order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2, "c": 3}, (left, right, idx) -> right == "b" and left == 2 and idx == 1)
      """
    When I run the application with this content
    Then the output should be:
      """
      {"b":2}
      """

  Scenario: filterObject with no matches
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2}, (k, v) -> v > 10)
      """
    When I run the application with this content
    Then the output should be:
      """
      {}
      """

  Scenario: filterObject with all matches
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"x": 5, "y": 10}, (k, v) -> v > 0)
      """
    When I run the application with this content
    Then the output should contain "\"x\":5"
    And the output should contain "\"y\":10"

  Scenario: filterObject with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({}, (k, v) -> true)
      """
    When I run the application with this content
    Then the output should be:
      """
      {}
      """

  Scenario: filterObject using key parameter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"first": 1, "second": 2, "third": 3}, (k, v) -> k != "third")
      """
    When I run the application with this content
    Then the output should contain "\"first\":1"
    And the output should contain "\"second\":2"
    And the output should not contain "\"third\""

  Scenario: filterObject requires lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1}, "not a lambda")
      """
    When I run the application and it fails
    Then the output should contain "filterObject expects a lambda function"

  Scenario: filterObject on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject("not an object", (k, v) -> true)
      """
    When I run the application and it fails
    Then the output should contain "filterObject expects an object"

  # Combination scenarios
  Scenario: entriesOf result can be filtered
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      entriesOf({"a": 1, "b": 2, "c": 3}) filter (entry) -> entry.value > 1
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2

  Scenario: keysOf in mapObject
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keysOf(mapObject({"x": 10, "y": 20}, (k, v) -> [k, v * 2]))
      """
    When I run the application with this content
    Then the output should be:
      """
      ["x","y"]
      """

  Scenario: valuesOf after filterObject
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      valuesOf(filterObject({"a": 1, "b": 2, "c": 3}, (k, v) -> v != 2))
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,3]
      """

  Scenario: nested object transformation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mapObject({"user": {"name": "John"}, "status": "active"}, (k, v) -> if (k == "user") [k, {"name": (v.name ++ " Doe")}] else [k, v])
      """
    When I run the application with this content
    Then the output should contain "\"user\""

  Scenario: multiple operations on object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2, "c": 3}, (k, v) -> v > 1) mapObject (k, v) -> [k, v * 10]
      """
    When I run the application with this content
    Then the output should contain "\"b\":20"
    And the output should contain "\"c\":30"
    And the output should not contain "\"a\""

  Scenario: filterObject operator chained with mapObject operator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2, "c": 3} filterObject (k, v) -> v > 1 mapObject (k, v) -> [k, v * 10]
      """
    When I run the application with this content
    Then the output should contain "\"b\":20"
    And the output should contain "\"c\":30"
    And the output should not contain "\"a\""

  Scenario: mapObject operator chained with filterObject operator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2, "c": 3} mapObject (k, v) -> [k, v * 2] filterObject (k, v) -> v > 2
      """
    When I run the application with this content
    Then the output should contain "\"b\":4"
    And the output should contain "\"c\":6"
    And the output should not contain "\"a\""
