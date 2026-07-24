Feature: Core evaluator runner coverage
  In order to measure cucumber coverage for core evaluator paths
  As a maintainer
  I want fundamental evaluator operations exercised in-process

  Scenario: Descriptor-backed registry dispatches eager and special builtins
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(["a", "b", "c"]) + (if (false) fail("wrong") else 4) + (null default 5)
      """
    When I run the script
    Then the output should be valid JSON with number: 12

  Scenario: Misspelled variable reference fails instead of becoming output data
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      missing
      """
    Then running the script should fail with error containing "unresolved reference: missing"
    And running the script should fail with error containing "4:1:"

  Scenario: Identifiers remain valid in variables, lambdas, builtins, and object keys
    Given the following script:
      """
      %im 0.1
      var increment = 1
      output application/json
      ---
      {values: [1, 2] map (value) -> value + increment, count: sizeOf([1, 2])}
      """
    When I run the script
    Then the output should be:
      """
      {"count":2,"values":[2,3]}
      """

  # --- Logical OR operator (lor) ---

  Scenario: Logical OR with both true
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      true or true
      """
    When I run the script
    Then the output should be true

  Scenario: Logical OR with true and false
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      true or false
      """
    When I run the script
    Then the output should be true

  Scenario: Logical OR with both false
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      false or false
      """
    When I run the script
    Then the output should be false

  Scenario: Logical OR with non-boolean fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 or true
      """
    Then running the script should fail with error containing "logical OR requires booleans"

  # --- Logical NOT operator (not) ---

  Scenario: Logical NOT true
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      !true
      """
    When I run the script
    Then the output should be false

  Scenario: Logical NOT false
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      !false
      """
    When I run the script
    Then the output should be true

  Scenario: Logical NOT with non-boolean fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      !42
      """
    Then running the script should fail with error containing "cannot apply NOT"

  # --- Comparison operators (leq, geq) ---

  Scenario: Less-than-or-equal true case
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      3 <= 3
      """
    When I run the script
    Then the output should be true

  Scenario: Less-than-or-equal false case
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      5 <= 3
      """
    When I run the script
    Then the output should be false

  Scenario: Greater-than-or-equal true case
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      5 >= 5
      """
    When I run the script
    Then the output should be true

  Scenario: Greater-than-or-equal false case
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      3 >= 5
      """
    When I run the script
    Then the output should be false

  # --- Array operators (append <<, prepend >>) ---

  Scenario: Array append with << operator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] << 3
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: << with non-array fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" << 1
      """
    Then running the script should fail with error containing "<< requires array on left side"

  Scenario: >> with non-array on right fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 >> 42
      """
    Then running the script should fail with error containing ">> requires array on right side"

  # --- Subtraction: object key removal (sub) ---

  Scenario: Object key removal with minus operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2, c: 3} - "b"
      """
    When I run the script
    Then the output should contain "a"
    And the output should contain "c"
    And the output should not contain "\"b\""

  Scenario: Numeric subtraction via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      100 - 37
      """
    When I run the script
    Then the output should be:
      """
      63
      """

  # --- Coercion equality (~=) and tryParseNumber ---

  Scenario: Coercion equality string to number
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ["42" ~= 42, "3.14" ~= 3.14, "abc" ~= "abc", 1 ~= 2]
      """
    When I run the script
    Then the output should be:
      """
      [true,true,true,false]
      """

  # --- Lambda MarshalJSON ---

  Scenario: Lambda serialized as JSON string via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (x) -> x + 1
      """
    When I run the script
    Then the output should contain "lambda:"
    And the output should contain "x"

  # --- User-defined function wrong arity (callUserDefinedFunction error) ---

  Scenario: User-defined function wrong arity fails
    Given the following script:
      """
      %im 0.1
      fun greet(name) = "Hi " ++ name
      output application/json
      ---
      greet("Alice", "Bob")
      """
    Then running the script should fail with error containing "expects"

  # --- Error formatting (FormatParseError, FormatEvalError, offsetToLineCol) ---

  Scenario: Parse error reports position via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 + / 2
      """
    Then running the script should fail with error containing ":"

  Scenario: Runtime division by zero with position via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      10 / 0
      """
    Then running the script should fail with error containing "division by zero"

  Scenario: Multiline script parse error maps to correct line
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        a: 1,
        b: 2 + / 3
      }
      """
    Then running the script should fail with error containing ":"

  # --- Combined operator paths ---

  Scenario: Combined logical operators via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (true or false) and (!false)
      """
    When I run the script
    Then the output should be true

  Scenario: Chained comparisons via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1 <= 2, 2 >= 1, 3 <= 3, 4 >= 4, 5 <= 4, 3 >= 5]
      """
    When I run the script
    Then the output should be:
      """
      [true,true,true,true,false,false]
      """

  Scenario: Mixed array append and prepend via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      0 >> ([1] << 2)
      """
    When I run the script
    Then the output should be:
      """
      [0,1,2]
      """

  # --- Equality operators ---

  Scenario: Equality and inequality via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [42 == 42, 42 != 43, "a" == "a", "a" != "b"]
      """
    When I run the script
    Then the output should be:
      """
      [true,true,true,true]
      """

  Scenario: Array and object equality via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [[1, 2] == [1, 2], {a: 1} == {a: 1}, [1] != [2]]
      """
    When I run the script
    Then the output should be:
      """
      [true,true,true]
      """

  # --- Numeric type coercion in operations ---

  Scenario: Mixed int and float arithmetic via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1 + 2.5, 10 - 3.5, 2 * 1.5, 6 / 2.5]
      """
    When I run the script
    Then the output should be:
      """
      [3.5,6.5,3,2.4]
      """

  Scenario: Type error in arithmetic via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      true + 1
      """
    Then running the script should fail with error containing "add"

  # --- ParamNames via multi-param lambda ---

  Scenario: Multi-param lambda with map exercises ParamNames
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [10, 20, 30] map (val, idx) -> {index: idx, doubled: val * 2}
      """
    When I run the script
    Then the output should be:
      """
      [{"doubled":20,"index":0},{"doubled":40,"index":1},{"doubled":60,"index":2}]
      """

  # --- callUserDefinedFunction via try/orElse builtins ---

  Scenario: try with zero-arg lambda invokes callUserDefinedFunction
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try(() -> 42)
      """
    When I run the script
    Then the output should contain "success"
    And the output should contain "true"
    And the output should contain "42"

  Scenario: orElse with successful try returns result
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> 10), 99)
      """
    When I run the script
    Then the output should be:
      """
      10
      """

  Scenario: orElse with failed try invokes fallback lambda
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> (1 / 0)), () -> 99)
      """
    When I run the script
    Then the output should be:
      """
      99
      """

  # --- Logical AND error path ---

  Scenario: Logical AND with non-boolean fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 and true
      """
    Then running the script should fail with error containing "logical AND requires booleans"

  # --- Negation of int ---

  Scenario: Negation of integer via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      -(5)
      """
    When I run the script
    Then the output should be valid JSON with number: -5

  Scenario: Negation of non-numeric fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      -("hello")
      """
    Then running the script should fail with error containing "cannot negate"

  # --- Comparison type error ---

  Scenario: Comparison with non-numeric fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "abc" < 5
      """
    Then running the script should fail with error containing "cannot compare"

  # --- Float validation ---

  Scenario: Division resulting in infinity fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1.0 / 0.0
      """
    Then running the script should fail with error containing "division by zero"

  # --- String indexing (evalIndex string path) ---

  Scenario: String character indexing via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello"[1]
      """
    When I run the script
    Then the output should be:
      """
      "e"
      """

  # --- Coercion equality fallback to string comparison ---

  Scenario: Coercion equality with non-numeric strings
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [true ~= "true", null ~= "null", false ~= true]
      """
    When I run the script
    Then the output should contain "true"

  # --- try with failing lambda exercises callUserDefinedFunction error path ---

  Scenario: try with failing lambda captures error
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try(() -> (1 / 0))
      """
    When I run the script
    Then the output should contain "false"
    And the output should contain "division by zero"

  # --- orElseTry exercises callUserDefinedFunction in orElseTry path ---

  Scenario: orElseTry with failed try and lambda fallback
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> fail("first")), () -> 99)
      """
    When I run the script
    Then the output should contain "success"
    And the output should contain "99"

  # --- Index null returns null ---

  Scenario: Indexing null returns null
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null[0]
      """
    When I run the script
    Then the output should be null

  # --- Negation of float ---

  Scenario: Negation of float via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      -(3.14)
      """
    When I run the script
    Then the output should be valid JSON with number close to -3.14
