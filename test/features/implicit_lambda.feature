Feature: Implicit Lambda Parameters ($ and $$)
  In order to write concise functional expressions
  As a developer
  I want to use $ for the current element and $$ for the index

  Scenario: Map with $ for current element
    Given the following input content:
      """
      %im 0.1
      var nums = [1, 2, 3, 4, 5]
      output application/json
      ---
      nums map $ * 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,4,6,8,10]
      """

  Scenario: Map with $ and $$ for element and index
    Given the following input content:
      """
      %im 0.1
      var nums = [10, 20, 30]
      output application/json
      ---
      nums map $ + $$
      """
    When I run the application with this content
    Then the output should be:
      """
      [10,21,32]
      """

  Scenario: Filter with $ for current element
    Given the following input content:
      """
      %im 0.1
      var nums = [1, 2, 3, 4, 5]
      output application/json
      ---
      nums filter $ > 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,4,5]
      """

  Scenario: Filter with $$ for index
    Given the following input content:
      """
      %im 0.1
      var nums = [10, 20, 30, 40, 50]
      output application/json
      ---
      nums filter $$ < 3
      """
    When I run the application with this content
    Then the output should be:
      """
      [10,20,30]
      """

  Scenario: Map with $ accessing object field
    Given the following input content:
      """
      %im 0.1
      var users = [{name: "Alice"}, {name: "Bob"}]
      output application/json
      ---
      users map $.name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice","Bob"]
      """

  Scenario: Map with implicit object literal body
    Given the following input content:
      """
      %im 0.1
      var users = [{name: "Alice"}, {name: "Bob"}]
      output application/json
      ---
      users map user: { name: $.name }
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"user":{"name":"Alice"}},{"user":{"name":"Bob"}}]
      """

  Scenario: Map with $ in arithmetic expression
    Given the following input content:
      """
      %im 0.1
      var nums = [1, 2, 3]
      output application/json
      ---
      nums map ($ * $) + $$
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,5,11]
      """

  Scenario: Filter even numbers using $
    Given the following input content:
      """
      %im 0.1
      var nums = [1, 2, 3, 4, 5, 6]
      output application/json
      ---
      nums filter $ > 3
      """
    When I run the application with this content
    Then the output should be:
      """
      [4,5,6]
      """

  Scenario: Map with $ string concatenation
    Given the following input content:
      """
      %im 0.1
      var names = ["Alice", "Bob"]
      output application/json
      ---
      names map "Hello, " + $
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Hello, Alice","Hello, Bob"]
      """

  Scenario: Chained map and filter with $
    Given the following input content:
      """
      %im 0.1
      var nums = [1, 2, 3, 4, 5]
      output application/json
      ---
      (nums map $ * 2) filter $ > 5
      """
    When I run the application with this content
    Then the output should be:
      """
      [6,8,10]
      """

  Scenario: Chained map and filter with $ without parentheses
    Given the following input content:
      """
      %im 0.1
      var nums = [1, 2, 3, 4, 5]
      output application/json
      ---
      nums map $ * 2 filter $ > 5
      """
    When I run the application with this content
    Then the output should be:
      """
      [6,8,10]
      """

  Scenario: Chained filter and map with $ without parentheses
    Given the following input content:
      """
      %im 0.1
      var nums = [1, 2, 3, 4, 5]
      output application/json
      ---
      nums filter $ > 1 map $ * 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [4,6,8,10]
      """

  Scenario: GroupBy with $ field selector
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "A", "value": 1}, {"name": "B", "value": 1}, {"name": "C", "value": 2}] groupBy $.value
      """
    When I run the application with this content
    Then the output should be:
      """
      {"1":[{"name":"A","value":1},{"name":"B","value":1}],"2":[{"name":"C","value":2}]}
      """

  Scenario: Sort with $ field selector
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"id": 3}, {"id": 1}, {"id": 2}] sort $.id
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"id":1},{"id":2},{"id":3}]
      """

  Scenario: DistinctBy with $ current element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 2, 3] distinctBy $
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: Map object literal with $ and $$ values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [10, 20, 30] map {index: $$, value: $}
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"index":0,"value":10},{"index":1,"value":20},{"index":2,"value":30}]
      """

  Scenario: Explicit lambda still works alongside implicit
    Given the following input content:
      """
      %im 0.1
      var nums = [1, 2, 3]
      output application/json
      ---
      nums map (x) -> x * 3
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,6,9]
      """
