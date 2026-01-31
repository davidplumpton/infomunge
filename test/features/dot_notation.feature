Feature: Dot Notation for Field Access
  In order to access object fields more conveniently
  As a developer
  I want to use dot notation instead of bracket notation

  Scenario: Simple field access with dot notation
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.name
      """
    When I run the application with this JSON input:
      """
      {"name": "Alice", "age": 30}
      """
    Then the output should be:
      """
      "Alice"
      """

  Scenario: Nested field access with dot notation
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.user.email
      """
    When I run the application with this JSON input:
      """
      {"user": {"email": "alice@example.com", "name": "Alice"}}
      """
    Then the output should be:
      """
      "alice@example.com"
      """

  Scenario: Mixed dot and bracket notation
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.items[0].name
      """
    When I run the application with this JSON input:
      """
      {"items": [{"name": "Item1"}, {"name": "Item2"}]}
      """
    Then the output should be:
      """
      "Item1"
      """

  Scenario: Dot notation with non-existent field returns nil
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.missing
      """
    When I run the application with this JSON input:
      """
      {"name": "Alice"}
      """
    Then the output should be:
      """
      null
      """

  Scenario: Bracket notation still works
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload["name"]
      """
    When I run the application with this JSON input:
      """
      {"name": "Bob"}
      """
    Then the output should be:
      """
      "Bob"
      """

  Scenario: Field names with numbers
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.item2
      """
    When I run the application with this JSON input:
      """
      {"item2": "value2"}
      """
    Then the output should be:
      """
      "value2"
      """

  Scenario: Dot notation with underscore field names
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.user_id
      """
    When I run the application with this JSON input:
      """
      {"user_id": 123}
      """
    Then the output should be:
      """
      123
      """

  Scenario: Chained dot notation in expression
    Given the following input content:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      payload.user.age + 1
      """
    When I run the application with this JSON input:
      """
      {"user": {"age": 30}}
      """
    Then the output should be:
      """
      31
      """
