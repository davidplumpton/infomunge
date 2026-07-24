Feature: Filter Operator
  In order to select elements from arrays based on a condition
  As a developer
  I want to use the filter operator with lambda expressions

  Scenario: Filter simple array by condition
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] filter (x) -> x > 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,4,5]
      """

  Scenario: Filter array of objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}, {"name": "Charlie", "age": 35}] filter (person) -> person.age > 26
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2
    And the output should contain "Alice"
    And the output should contain "Charlie"
    And the output should not contain "Bob"

  Scenario: Filter with equality check
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["apple", "banana", "apple", "cherry"] filter (item) -> item == "apple"
      """
    When I run the application with this content
    Then the output should be:
      """
      ["apple","apple"]
      """

  Scenario: Filter all elements
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] filter (x) -> x > 0
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3,4,5]
      """

  Scenario: Filter no elements
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] filter (x) -> x > 10
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: Filter with nested field access
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"user": {"status": "active"}}, {"user": {"status": "inactive"}}, {"user": {"status": "active"}}] filter (item) -> item.user.status == "active"
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"user":{"status":"active"}},{"user":{"status":"active"}}]
      """

  Scenario: Filter with sizeOf function
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[], [1], [1, 2], [1, 2, 3]] filter (arr) -> sizeOf(arr) > 1
      """
    When I run the application with this content
    Then the output should be:
      """
      [[1,2],[1,2,3]]
      """

  Scenario: Filter with multiple conditions using AND
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}, {"name": "Charlie", "age": 35}] filter (person) -> person.age > 26 and person.age < 34
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 1
    And the output should contain "Alice"
    And the output should not contain "Bob"
    And the output should not contain "Charlie"

  Scenario: Filter with if/else in lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] filter (x) -> if (x > 3) true else false
      """
    When I run the application with this content
    Then the output should be:
      """
      [4,5]
      """

  Scenario: Filter empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] filter (x) -> x > 0
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: Filter non-array raises error
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "not an array" filter (x) -> x > 0
      """
    When I run the application and it fails
    Then the output should contain "filter expects an array"

  Scenario: Filter on object suggests filterObject
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1} filter (value, key) -> value > 0
      """
    When I run the application and it fails
    Then the output should contain "Did you mean to use 'filterObject'?"
