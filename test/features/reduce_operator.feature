Feature: Reduce Operator
  In order to aggregate arrays into a single value
  As a developer
  I want to use the reduce operator with lambda expressions

  Scenario: Reduce array to sum
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] reduce (acc, x) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      15
      """

  Scenario: Reduce array to product
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] reduce (acc, x) -> acc * x
      """
    When I run the application with this content
    Then the output should be:
      """
      120
      """

  Scenario: Reduce strings to concatenation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["hello", " ", "world"] reduce (acc, s) -> acc + s
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello world"
      """

  Scenario: Reduce single element array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [42] reduce (acc, x) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Reduce to find maximum value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [3, 1, 4, 1, 5, 9, 2, 6] reduce (max, x) -> if (x > max) x else max
      """
    When I run the application with this content
    Then the output should be:
      """
      9
      """

  Scenario: Reduce to find minimum value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [3, 1, 4, 1, 5, 9, 2, 6] reduce (min, x) -> if (x < min) x else min
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: Reduce with three parameters including index
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [10, 20, 30] reduce (acc, x, idx) -> acc + idx
      """
    When I run the application with this content
    Then the output should be:
      """
      13
      """

  Scenario: Reduce array of objects by extracting values first
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ([{"value": 10}, {"value": 20}, {"value": 30}] map (item) -> item.value) reduce (acc, x) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      60
      """

  Scenario: Reduce to build comma-separated string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["a", "b", "c", "d"] reduce (acc, item) -> acc + "," + item
      """
    When I run the application with this content
    Then the output should be:
      """
      "a,b,c,d"
      """

  Scenario: Reduce with nested if-else for complex logic
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [5, 3, 8, 1, 9, 2] reduce (acc, x) -> if (x > 5) acc + x else acc
      """
    When I run the application with this content
    Then the output should be:
      """
      22
      """

  Scenario: Reduce with multiplication by zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [5, 3, 0, 2, 1] reduce (acc, x) -> acc * x
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: Reduce large array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20] reduce (acc, x) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      210
      """

  Scenario: Reduce after filter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ([1, 2, 3, 4, 5, 6] filter (x) -> x > 3) reduce (acc, x) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      15
      """

  Scenario: Reduce after map
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ([1, 2, 3] map (x) -> x * 2) reduce (acc, x) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      12
      """

  Scenario: Reduce with variable from header
    Given the following input content:
      """
      %im 0.1
      output application/json
      var multiplier = 10
      ---
      [1, 2, 3] reduce (acc, x) -> acc + x * multiplier
      """
    When I run the application with this content
    Then the output should be:
      """
      51
      """

  Scenario: Reduce with initial value (DataWeave style)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] reduce (x, acc = 0) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      6
      """

  Scenario: Reduce with initial value starting from 10
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] reduce (x, acc = 10) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      16
      """

  Scenario: Reduce with initial value on first parameter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] reduce (acc = 5, x) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      11
      """

  Scenario: Reduce empty array with initial value returns initial
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] reduce (x, acc = 42) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Reduce with initial value for multiplication
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [2, 3, 4] reduce (x, product = 1) -> product * x
      """
    When I run the application with this content
    Then the output should be:
      """
      24
      """

  Scenario: Reduce decodes escapes in a string initial value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] reduce (x, acc = "line1\nline2\t\"quoted\"\\path\u263A") -> acc
      """
    When I run the application with this content
    Then the output should be:
      """
      "line1\nline2\t\"quoted\"\\path☺"
      """

  Scenario: Reduce decodes escapes in nested initial values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] reduce (x, acc = ["a\tb", {message: "hello\nworld", nested: ["x\u263Ay", "\\"]}]) -> acc
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a\tb",{"message":"hello\nworld","nested":["x☺y","\\"]}]
      """

  Scenario: Reduce rejects a malformed escape in an initial value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] reduce (x, acc = "bad\q") -> acc
      """
    When I run the application and it fails
    Then the output should contain "invalid default value for parameter acc"
    And the output should contain "invalid string literal"
    And the output should contain "4:11"

  Scenario: Reduce empty array without initial value returns null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] reduce (acc, x) -> acc + x
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: Reduce non-array raises error
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "not an array" reduce (acc, x) -> acc + x
      """
    When I run the application and it fails
    Then the output should contain "reduce expects an array"

  Scenario: Reduce with lambda having only one parameter raises error
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] reduce (x) -> x + 1
      """
    When I run the application and it fails
    Then the output should contain "reduce lambda must have 2 or 3 parameters"

  Scenario: Reduce with parenthesized implicit lambda (no space after reduce)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] reduce($ + $$)
      """
    When I run the application with this content
    Then the output should be:
      """
      6
      """

  Scenario: Reduce with parenthesized implicit lambda with splitBy
    Given the following input content:
      """
      %im 0.1
      output application/json
      fun reduceMapFor(data) = data reduce(($$ splitBy ":")[0] ++ "," ++ ($ splitBy ":")[0])
      ---
      reduceMapFor(["a:b", "c:d"])
      """
    When I run the application with this content
    Then the output should be:
      """
      "c,a"
      """
