Feature: Lambda Expression Support
  In order to use functional programming features like map, filter, and reduce
  As a developer
  I want to define and use lambda (anonymous) functions

  Scenario: Simple single-parameter lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> x + 1
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda output includes representation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> x + 1
      """
    When I run the application with this content
    Then the output should contain "lambda:"
    And the output should contain "x + 1"

  Scenario: Lambda with two parameters
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (a, b) -> a + b
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with string concatenation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (name) -> "Hello " + name
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with comparison expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> x > 10
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with if/else expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> if (x > 0) "positive" else "non-positive"
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with array access
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (arr) -> arr[0]
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with multiplication
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> x * 2
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with division
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> x / 2
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with nested arithmetic
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> (x + 1) * 2
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with logical AND
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x, y) -> x > 0 and y > 0
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with equality check
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> x == 42
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with less-than comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (age) -> age < 18
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with less-than-or-equal comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (score) -> score <= 100
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with greater-than-or-equal comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (value) -> value >= 0
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with field access using dot notation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (obj) -> obj.name
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with nested field access
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (obj) -> obj.person.name
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with sizeOf function call
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (arr) -> sizeOf(arr)
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with complex expression using sizeOf
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (arr) -> sizeOf(arr) > 0
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with default operator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> x default 0
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Three-parameter lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x, y, z) -> x + y + z
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda body with parenthesized subexpression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> (x + 5) * (x - 2)
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with nested if/else
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x) -> if (x > 10) "high" else if (x > 0) "low" else "zero"
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with object literal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (name, age) -> {name: name, age: age}
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with array literal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (a, b) -> [a, b]
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda with underscore in parameter name
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (item_value) -> item_value + 1
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Lambda body accessing variable from context
    Given the following input content:
      """
      %im 0.1
      var multiplier = 3
      output application/json
      ---
      (x) -> x * multiplier
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Multiline map body with brace on separate line
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] map
          {
              value: $
          }
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"value":1},{"value":2},{"value":3}]
      """

  Scenario: Multiline map body with DataWeave header
    Given the following input content:
      """
      %dw 2.0
      output application/json
      ---
      [{"name": "alice"}, {"name": "bob"}] map
          {
              greeting: "hello " ++ $.name
          }
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"greeting":"hello alice"},{"greeting":"hello bob"}]
      """

  Scenario: Nested lambdas keep inner and outer parameters distinct
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] map (x, i) -> ([x, i] map (y, j) -> y + j + x)
      """
    When I run the application with this content
    Then the output should be:
      """
      [[2,2],[4,4]]
      """
