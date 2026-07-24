Feature: Error Reporting
  In order to debug InfoMunge scripts
  As a developer
  I want to see accurate line and column numbers in error messages

  Scenario: Syntax error in a simple expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      1 + / 2
      """
    When I run the application and it fails
    Then the error should contain "4:5:"

  Scenario: Syntax error in an object with unquoted keys
    Given the following input content:
      """
      %im 0.1
      ---
      {
        name: "Alice",
        age: 30 + (5 * )
      }
      """
    When I run the application and it fails
    Then the error should contain "5:18:"

  Scenario: Syntax error in a header variable
    Given the following input content:
      """
      %im 0.1
      var x = 1 + (
      ---
      x
      """
    When I run the application and it fails
    Then the error should contain "2:13:"

  Scenario: Header directive error includes line text
    Given the following input content:
      """
      %im 0.1
      bogus directive
      ---
      1
      """
    When I run the application and it fails
    Then the error should contain "2:1:"
    And the error should contain "line: bogus directive"

  Scenario: Malformed header variable reports explicit declaration error
    Given the following input content:
      """
      %im 0.1
      var x
      ---
      1
      """
    When I run the application and it fails
    Then the error should contain "2:1:"
    And the error should contain "invalid variable declaration: missing '='"

  Scenario: Runtime error - division by zero
    Given the following input content:
      """
      %im 0.1
      ---
      10 / (5 - 5)
      """
    When I run the application and it fails
    Then the error should contain "3:1:"

  Scenario: Builtin runtime errors report their source position
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
        sqrt(-1)
      """
    When I run the application and it fails
    Then the error should contain "4:3: sqrt: cannot take square root of negative number -1"

  Scenario: Syntax error after a line comment preserves line numbers
    Given the following input content:
      """
      %im 0.1
      ---
      {
        a: 1,
        // comment that should not shift lines
        b: 2 + / 3
      }
      """
    When I run the application and it fails
    Then the error should contain "6:10:"

  Scenario: Malformed key attributes report syntax error location
    Given the following input content:
      """
      %im 0.1
      ---
      {
        root: {
          item @(lang: ): "x"
        }
      }
      """
    When I run the application and it fails
    Then the error should contain "6:1:"

  Scenario: Error in nested if-else reports correct line
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (true) (if (false) 1 + / 2 else 0) else -1
      """
    When I run the application and it fails
    Then the error should contain "4:"

  Scenario: Parse error after and/or rewriting reports correct position
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      a and (1 + / 2)
      """
    When I run the application and it fails
    Then the error should contain "4:"

  Scenario: Parse error after map operator rewriting reports exact column
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1,2,3] map ((x) -> x + )
      """
    When I run the application and it fails
    Then the error should contain "4:23:"

  Scenario: Parse error after default and dot-notation rewriting reports exact column
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      payload.user default (1 + )
      """
    When I run the application and it fails
    Then the error should contain "4:27:"

  Scenario: Parse error after configured then operator rewriting reports exact column
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null then (1 + )
      """
    When I run the application and it fails
    Then the error should contain "4:16:"

  Scenario: Parse error after recursive selector rewriting reports exact column
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      items..name map ((x) -> x + )
      """
    When I run the application and it fails
    Then the error should contain "4:27:"

  Scenario: Parse error after top-level object wrapping reports exact column
    Given the following input content:
      """
      %im 0.1
      ---
      name: 1 +
      """
    When I run the application and it fails
    Then the error should contain "3:9:"

  Scenario: Parse error after full preprocessing path reports exact column
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      foo: /a+/ default (1 + )
      """
    When I run the application and it fails
    Then the error should contain "4:24:"

  Scenario: Parse error after as operator config rewriting reports exact column
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      1 as Number {format: 1 + }
      """
    When I run the application and it fails
    Then the error should contain "4:26:"

  Scenario: Parse error in multiline header variable reports exact line
    Given the following input content:
      """
      %im 0.1
      var total =
        1 +
      ---
      total
      """
    When I run the application and it fails
    Then the error should contain "3:5:"

  Scenario: Parse error in multiline header function reports exact line
    Given the following input content:
      """
      %im 0.1
      fun broken(x) =
        x +
      output application/json
      ---
      broken(1)
      """
    When I run the application and it fails
    Then the error should contain "3:5:"

  Scenario: Parse error after multiline if-else rewriting reports exact line
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (true)
        1
      else
        2 +
      """
    When I run the application and it fails
    Then the error should contain "7:5:"

  Scenario: Parse error after regex literal rewriting reports correct column
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      /abc/ + (1 + )
      """
    When I run the application and it fails
    Then the error should contain "4:"

  Scenario: Parse error after implicit object wrapping in map reports correct line
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1] map name: $ +
      """
    When I run the application and it fails
    Then the error should contain "4:"
