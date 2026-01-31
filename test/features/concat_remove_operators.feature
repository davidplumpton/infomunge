Feature: Concatenation and Remove Operators
  In order to combine and remove elements
  As a developer
  I want to use the ++ (concatenate) and -- (remove) operators

  # ++ Operator (concatenation) - Strings
  Scenario: concatenate two strings with ++
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "Hello" ++ " World"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello World"
      """

  Scenario: concatenate multiple strings with ++
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "a" ++ "b" ++ "c"
      """
    When I run the application with this content
    Then the output should be:
      """
      "abc"
      """

  Scenario: concatenate string and number with ++
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" ++ 123
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello123"
      """

  Scenario: concatenate empty strings
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "" ++ "hello"
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: concatenate strings with special characters
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "Hello\n" ++ "World\t!"
      """
    When I run the application with this content
    Then the output should contain "Hello"
    And the output should contain "World"

  Scenario: concatenate single character strings
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "a" ++ "b" ++ "c" ++ "d"
      """
    When I run the application with this content
    Then the output should be:
      """
      "abcd"
      """

  # ++ Operator (concatenation) - Arrays
  Scenario: concatenate two arrays with ++
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] ++ [3, 4]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3,4]
      """

  Scenario: concatenate multiple arrays with ++
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1] ++ [2] ++ [3]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: concatenate empty arrays
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] ++ [1, 2]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2]
      """

  Scenario: concatenate arrays with different types
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, "two"] ++ [true, null]
      """
    When I run the application with this content
    Then the output should contain "1"
    And the output should contain "two"
    And the output should contain "true"
    And the output should contain "null"

  Scenario: concatenate nested arrays
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[1, 2]] ++ [[3, 4]]
      """
    When I run the application with this content
    Then the output should be:
      """
      [[1,2],[3,4]]
      """

  Scenario: concatenate array and object should fail
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] ++ {"a": 3}
      """
    When I run the application and it fails
    Then the error should contain "cannot concatenate mixed types"

  # ++ Operator (concatenation) - Objects
  Scenario: concatenate two objects with ++
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1} ++ {"b": 2}
      """
    When I run the application with this content
    Then the output should contain "a"
    And the output should contain "1"
    And the output should contain "b"
    And the output should contain "2"

  Scenario: concatenate multiple objects with ++
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1} ++ {"b": 2} ++ {"c": 3}
      """
    When I run the application with this content
    Then the output should contain "a"
    And the output should contain "b"
    And the output should contain "c"

  Scenario: concatenate objects with overlapping keys
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2} ++ {"b": 3, "c": 4}
      """
    When I run the application with this content
    Then the output should contain "a"
    And the output should contain "c"
    And the output should contain "3"

  Scenario: concatenate empty objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {} ++ {"x": 10}
      """
    When I run the application with this content
    Then the output should contain "x"
    And the output should contain "10"

  Scenario: concatenate objects with nested values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"user": {"name": "Alice"}} ++ {"user": {"age": 30}}
      """
    When I run the application with this content
    Then the output should contain "age"
    And the output should contain "30"

  # -- Operator (remove) - Basic usage
  Scenario: remove single element from array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 2] -- 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,3]
      """

  Scenario: remove element not in array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] -- 4
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: remove from empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] -- 1
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: remove all occurrences
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 1, 1] -- 1
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: remove string from array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["a", "b", "a", "c"] -- "a"
      """
    When I run the application with this content
    Then the output should be:
      """
      ["b","c"]
      """

  Scenario: remove null from array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, null, 2, null] -- null
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2]
      """

  Scenario: remove boolean from array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [true, false, true] -- true
      """
    When I run the application with this content
    Then the output should be:
      """
      [false]
      """

  # -- Operator (remove) - Object key removal by array
  Scenario: remove multiple keys from object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2, c: 3} -- ["b", "c"]
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":1}
      """

  Scenario: remove single key from object using array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {x: 10, y: 20, z: 30} -- ["y"]
      """
    When I run the application with this content
    Then the output should contain "x"
    And the output should contain "z"
    And the output should not contain "y"

  Scenario: remove keys not present in object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2} -- ["x", "y"]
      """
    When I run the application with this content
    Then the output should contain "a"
    And the output should contain "b"

  Scenario: remove all keys from object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2} -- ["a", "b"]
      """
    When I run the application with this content
    Then the output should be:
      """
      {}
      """

  Scenario: remove keys with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2} -- []
      """
    When I run the application with this content
    Then the output should contain "a"
    And the output should contain "b"

  # Error cases
  Scenario: remove from non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" -- "l"
      """
    When I run the application and it fails
    Then the output should contain "expects first argument to be an array or object"

  Scenario: concatenate string and array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" ++ [1, 2]
      """
    When I run the application and it fails
    Then the output should contain "cannot concatenate mixed types"

  Scenario: concatenate array and object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] ++ {"a": 1}
      """
    When I run the application and it fails
    Then the output should contain "cannot concatenate mixed types"

  # Combined scenarios
  Scenario: filter then concatenate arrays
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ([1, 2, 3] filter (x) -> x > 1) ++ [4, 5]
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,3,4,5]
      """

  Scenario: concatenate and remove
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ([1, 2] ++ [3, 2]) -- 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,3]
      """

  Scenario: build string with concatenation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "Result: " ++ "3"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Result: 3"
      """

  Scenario: deduplicate array using remove
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 1, 3, 2] -- 1
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,3,2]
      """

  Scenario: merge multiple objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"x": 1} ++ {"y": 2} ++ {"z": 3}
      """
    When I run the application with this content
    Then the output should contain "x"
    And the output should contain "y"
    And the output should contain "z"

  Scenario: flatten nested arrays using reduce
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      flatten([[1, 2], [3, 4], [5]])
      """
    When I run the application with this content
    Then the output should contain "1"
    And the output should contain "5"
