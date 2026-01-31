Feature: Composite Data Types
  In order to represent structured data
  As a developer
  I want to parse and evaluate Arrays and Objects

  Scenario: Evaluate an Array of Numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: Evaluate an Object with Simple Fields
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      { name: "InfoMunge", version: 1 }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"InfoMunge","version":1}
      """

  Scenario: Evaluate nested structure
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      { items: [10, 20], meta: { active: true } }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"items":[10,20],"meta":{"active":true}}
      """