Feature: Update Expression
  In order to update nested structures using pattern matching
  As a developer
  I want to use the update expression syntax with case statements

  Scenario: Update single field with case statement
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {name: "John", age: 30} update {
        case age at .age -> age + 1
      }
      """
    When I run the application with this content
    Then the output should contain "John"
    And the output should contain "31"

  Scenario: Update field without changing others
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {name: "John", age: 30, city: "NYC"} update {
        case a at .age -> a * 2
      }
      """
    When I run the application with this content
    Then the output should contain "John"
    And the output should contain "60"
    And the output should contain "NYC"

  Scenario: Update with string transformation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {name: "john", status: "active"} update {
        case n at .name -> upper(n)
      }
      """
    When I run the application with this content
    Then the output should contain "JOHN"
    And the output should contain "active"

  Scenario: Update nested field
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {user: {name: "John", age: 30}} update {
        case a at .user.age -> a + 5
      }
      """
    When I run the application with this content
    Then the output should contain "35"
    And the output should contain "John"

  Scenario: Update array element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {items: [10, 20, 30]} update {
        case x at .items[1] -> x * 2
      }
      """
    When I run the application with this content
    Then the output should contain "40"
    And the output should contain "10"
    And the output should contain "30"

  Scenario: Update the last top-level array element with a negative index
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] update {
        case x at [-1] -> x * 10
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,30]
      """

  Scenario: Update the last nested array element with a negative index
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: [1, 2, 3]} update {
        case x at .a[-1] -> x * 10
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":[1,2,30]}
      """

  Scenario: Multiple case statements
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {first: "hello", second: "world"} update {
        case f at .first -> upper(f)
        case s at .second -> upper(s)
      }
      """
    When I run the application with this content
    Then the output should contain "HELLO"
    And the output should contain "WORLD"

  Scenario: Ancestor update wins over a later descendant update
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: {b: 1}} update {
        case whole at .a -> {b: whole.b + 1}
        case leaf at .a.b -> leaf * 10
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":{"b":2}}
      """

  Scenario: Ancestor update wins over an earlier descendant update
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: {b: 1}} update {
        case leaf at .a.b -> leaf * 10
        case whole at .a -> {b: whole.b + 1}
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":{"b":2}}
      """

  Scenario: First update wins for repeated identical selectors
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: {b: 1}} update {
        case first at .a -> {b: first.b + 1}
        case second at .a -> {b: second.b * 10}
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":{"b":2}}
      """

  Scenario: Disjoint update selectors both apply
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2} update {
        case left at .a -> left + 1
        case right at .b -> right * 10
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":2,"b":20}
      """

  Scenario: Array ancestor update wins over an object descendant update
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {items: [{value: 1}]} update {
        case value at .items[0].value -> value * 10
        case item at .items[0] -> {value: item.value + 1}
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"items":[{"value":2}]}
      """

  Scenario: Update with variable
    Given the following input content:
      """
      %im 0.1
      output application/json
      var person = {name: "John", age: 30}
      ---
      person update {
        case a at .age -> a + 10
      }
      """
    When I run the application with this content
    Then the output should contain "John"
    And the output should contain "40"

  Scenario: Update non-existent field is no-op
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {name: "John"} update {
        case a at .age -> a + 1
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"John"}
      """

  Scenario: Upsert missing field with null binding
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {name: "John"} update {
        case age at .age! -> if (age == null) 30 else age
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"John","age":30}
      """

  Scenario: Upsert nested field and create missing parents
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {name: "John"} update {
        case city at .address.city! -> "New York"
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"John","address":{"city":"New York"}}
      """

  Scenario: Update deeply nested field
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: {b: {c: 100}}} update {
        case x at .a.b.c -> x / 2
      }
      """
    When I run the application with this content
    Then the output should contain "50"

  Scenario: Chained update expressions
    Given the following input content:
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
    When I run the application with this content
    Then the output should contain "10"
    And the output should contain "20"

  Scenario: Update with out-of-bounds selector is a no-op
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: [[1]]} update {
        case x at .a[5][0] -> 99
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":[[1]]}
      """

  Scenario: Update with a negative out-of-bounds selector is a no-op
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: [1, 2, 3]} update {
        case x at .a[-4] -> 99
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":[1,2,3]}
      """

  Scenario: Update expression used in concatenation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ({a: 1} update {
        case x at .a -> x + 1
      }) ++ {b: 2}
      """
    When I run the application with this content
    Then the output should contain "\"a\":2"
    And the output should contain "\"b\":2"

  Scenario: Update block without trailing whitespace before closing brace
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1} update {case x at .a -> x + 1}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":2}
      """

  Scenario: Grouped explicit map callback keeps update in its body
    Given the following script:
      """
      %dw 2.0
      output application/json
      ---
      [{a: 1}, {a: 2}] map ((item) -> item update {
        case value at .a -> value + 10
      })
      """
    When I execute the script
    Then the output should be:
      """
      [{"a":11},{"a":12}]
      """

  Scenario: Grouped completed map applies update to the collection
    Given the following script:
      """
      %dw 2.0
      output application/json
      ---
      ([{a: 1}, {a: 2}] map (item) -> item) update {
        case value at [0].a -> value + 10
      }
      """
    When I execute the script
    Then the output should be:
      """
      [{"a":11},{"a":2}]
      """

  Scenario: Update case inside do block uses the shared nested expression compiler
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      do {
        var base = {a: 1}
        ---
        base update {case x at .a -> if (x == 1) x + 1 else 0}
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":2}
      """

  Scenario: Update expression remains inside an explicit map callback
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{a: 1}, {a: 2}] map (item) -> item update {
        case value at .a -> value * 10
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"a":10},{"a":20}]
      """
