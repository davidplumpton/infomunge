Feature: Lambda Expression Support
  In order to use functional programming features like map, filter, and reduce
  As a developer
  I want to define and use lambda (anonymous) functions

  # Representation-only checks are intentionally minimal.
  Scenario: Lambda representation includes expected markers
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

  Scenario: Three-parameter lambda representation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (x, y, z) -> x + y + z
      """
    When I run the application with this content
    Then the output should contain a lambda function

  Scenario: Simple single-parameter lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] map (x) -> x + 1
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,3,4]
      """

  Scenario: Lambda with two parameters
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [10, 20] map (a, b) -> a + b
      """
    When I run the application with this content
    Then the output should be:
      """
      [10,21]
      """

  Scenario: Lambda with string concatenation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["Alice", "Bob"] map (name) -> "Hello " + name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Hello Alice","Hello Bob"]
      """

  Scenario: Explicit map lambda followed by implicit filter lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] map (x) -> (x + 1) filter $ > 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,4]
      """

  Scenario: Lambda with comparison expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [5, 11] map (x) -> x > 10
      """
    When I run the application with this content
    Then the output should be:
      """
      [false,true]
      """

  Scenario: Lambda with if/else expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [-1, 2] map (x) -> if (x > 0) "positive" else "non-positive"
      """
    When I run the application with this content
    Then the output should be:
      """
      ["non-positive","positive"]
      """

  Scenario: Lambda with array access
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[1, 2], [3, 4]] map (arr) -> arr[0]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,3]
      """

  Scenario: Lambda with multiplication
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [2, 3] map (x) -> x * 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [4,6]
      """

  Scenario: Lambda with division
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [4, 10] map (x) -> x / 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,5]
      """

  Scenario: Lambda with nested arithmetic
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 3] map (x) -> (x + 1) * 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [4,8]
      """

  Scenario: Lambda with logical AND
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] map (x, y) -> x > 0 and y > 0
      """
    When I run the application with this content
    Then the output should be:
      """
      [false,true]
      """

  Scenario: Lambda with equality check
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [41, 42] map (x) -> x == 42
      """
    When I run the application with this content
    Then the output should be:
      """
      [false,true]
      """

  Scenario: Lambda with less-than comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [17, 18] map (age) -> age < 18
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,false]
      """

  Scenario: Lambda with less-than-or-equal comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [100, 101] map (score) -> score <= 100
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,false]
      """

  Scenario: Lambda with greater-than-or-equal comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [-1, 0] map (value) -> value >= 0
      """
    When I run the application with this content
    Then the output should be:
      """
      [false,true]
      """

  Scenario: Lambda with field access using dot notation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{name: "alice"}, {name: "bob"}] map (obj) -> obj.name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["alice","bob"]
      """

  Scenario: Lambda with nested field access
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{person: {name: "alice"}}, {person: {name: "bob"}}] map (obj) -> obj.person.name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["alice","bob"]
      """

  Scenario: Lambda with sizeOf function call
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[1], [1, 2, 3]] map (arr) -> sizeOf(arr)
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,3]
      """

  Scenario: Lambda with complex expression using sizeOf
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[], [1]] map (arr) -> sizeOf(arr) > 0
      """
    When I run the application with this content
    Then the output should be:
      """
      [false,true]
      """

  Scenario: Lambda with default operator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [null, 5] map (x) -> x default 0
      """
    When I run the application with this content
    Then the output should be:
      """
      [0,5]
      """

  Scenario: Lambda body with parenthesized subexpression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [3, 5] map (x) -> (x + 5) * (x - 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      [8,30]
      """

  Scenario: Lambda with nested if/else
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [-1, 5, 11] map (x) -> if (x > 10) "high" else if (x > 0) "low" else "zero"
      """
    When I run the application with this content
    Then the output should be:
      """
      ["zero","low","high"]
      """

  Scenario: Lambda with object literal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["alice", "bob"] map (name, age) -> {name: name, age: age}
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"name":"alice","age":0},{"name":"bob","age":1}]
      """

  Scenario: Lambda with array literal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["a", "b"] map (a, b) -> [a, b]
      """
    When I run the application with this content
    Then the output should be:
      """
      [["a",0],["b",1]]
      """

  Scenario: Lambda with underscore in parameter name
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] map (item_value) -> item_value + 1
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,3]
      """

  Scenario: Lambda body accessing variable from context
    Given the following input content:
      """
      %im 0.1
      var multiplier = 3
      output application/json
      ---
      [2, 4] map (x) -> x * multiplier
      """
    When I run the application with this content
    Then the output should be:
      """
      [6,12]
      """

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

  Scenario: Identifier collection pipelines stay inside bounded explicit lambdas
    Given the following script:
      """
      %im 0.1
      var nested = [[1], [2, 3]]
      var values = [1, 2, 3, 4]
      var entries = {a: 1, b: 2}
      output application/json
      fun apply(value, callback) = callback(value)
      ---
      {
        mapped: [0, 0] map (ignored) -> nested map sizeOf($),
        filtered: apply(values, (items) -> items filter ($ > 2)),
        reduced: values then (items) -> items reduce (item, acc = 0) -> acc + item,
        objectMapped: null onNull () -> entries mapObject (value, key) -> {(key): value * 10}
      }
      """
    When I execute the script
    Then the output should be:
      """
      {"mapped":[[1,2],[1,2]],"filtered":[3,4],"reduced":10,"objectMapped":{"a":10,"b":20}}
      """

  Scenario Outline: Nested collection expressions remain inside outer lambda bodies
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      <expression>
      """
    When I execute the script
    Then the output should be:
      """
      <output>
      """

    Examples:
      | expression                                             | output            |
      | [[1]] map ($ map 1)                                   | [[1]]             |
      | [[1]] map ((x) -> x map (y) -> y)                     | [[1]]             |
      | [[1]] map ((x) -> (x map (y) -> y))                   | [[1]]             |
      | [[1, 2], [3]] map ((x) -> x filter ($ > 1))           | [[2],[3]]         |
      | [[1], [2]] map ($ flatMap ((y) -> [y, y]))            | [[1,1],[2,2]]     |
      | [[10, 20], [30]] map ($ map ($ + $$))                 | [[10,21],[30]]    |
      | [[1], [2]] map ((x) -> x flatMap [$, $])              | [[1,1],[2,2]]     |
      | [1] map (i, x) -> [] flatMap (x0) -> [true]           | [[]]              |
      | [1,2] map (x) -> flatten([[x, x + 1]]) reduce (v,a) -> v + a | [3,5]       |
      | [1,2] map (x) -> {a:x,b:x+1} pluck $                  | [[1,2],[2,3]]     |
      | [1,2] map (x) -> [x] ++ [x + 1] map (v) -> v * 2     | [[2,4],[4,6]]     |
      | [1,2] map (x) -> flatten(if (x > 1) [[x]] else [[x + 1]]) filter (v) -> v > 0 | [[2],[2]] |
