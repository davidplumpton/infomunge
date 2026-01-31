Feature: Simple Data Types
  In order to represent basic data
  As a developer
  I want to parse and evaluate simple data types like Strings, Numbers, and Booleans

  Scenario: Evaluate a Double Quoted String
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "Hello World"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello World"
      """

  Scenario: Evaluate a Single Quoted String
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      'Hello World'
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello World"
      """

  Scenario: Evaluate a Boolean True
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      true
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Evaluate a Boolean False
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      false
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Evaluate a Floating Point Number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      3.14159
      """
    When I run the application with this content
    Then the output should be:
      """
      3.14159
      """
