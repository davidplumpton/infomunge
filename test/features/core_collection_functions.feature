Feature: Core Collection Functions
  In order to work with collections and objects
  As a developer
  I want to use the core collection functions: isEmpty, flatten, unique, reverse, keys, values

  # isEmpty tests
  Scenario: isEmpty with empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty("")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: isEmpty with non-empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty("hello")
      """
    When I run the application with this content
    Then the output should be false

  Scenario: isEmpty with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty([])
      """
    When I run the application with this content
    Then the output should be true

  Scenario: isEmpty with non-empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty([1, 2, 3])
      """
    When I run the application with this content
    Then the output should be false

  Scenario: isEmpty with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty({})
      """
    When I run the application with this content
    Then the output should be true

  Scenario: isEmpty with non-empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty({"key": "value"})
      """
    When I run the application with this content
    Then the output should be false

  Scenario: isEmpty with null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty(null)
      """
    When I run the application with this content
    Then the output should be true

  Scenario: isEmpty with number returns false
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty(0)
      """
    When I run the application with this content
    Then the output should be false

  Scenario: isEmpty with boolean returns false
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEmpty(false)
      """
    When I run the application with this content
    Then the output should be false

  # flatten tests
  Scenario: flatten removes exactly one array nesting level
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(flatten([[1, [2, 3]], [4]]))
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  # unique tests
  Scenario: unique basic usage with numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique([1, 2, 2, 3, 3, 3])
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: unique with strings
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique(["a", "b", "a", "c", "b"])
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","b","c"]
      """

  Scenario: unique preserves order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique([3, 1, 2, 1, 3, 2])
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,1,2]
      """

  Scenario: unique with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique([])
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: unique with single element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique([42])
      """
    When I run the application with this content
    Then the output should be:
      """
      [42]
      """

  Scenario: unique with all duplicates
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique([5, 5, 5, 5])
      """
    When I run the application with this content
    Then the output should be:
      """
      [5]
      """

  Scenario: unique distinguishes nested values by language type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique([{a: 1}, {a: "1"}, {a: 1}])
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"a":1},{"a":"1"}]
      """

  Scenario: unique on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique("not an array")
      """
    When I run the application and it fails
    Then the output should contain "unique: argument 1 expected array, got string"

  Scenario: unique distinguishes different types with same string representation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      unique([1, "1", true, "true", null, 2, "1", 1])
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,"1",true,"true",null,2]
      """

  # reverse tests
  Scenario: reverse basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      reverse([1, 2, 3, 4, 5])
      """
    When I run the application with this content
    Then the output should be:
      """
      [5,4,3,2,1]
      """

  Scenario: reverse with strings
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      reverse(["a", "b", "c"])
      """
    When I run the application with this content
    Then the output should be:
      """
      ["c","b","a"]
      """

  Scenario: reverse with mixed types
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      reverse([1, "two", 3, "four"])
      """
    When I run the application with this content
    Then the output should be:
      """
      ["four",3,"two",1]
      """

  Scenario: reverse with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      reverse([])
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: reverse with single element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      reverse([42])
      """
    When I run the application with this content
    Then the output should be:
      """
      [42]
      """

  Scenario: reverse with nested arrays
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      reverse([[1, 2], [3, 4], [5, 6]])
      """
    When I run the application with this content
    Then the output should be:
      """
      [[5,6],[3,4],[1,2]]
      """

  Scenario: reverse on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      reverse("not an array")
      """
    When I run the application and it fails
    Then the output should contain "reverse: argument 1 expected array, got string"

  # keys tests
  Scenario: keys basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keys({"name": "Alice", "age": 30})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["name","age"]
      """

  Scenario: keys with single key
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keys({"id": 123})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["id"]
      """

  Scenario: keys returns keys in insertion order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keys({"z": 1, "a": 2, "m": 3, "b": 4})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["z","a","m","b"]
      """

  Scenario: keys with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keys({})
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: keys propagates literal and missing-selector null values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [keys(null), keys({}["active"])]
      """
    When I run the application with this content
    Then the output should be:
      """
      [null,null]
      """

  Scenario: keys on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keys("not an object")
      """
    When I run the application and it fails
    Then the output should contain "keys: argument 1 expected object"

  Scenario: keys on array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      keys([1, 2, 3])
      """
    When I run the application and it fails
    Then the output should contain "keys: argument 1 expected object"

  # values tests
  Scenario: values basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      values({"a": 1, "b": 2, "c": 3})
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: values returns values in insertion order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      values({"z": "last", "a": "first", "m": "middle"})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["last","first","middle"]
      """

  Scenario: values with mixed types
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      values({"str": "text", "num": 42, "bool": true, "null": null})
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 4

  Scenario: values with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      values({})
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: values on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      values(123)
      """
    When I run the application and it fails
    Then the output should contain "values: argument 1 expected object"

  Scenario: values on array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      values([1, 2, 3])
      """
    When I run the application and it fails
    Then the output should contain "values: argument 1 expected object"

  # Combination scenarios
  Scenario: isEmpty combined with filter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[], [1], [], [2, 3], []] filter (x) -> !isEmpty(x)
      """
    When I run the application with this content
    Then the output should be:
      """
      [[1],[2,3]]
      """

  Scenario: unique and reverse combined
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      reverse(unique([1, 2, 2, 3, 1]))
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,2,1]
      """

  Scenario: keys and values have matching order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "keys": keys({"b": 2, "a": 1, "c": 3}),
        "values": values({"b": 2, "a": 1, "c": 3})
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"keys":["b","a","c"],"values":[2,1,3]}
      """
