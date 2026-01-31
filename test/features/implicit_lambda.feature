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
