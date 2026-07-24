Feature: Pluck Function
  In order to transform objects and extract array fields
  As a developer
  I want to use the pluck function on both objects and arrays

  Scenario: pluck on object with (value, key) lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2} pluck (v, k) -> k
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","b"]
      """

  Scenario: pluck on object with (value, key, index) lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"x": 10, "y": 20} pluck (v, k, i) -> k ++ (i as String)
      """
    When I run the application with this content
    Then the output should be:
      """
      ["x0","y1"]
      """

  Scenario: pluck on object with DataWeave parameter order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2} pluck (value, key) -> value + 10
      """
    When I run the application with this content
    Then the output should be:
      """
      [11,12]
      """

  Scenario: pluck on object with arbitrary parameter names uses DataWeave order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2} pluck (first, second) -> second ++ ":" ++ (first as String)
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a:1","b:2"]
      """

  Scenario: pluck on object with arbitrary parameter names and index uses DataWeave order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2} pluck (left, right, idx) -> right ++ ":" ++ (idx as String) ++ ":" ++ (left as String)
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a:0:1","b:1:2"]
      """

  Scenario: pluck on array with string selector (existing field extraction)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}] pluck "name"
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice","Bob"]
      """

  Scenario: pluck on array with nested string selector
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"user": {"id": 1}}, {"user": {"id": 2}}] pluck "user.id"
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2]
      """

  Scenario: pluck on array with lambda acts like map
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] pluck (x) -> x * 10
      """
    When I run the application with this content
    Then the output should be:
      """
      [10,20,30]
      """

  Scenario: pluck on empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {} pluck (v, k) -> k
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: pluck on empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] pluck "name"
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: pluck requires two arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      pluck([1, 2, 3])
      """
    When I run the application and it fails
    Then the output should contain "pluck requires exactly 2 arguments"

  Scenario: pluck invalid source type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      123 pluck "name"
      """
    When I run the application and it fails
    Then the output should contain "pluck expects an array or object"

  Scenario: pluck on object requires a lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1} pluck "a"
      """
    When I run the application and it fails
    Then the output should contain "pluck on object expects a lambda function"
