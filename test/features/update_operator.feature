Feature: Update Operator (~)
  In order to update nested structures with new values
  As a developer
  I want to use the ~ operator to merge objects

  Scenario: Update single field
    Given the following input content:
      """
      %im 0.1
      output application/json
      var person = {name: "John", age: 30}
      ---
      person ~ {age: 31}
      """
    When I run the application with this content
    Then the output should contain "John"
    And the output should contain "31"

  Scenario: Add new field
    Given the following input content:
      """
      %im 0.1
      output application/json
      var obj = {a: 1}
      ---
      obj ~ {b: 2}
      """
    When I run the application with this content
    Then the output should be valid JSON with 2 keys

  Scenario: Update with variable
    Given the following input content:
      """
      %im 0.1
      output application/json
      var base = {name: "John", age: 30}
      var updates = {age: 31, city: "NYC"}
      ---
      base ~ updates
      """
    When I run the application with this content
    Then the output should contain "John"
    And the output should contain "31"
    And the output should contain "NYC"

  Scenario: Update preserves unchanged fields
    Given the following input content:
      """
      %im 0.1
      output application/json
      var config = {host: "localhost", port: 8080, debug: true}
      ---
      config ~ {port: 3000}
      """
    When I run the application with this content
    Then the output should contain "localhost"
    And the output should contain "3000"
    And the output should contain "true"

  Scenario: Chain multiple updates
    Given the following input content:
      """
      %im 0.1
      output application/json
      var obj = {a: 1}
      ---
      (obj ~ {b: 2}) ~ {c: 3}
      """
    When I run the application with this content
    Then the output should be valid JSON with 3 keys

  Scenario: Update with nested object
    Given the following input content:
      """
      %im 0.1
      output application/json
      var user = {name: "John", settings: {theme: "dark"}}
      ---
      user ~ {settings: {theme: "light"}}
      """
    When I run the application with this content
    Then the output should contain "light"

  Scenario: Update empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      var obj = {}
      ---
      obj ~ {name: "test"}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"test"}
      """

  Scenario: Update with empty object (no change)
    Given the following input content:
      """
      %im 0.1
      output application/json
      var obj = {name: "John"}
      ---
      obj ~ {}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"John"}
      """

  Scenario: Update non-object raises error
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "not an object" ~ {a: 1}
      """
    When I run the application and it fails
    Then the output should contain "update operator (~) expects an object"

  Scenario: Update with non-object on right raises error
    Given the following input content:
      """
      %im 0.1
      output application/json
      var obj = {a: 1}
      ---
      obj ~ "not an object"
      """
    When I run the application and it fails
    Then the output should contain "update operator (~) expects an object"

  Scenario: Update with string value
    Given the following input content:
      """
      %im 0.1
      output application/json
      var user = {name: "John", email: "old@example.com"}
      ---
      user ~ {email: "new@example.com"}
      """
    When I run the application with this content
    Then the output should contain "new@example.com"
    And the output should not contain "old@example.com"

  Scenario: Update with array value
    Given the following input content:
      """
      %im 0.1
      output application/json
      var data = {items: [1, 2]}
      ---
      data ~ {items: [3, 4, 5]}
      """
    When I run the application with this content
    Then the output should contain "[3,4,5]"
