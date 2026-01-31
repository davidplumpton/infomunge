Feature: Array Indexing
  In order to access specific elements in a list
  As a developer
  I want to use index notation to retrieve values from an array

  Scenario: Access first element of an array literal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [10, 20, 30][0]
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  Scenario: Access element using a variable
    Given the following input content:
      """
      %im 0.1
      var list = ["a", "b", "c"]
      output application/json
      ---
      list[1]
      """
    When I run the application with this content
    Then the output should be:
      """
      "b"
      """

  Scenario: Access element using an expression index
    Given the following input content:
      """
      %im 0.1
      var list = [100, 200, 300]
      output application/json
      ---
      list[1 + 1]
      """
    When I run the application with this content
    Then the output should be:
      """
      300
      """

  Scenario: Extract field from all array elements using dot notation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [
        { name: "Alice", age: 30 },
        { name: "Bob", age: 25 },
        { name: "Charlie", age: 35 }
      ].name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice","Bob","Charlie"]
      """

  Scenario: Extract field from array elements with different field names
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "names": [{"name": "Alice"}, {"name": "Bob"}].name,
        "ages": [{"age": 30}, {"age": 25}].age
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"ages":[30,25],"names":["Alice","Bob"]}
      """

  Scenario: Extract field from mixed array elements (with nulls for missing fields)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [
        { name: "Alice", age: 30 },
        { age: 25 },
        { name: "Charlie" }
      ].name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice",null,"Charlie"]
      """

  Scenario: Dot notation on arrays is equivalent to recursive descent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "dot": [{"name": "A"}, {"name": "B"}].name,
        "descent": [{"name": "A"}, {"name": "B"}]..name
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"descent":["A","B"],"dot":["A","B"]}
      """
