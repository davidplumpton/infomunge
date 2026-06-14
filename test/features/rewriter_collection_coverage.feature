Feature: Preprocessor rewriter and collection transformer coverage
  In order to exercise uncovered preprocessor rewriter handler paths
  As a developer
  I want to verify rewrite transforms via the in-process runner

  # --- not operator (handleNot) ---

  Scenario: not operator after space
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      not true
      """
    When I run the script
    Then the output should be false

  Scenario: not operator after opening paren
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (not false) and true
      """
    When I run the script
    Then the output should be true

  Scenario: not operator after opening bracket
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [not true, not false]
      """
    When I run the script
    Then the output should be:
      """
      [false,true]
      """

  Scenario: not operator after comma
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [true, not true]
      """
    When I run the script
    Then the output should be:
      """
      [true,false]
      """

  Scenario: not operator after colon in object
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {result: not false}
      """
    When I run the script
    Then the output should be:
      """
      {"result":true}
      """

  Scenario: not operator combined with and
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      not false and not false
      """
    When I run the script
    Then the output should be true

  # --- Single-quoted string edge cases (processSingleQuoteContent) ---

  Scenario: Single-quoted string with escaped single quote
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      'it\'s here'
      """
    When I run the script
    Then the output should be:
      """
      "it's here"
      """

  Scenario: Single-quoted string with double backslash
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      'back\\slash'
      """
    When I run the script
    Then the output should be:
      """
      "back\\slash"
      """

  Scenario: Single-quoted string containing double quotes
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      'say "hello"'
      """
    When I run the script
    Then the output should be:
      """
      "say \"hello\""
      """

  Scenario: Single-quoted string with other backslash sequence
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      'hello\nworld'
      """
    When I run the script
    Then the output should contain "hello"

  # --- update expression block (handleUpdateExprBlock) via in-process runner ---

  Scenario: Update expression block via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {name: "John", age: 30} update {
        case a at .age -> a + 1
      }
      """
    When I run the script
    Then the output should contain "31"
    And the output should contain "John"

  Scenario: Update expression with nested field via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {user: {name: "John", age: 30}} update {
        case a at .user.age -> a + 5
      }
      """
    When I run the script
    Then the output should contain "35"

  Scenario: Update expression with multiple cases via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {first: "hello", second: "world"} update {
        case f at .first -> upper(f)
        case s at .second -> upper(s)
      }
      """
    When I run the script
    Then the output should contain "HELLO"
    And the output should contain "WORLD"

  Scenario: Chained update expression via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ({a: 1, b: 2} update {
        case x at .a -> x * 10
      }) update {
        case y at .b -> y * 10
      }
      """
    When I run the script
    Then the output should contain "10"
    And the output should contain "20"

  Scenario: Update expression non-existent field via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {name: "John"} update {
        case a at .age -> a + 1
      }
      """
    When I run the script
    Then the output should be:
      """
      {"name":"John"}
      """

  # --- if/else edge cases via runner ---

  Scenario: If without else returns null for false branch via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      if (false) "yes"
      """
    When I run the script
    Then the output should be null

  Scenario: If with and/or in condition via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      if (true and true) "both" else "neither"
      """
    When I run the script
    Then the output should be:
      """
      "both"
      """

  Scenario: If with or in condition via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      if (false or true) "one" else "none"
      """
    When I run the script
    Then the output should be:
      """
      "one"
      """

  Scenario: Nested if/else via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      if (true) (if (false) "inner" else "outer-true") else "outer-false"
      """
    When I run the script
    Then the output should be:
      """
      "outer-true"
      """

  # --- using expression edge cases (handleUsing, findUsingExpressionEnd) ---

  Scenario: using expression terminated by end of input
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      using(x = 5, y = 10) x + y
      """
    When I run the script
    Then the output should be:
      """
      15
      """

  Scenario: using expression inside array terminated by bracket
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [using(x = 3) x * 2, using(y = 4) y * 3]
      """
    When I run the script
    Then the output should be:
      """
      [6,12]
      """

  Scenario: using expression inside parentheses
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (using(x = 7) x + 1) + 2
      """
    When I run the script
    Then the output should be:
      """
      10
      """

  Scenario: using expression with fun declaration
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      using(fun double(n) = n * 2) double(5)
      """
    When I run the script
    Then the output should be:
      """
      10
      """

  # --- do block (handleDoBlock) via runner ---

  Scenario: do block with declaration and expression via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      do {var x = 10---x * 2}
      """
    When I run the script
    Then the output should be:
      """
      20
      """

  Scenario: do block with multiple declarations via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      do {var a = 3
      var b = 4---a + b}
      """
    When I run the script
    Then the output should be:
      """
      7
      """

  Scenario: do block declarations share typed function and multiline var parsing
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      do {
        fun double(n: Number) =
          n * 2
        var payload =
          {
            value: 4
          }
        ---
        double(payload.value)
      }
      """
    When I run the script
    Then the output should be:
      """
      8
      """

  # --- while / break / continue via runner (handleWhileLoop, handleBreak, handleContinue) ---

  Scenario: break at end of expression
    Given the following script:
      """
      var i = 0
      ---
      while(i < 100) {
        if(i == 3) break
        i = i + 1
      }
      i
      """
    When I run the script
    Then the output should be valid JSON with number: 3

  Scenario: continue skips iteration
    Given the following script:
      """
      var i = 0
      var sum = 0
      ---
      while(i < 5) {
        i = i + 1
        if(i == 3) continue
        sum = sum + i
      }
      sum
      """
    When I run the script
    Then the output should be valid JSON with number: 12

  # --- coerce equals operator (handleCoerceEquals) ---

  Scenario: Coerce equals with matching types
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "42" ~= 42
      """
    When I run the script
    Then the output should be true

  Scenario: Coerce equals with non-matching values
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" ~= 42
      """
    When I run the script
    Then the output should be false

  # --- range builtin (handleRangeBuiltin) ---

  Scenario: range builtin via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(4)
      """
    When I run the script
    Then the output should be:
      """
      [0,1,2,3]
      """

  # --- Conditional object literals via runner (handleCloseBrace, tryRewriteConditionalObjectLiteral) ---

  Scenario: Conditional object literal with true condition
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {(a: 1) if (true), (b: 2) if (false)}
      """
    When I run the script
    Then the output should be:
      """
      {"a":1}
      """

  Scenario: Conditional object literal all true
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {(x: 10) if (true), (y: 20) if (true)}
      """
    When I run the script
    Then the output should contain "\"x\":10"
    And the output should contain "\"y\":20"

  Scenario: Conditional object literal all false yields empty object
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {(a: 1) if (false), (b: 2) if (false)}
      """
    When I run the script
    Then the output should be:
      """
      {}
      """

  Scenario: One-line object field uses inline if value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      foo: if (false) 1 else 2
      """
    When I run the script
    Then the output should be:
      """
      {"foo":2}
      """

  # --- if/else false branch terminated by delimiter (findExpressionEnd) ---

  Scenario: If/else false branch inside parentheses
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (if (false) 10 else 20) + 5
      """
    When I run the script
    Then the output should be:
      """
      25
      """

  Scenario: If/else with array expression in false branch
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      if (false) [1] else [2, 3]
      """
    When I run the script
    Then the output should be:
      """
      [2,3]
      """

  Scenario: If/else with object expression in false branch
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      if (false) {a: 1} else {b: 2}
      """
    When I run the script
    Then the output should be:
      """
      {"b":2}
      """

  # --- handleCloseBrace error paths (mismatched braces) ---

  Scenario: Object and array literals combined via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {items: [1, 2, 3], name: "test"}
      """
    When I run the script
    Then the output should contain "\"items\""
    And the output should contain "\"name\""

  Scenario: Nested array in object via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {a: {b: [1, 2]}}
      """
    When I run the script
    Then the output should contain "\"a\""

  # --- break/continue word boundary edge cases ---

  Scenario: break followed by newline
    Given the following script:
      """
      var i = 0
      ---
      while(i < 10) {
        i = i + 1
        if(i == 2) break
      }
      i
      """
    When I run the script
    Then the output should be valid JSON with number: 2

  Scenario: break at end of input
    Given the following script:
      """
      var i = 0
      ---
      while(i < 10) {
        if(i == 0) break
        i = i + 1
      }
      i
      """
    When I run the script
    Then the output should be valid JSON with number: 0

  # --- Collection transformer operators via runner ---

  Scenario: groupBy infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [{type: "a", v: 1}, {type: "b", v: 2}, {type: "a", v: 3}] groupBy (item) -> item.type
      """
    When I run the script
    Then the output should contain "\"a\""
    And the output should contain "\"b\""

  Scenario: orderBy infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [3, 1, 2] orderBy (x) -> x
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: sort infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [3, 1, 2] sort (x) -> x
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: flatMap infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [[1, 2], [3, 4]] flatMap (arr) -> arr
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3,4]
      """

  Scenario: maxBy infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [{v: 1}, {v: 5}, {v: 3}] maxBy (item) -> item.v
      """
    When I run the script
    Then the output should contain "5"

  Scenario: minBy infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [{v: 1}, {v: 5}, {v: 3}] minBy (item) -> item.v
      """
    When I run the script
    Then the output should contain "1"

  Scenario: mapObject infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2} mapObject (val, key) -> {(key): val * 10}
      """
    When I run the script
    Then the output should contain "10"
    And the output should contain "20"

  Scenario: filterObject infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2, c: 3} filterObject (val, key) -> val > 1
      """
    When I run the script
    Then the output should contain "\"b\""
    And the output should contain "\"c\""
    And the output should not contain "\"a\""

  Scenario: pluck infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2, c: 3} pluck (val, key) -> val
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: reduce infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4] reduce (item, acc = 0) -> acc + item
      """
    When I run the script
    Then the output should be:
      """
      10
      """

  Scenario: map infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] map (x) -> x * 2
      """
    When I run the script
    Then the output should be:
      """
      [2,4,6]
      """

  Scenario: filter infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] filter (x) -> x > 3
      """
    When I run the script
    Then the output should be:
      """
      [4,5]
      """

  Scenario: distinctBy infix operator via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 1, 3, 2] distinctBy (x) -> x
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3]
      """
