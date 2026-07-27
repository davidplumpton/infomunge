Feature: Default Operator
  In order to handle null/missing values gracefully
  As a developer
  I want to use the default operator for null coalescing

  Scenario: Default with nil value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      nil default "anonymous"
      """
    When I run the application with this content
    Then the output should be:
      """
      "anonymous"
      """

  Scenario: Default with non-nil value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "alice" default "anonymous"
      """
    When I run the application with this content
    Then the output should be:
      """
      "alice"
      """

  Scenario: Default with numeric values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      nil default 42
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Default preserves non-nil numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      0 default 100
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: Default with chained operations
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      nil default 5 + 10
      """
    When I run the application with this content
    Then the output should be:
      """
      15
      """

  Scenario: Default with context variables
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload default "empty"
      """
    When I run the application with this JSON input:
      """
      null
      """
    Then the output should be:
      """
      "empty"
      """

  Scenario: Default with existing context variables
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload default "empty"
      """
    When I run the application with this JSON input:
      """
      "data"
      """
    Then the output should be:
      """
      "data"
      """

  Scenario: Multiple default operators
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      nil default nil default "final"
      """
    When I run the application with this content
    Then the output should be:
      """
      "final"
      """

  Scenario: Default inside object literal field
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {value: nil default "fallback"}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"value":"fallback"}
      """

  Scenario: Default to an array feeds a collection pipeline
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (null default [1, 2]) map $ * 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,4]
      """

  Scenario: An ungrouped collection operator consumes the completed default expression
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.name default flatten([[0]]) reduce (item, x) -> x
      """
    When I run the application with this JSON input:
      """
      {"name":[]}
      """
    Then the output should be:
      """
      null
      """

  Scenario: Grouping keeps a collection operator inside the default fallback
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null default (flatten([[0]]) reduce (item, x) -> x)
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: Default to an array inside a lambda body
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [null] map ($ default [[1], [2, 3]])
      """
    When I run the application with this content
    Then the output should be:
      """
      [[[1],[2,3]]]
      """

  Scenario: Ungrouped default consumes a complete additive expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      var payload = {name: -635}
      ---
      1 + payload.label default 5
      """
    When I run the application and it fails
    Then the error should contain "cannot add int and <nil>"

  Scenario: Grouping keeps default local to an additive operand
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      1 + (payload.label default 5)
      """
    When I run the application with this JSON input:
      """
      {"name":-635}
      """
    Then the output should be:
      """
      6
      """

  Scenario: Ungrouped default consumes a complete unary expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      var payload = {name: -635}
      ---
      [-(payload.label) default 5, !(payload.label) default true]
      """
    When I run the application and it fails
    Then the error should contain "cannot negate <nil>"

  Scenario: Default remains inside an explicit collection lambda body
    Given the following input content:
      """
      %im 0.1
      output application/json
      var payload = {name: -635}
      ---
      [1, 2] map (x) -> x + payload.label default 10
      """
    When I run the application and it fails
    Then the error should contain "cannot add int and <nil>"

  Scenario: Default handles composed expressions from missing input fields
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      fun plusOne(value) = value + 1
      ---
      [
        payload.label default 5,
        payload.user.name default "missing",
        -(payload.label) default 6,
        payload.label + 1 default 7,
        abs(payload.label) default 8,
        plusOne(payload.label) default 9,
        [1, 2] map (x) -> x + payload.label default 10
      ]
      """
    When I run the application with this JSON input:
      """
      {"name":-635}
      """
    Then the output should be:
      """
      [5,"missing",6,7,8,9,[10,10]]
      """

  Scenario: Default does not hide explicit null arithmetic failures
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null + 1 default 5
      """
    When I run the application and it fails
    Then the error should contain "cannot add <nil> and int"

  Scenario: Input absence does not hide unrelated division failures
    Given the following JSON input:
      """
      {"name":-635}
      """
    And the following script:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.label + (1 / 0) default 5
      """
    When running the script should fail with error containing "division by zero"
