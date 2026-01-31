Feature: Recursive Descent Operator
  In order to search deeply nested structures
  As a developer
  I want to use the '..' operator to find all matching fields at any depth

  Scenario: Simple recursive field search
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload..name
      """
    When I run the application with this JSON input:
      """
      {"name": "top", "nested": {"name": "inner"}}
      """
    Then the output should be:
      """
      ["top","inner"]
      """

  Scenario: Recursive search in array of objects
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload..name
      """
    When I run the application with this JSON input:
      """
      {"items": [{"name": "Alice"}, {"name": "Bob"}]}
      """
    Then the output should be:
      """
      ["Alice","Bob"]
      """

  Scenario: Mixed nested structures
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload..name
      """
    When I run the application with this JSON input:
      """
      {"name": "top", "items": [{"name": "a"}, {"other": {"name": "b"}}]}
      """
    Then the output should be:
      """
      ["top","a","b"]
      """

  Scenario: No matches returns empty array
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload..missing
      """
    When I run the application with this JSON input:
      """
      {"name": "test", "items": [{"value": 1}]}
      """
    Then the output should be:
      """
      []
      """

  Scenario: Recursive descent with dot notation prefix
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.data..id
      """
    When I run the application with this JSON input:
      """
      {"data": {"users": [{"id": 1}, {"id": 2}]}}
      """
    Then the output should be:
      """
      [1,2]
      """

  Scenario: Deeply nested search
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload..value
      """
    When I run the application with this JSON input:
      """
      {"level1": {"level2": {"level3": {"value": 42}}}}
      """
    Then the output should be:
      """
      [42]
      """

  Scenario: Recursive descent with array index
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.items[0]..name
      """
    When I run the application with this JSON input:
      """
      {"items": [{"users": [{"name": "Alice"}, {"name": "Bob"}]}, {"users": [{"name": "Charlie"}]}]}
      """
    Then the output should be:
      """
      ["Alice","Bob"]
      """

  Scenario: Use sizeOf with recursive descent
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      sizeOf(payload..name)
      """
    When I run the application with this JSON input:
      """
      {"items": [{"name": "a"}, {"name": "b"}, {"name": "c"}]}
      """
    Then the output should be:
      """
      3
      """

  Scenario: Recursive descent combined with map
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload..name map toUpper($)
      """
    When I run the application with this JSON input:
      """
      {"items": [{"name": "alice"}, {"name": "bob"}]}
      """
    Then the output should be:
      """
      ["ALICE","BOB"]
      """
