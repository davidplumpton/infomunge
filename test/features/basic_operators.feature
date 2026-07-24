Feature: Basic Operators
  In order to perform calculations and logic
  As a developer
  I want to evaluate expressions using standard arithmetic and logical operators

  Scenario: Arithmetic Addition
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      1 + 2
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: Arithmetic Subtraction
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      10 - 4
      """
    When I run the application with this content
    Then the output should be:
      """
      6
      """

  Scenario: Arithmetic Multiplication
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      6 * 7
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Arithmetic Exponent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      2 ** 10
      """
    When I run the application with this content
    Then the output should be:
      """
      1024
      """

  Scenario: Arithmetic Exponent with negative exponent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      2 ** -2
      """
    When I run the application with this content
    Then the output should be valid JSON with number close to 0.25

  Scenario: Arithmetic Exponent with fractional exponent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      9 ** 0.5
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: Arithmetic Exponent with chained operands
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      2 ** 3 ** 2
      """
    When I run the application with this content
    Then the output should be:
      """
      512
      """

  Scenario: Arithmetic Exponent with expression operands
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (1 + 1) ** (2 + 1)
      """
    When I run the application with this content
    Then the output should be:
      """
      8
      """

  Scenario: Arithmetic Exponent precedence with multiplication
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      2 ** 3 * 4
      """
    When I run the application with this content
    Then the output should be:
      """
      32
      """

  Scenario: Arithmetic Exponent precedence with addition
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      2 + 3 ** 2
      """
    When I run the application with this content
    Then the output should be:
      """
      11
      """

  Scenario: Arithmetic Exponent zero to zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      0 ** 0
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: Arithmetic Exponent with negative base and fractional exponent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (-4) ** 0.5
      """
    When I run the application with this content
    Then the output should be:
      """
      NaN
      """

  Scenario: Arithmetic Division
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      20 / 4
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: Arithmetic percent modulo
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [5 % 2, -5 % 2, 5.5 % 2, 2 + 5 % 3 * 4]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,-1,1.5,10]
      """

  Scenario: String Concatenation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "Hello " + "World"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello World"
      """

  Scenario: String Concatenation With Lambda Is Deterministic
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        first: "prefix-" + ((x) -> x + 1),
        second: "prefix-" + ((x) -> x + 1),
        stable: ("prefix-" + ((x) -> x + 1)) == ("prefix-" + ((x) -> x + 1))
      }
      """
    When I run the application with this content
    Then the output should contain "\"first\":\"prefix-\\u003cfunction\\u003e\""
    And the output should contain "\"second\":\"prefix-\\u003cfunction\\u003e\""
    And the output should contain "\"stable\":true"

  Scenario: Relational Equality
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      42 == 42
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Logical AND
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      true and false
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Logical AND skips a failing right operand when false
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      false and (1 / 0 == 0)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Logical OR skips a failing right operand when true
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      true or (1 / 0 == 0)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Logical short-circuit skips an invalid nondeterministic builtin call
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      false and (randomInt(0) == 0)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Array contains operator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] contains 2
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Array prepend with >> operator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      0 >> [1, 2, 3]
      """
    When I run the application with this content
    Then the output should be:
      """
      [0,1,2,3]
      """

  Scenario: Object key removal with - operator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2} - "b"
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":1}
      """

  Scenario: Object key removal for non-existent key
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1, "b": 2} - "c"
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":1,"b":2}
      """

  Scenario: Object key removal from empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {} - "a"
      """
    When I run the application with this content
    Then the output should be:
      """
      {}
      """

  Scenario: Chained object key removal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ({"a": 1, "b": 2, "c": 3} - "a") - "c"
      """
    When I run the application with this content
    Then the output should be:
      """
      {"b":2}
      """

  Scenario: Comparison operators include <= and >=
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [3 <= 3, 3 >= 4]
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,false]
      """

  Scenario: Coercion equality operator with ~=
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["42" ~= 42, "3.14" ~= 3.14]
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,true]
      """

  Scenario: Coercion equality operator ~= with arrays returns false for different arrays
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] ~= [1, 3]
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Coercion equality operator ~= with identical arrays returns true
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] ~= []
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Coercion equality operator ~= with objects returns false for different objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1} ~= {"a": 2}
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Coercion equality operator ~= with identical objects returns true
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"a": 1} ~= {"a": 1}
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Equality compares nested numeric values recursively
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[1] == [1.0], {a: 1} == {a: 1.0}]
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,true]
      """

  Scenario: Coercion equality applies recursively to collections
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [["1"] ~= [1], {a: "1"} ~= {a: 1}]
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,true]
      """

  Scenario: Coercion equality does not expose Go null formatting
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null ~= "<nil>"
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Equality safely compares repeated object values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ({a: 1, a: 2}).a == ({a: 1, a: 2}).a
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Algebraic identity smoke checks
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [
        (7 + 0) == 7,
        (7 * 1) == 7,
        ("x" default "x") == "x",
        ([] ++ [1, 2]) == [1, 2]
      ]
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,true,true,true]
      """
