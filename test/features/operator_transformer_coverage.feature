Feature: Operator transformer rewrite coverage
  In order to exercise preprocessor operator transformer rewrite paths
  As a developer
  I want to verify operator rewrites via the in-process runner

  # --- Exponent operator edge cases ---

  Scenario: Exponent with addition operator stops right operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      2 ** 3 + 1
      """
    When I run the script
    Then the output should be:
      """
      9
      """

  Scenario: Exponent with subtraction operator stops right operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      2 ** 3 - 1
      """
    When I run the script
    Then the output should be:
      """
      7
      """

  Scenario: Exponent with multiplication operator stops right operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      2 ** 3 * 2
      """
    When I run the script
    Then the output should be:
      """
      16
      """

  Scenario: Exponent with division operator stops right operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      2 ** 4 / 2
      """
    When I run the script
    Then the output should be:
      """
      8
      """

  Scenario: Exponent with comma in array context
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [2 ** 3, 3 ** 2]
      """
    When I run the script
    Then the output should be:
      """
      [8,9]
      """

  Scenario: Exponent with grouped right operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      2 ** (1 + 2)
      """
    When I run the script
    Then the output should be:
      """
      8
      """

  Scenario: Exponent with grouped left operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (1 + 3) ** 2
      """
    When I run the script
    Then the output should be:
      """
      16
      """

  Scenario: Exponent with negative right operand sign
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      4 ** -1
      """
    When I run the script
    Then the output should be valid JSON with number close to 0.25

  Scenario: Exponent in comparison context
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      2 ** 3 == 8
      """
    When I run the script
    Then the output should be true

  Scenario: Exponent with closing bracket terminates right operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (2 ** 3)
      """
    When I run the script
    Then the output should be:
      """
      8
      """

  # --- Type coercion (as) operator edge cases ---

  Scenario: Coerce string to number with as operator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "42" as Number
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: Coerce number to string with as operator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      42 as String
      """
    When I run the script
    Then the output should be:
      """
      "42"
      """

  Scenario: As operator with expression result
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ("42" as Number) + 8
      """
    When I run the script
    Then the output should be:
      """
      50
      """

  Scenario: Default operator preserves as-operator precedence on rhs
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null default "42" as Number
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: As operator preserves following mod operator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      5 as Number mod 2
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  Scenario: As operator preserves following contains operator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "abc" as String contains "a"
      """
    When I run the script
    Then the output should be true

  Scenario: As operator preserves following concatenate operator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 as String ++ "2"
      """
    When I run the script
    Then the output should be:
      """
      "12"
      """

  Scenario: Default is evaluated before a preceding as coercion
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null as String default "x"
      """
    When I run the script
    Then the output should be:
      """
      "x"
      """

  # --- Type checking (is) operator edge cases ---

  Scenario: Is operator checks Number type
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      42 is Number
      """
    When I run the script
    Then the output should be true

  Scenario: Is operator checks String type
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" is String
      """
    When I run the script
    Then the output should be true

  Scenario: Is operator returns false for wrong type
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      42 is String
      """
    When I run the script
    Then the output should be false

  Scenario: Is operator with null
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null is Null
      """
    When I run the script
    Then the output should be true

  Scenario: Is operator in boolean expression
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      42 is Number and "x" is String
      """
    When I run the script
    Then the output should be true

  # --- Pipe-to-function operator edge cases ---

  Scenario: Pipe to upper function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" | upper
      """
    When I run the script
    Then the output should be:
      """
      "HELLO"
      """

  Scenario: Pipe to sizeOf function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] | sizeOf
      """
    When I run the script
    Then the output should be:
      """
      3
      """

  Scenario: Pipe to abs function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (-5) | abs
      """
    When I run the script
    Then the output should be:
      """
      5
      """

  Scenario: Pipe to flatten function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [[1, 2], [3, 4]] | flatten
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3,4]
      """

  # --- Method-call operator no-left-operand edge ---

  Scenario: Substring method call on string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello world" substring(0, 5)
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  # --- Infix mod operator ---

  Scenario: Infix mod operator basic
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      17 mod 5
      """
    When I run the script
    Then the output should be:
      """
      2
      """

  Scenario: Infix mod operator in comparison
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      10 mod 3 == 1
      """
    When I run the script
    Then the output should be true

  Scenario: Infix mod uses DataWeave precedence and preserves parentheses
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [5 mod 2 * 2, 5 - 2 mod 2, 20 mod 6 / 2, 8 / 4 mod 3, 20 % 6 mod 4, 20 mod 6 mod 4, 20 mod 6 * 4, 20 mod 6 % 4, (8 / 4) mod 3, 8 / (4 mod 3), (5 mod 2) * 2, 5 - (2 mod 2)]
      """
    When I run the script
    Then the output should be:
      """
      [1,1,2,2,2,2,20,0,2,8,2,5]
      """

  # --- Infix repeat operator ---

  Scenario: Infix repeat operator basic
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "ab" repeat 3
      """
    When I run the script
    Then the output should be:
      """
      "ababab"
      """

  # --- Infix splitBy and joinBy operators ---

  Scenario: Infix splitBy operator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "a-b-c" splitBy "-"
      """
    When I run the script
    Then the output should be:
      """
      ["a","b","c"]
      """

  Scenario: Infix joinBy operator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ["a", "b", "c"] joinBy "-"
      """
    When I run the script
    Then the output should be:
      """
      "a-b-c"
      """

  # --- Infix to operator ---

  Scenario: Infix to operator basic range
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 to 5
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3,4,5]
      """

  # --- Infix match and contains operators ---

  Scenario: Infix match operator with regex
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello123" match /([a-z]+)(\d+)/
      """
    When I run the script
    Then the output should be:
      """
      ["hello123","hello","123"]
      """

  Scenario: Infix contains operator with string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello world" contains "world"
      """
    When I run the script
    Then the output should be true

  Scenario: Infix matches operator with regex
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "abc123" matches /^[a-z]+\d+$/
      """
    When I run the script
    Then the output should be true

  # --- Higher-precedence additive left operands ---

  Scenario: Concatenated arrays feed joinBy
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ["a"] ++ ["b"] joinBy "-"
      """
    When I run the script
    Then the output should be:
      """
      "a-b"
      """

  Scenario: Concatenated arrays feed contains
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1] ++ [2] contains 2
      """
    When I run the script
    Then the output should be true

  Scenario: Concatenated strings feed matches as one left operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "a" ++ "b" matches /ab/
      """
    When I run the script
    Then the output should be true

  Scenario: Concatenated strings feed match as one left operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "a" ++ "b" match /(ab)/
      """
    When I run the script
    Then the output should be:
      """
      ["ab","ab"]
      """

  Scenario: Concatenated strings feed scan as one left operand
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "a" ++ "1" scan /([0-9])/
      """
    When I run the script
    Then the output should be:
      """
      [["1","1"]]
      """

  Scenario: Concatenated arrays feed find
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ["a"] ++ ["b"] find "b"
      """
    When I run the script
    Then the output should be:
      """
      [1]
      """

  Scenario: Split and concatenated arrays feed joinBy
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "a,b" splitBy "," ++ ["c"] joinBy "-"
      """
    When I run the script
    Then the output should be:
      """
      "a-b-c"
      """

  Scenario: Removed range values feed joinBy
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 to 3 -- [2] joinBy "-"
      """
    When I run the script
    Then the output should be:
      """
      "1-3"
      """

  # --- Infix default and onNull operators ---

  Scenario: Infix default operator with null
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null default "fallback"
      """
    When I run the script
    Then the output should be:
      """
      "fallback"
      """

  Scenario: Infix onNull operator with null
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null onNull (() -> "recovered")
      """
    When I run the script
    Then the output should be:
      """
      "recovered"
      """

  Scenario: Infix then operator with lambda
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      5 then (x) -> x * 2
      """
    When I run the script
    Then the output should be:
      """
      10
      """

  # --- Concatenate and remove operators ---

  Scenario: Infix concatenate arrays
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] ++ [3, 4]
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3,4]
      """

  Scenario: Infix remove keys from object
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2, c: 3} -- ["b", "c"]
      """
    When I run the script
    Then the output should be:
      """
      {"a":1}
      """

  # --- Update operator ---

  Scenario: Infix update operator merges objects
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2} ~ {b: 3, c: 4}
      """
    When I run the script
    Then the output should contain "\"b\":3"

  # --- Find operator ---

  Scenario: Infix find operator in string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "abcabc" find "bc"
      """
    When I run the script
    Then the output should be:
      """
      [1,4]
      """
