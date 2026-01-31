Feature: Null Handling Functions
  In order to handle null values gracefully
  As a developer
  I want to use onNull and then functions

  # onNull Tests

  Scenario: onNull with non-null value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "valid" onNull (() -> "fallback")
      """
    When I run the application with this content
    Then the output should be:
      """
      "valid"
      """

  Scenario: onNull with null value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null onNull (() -> "fallback")
      """
    When I run the application with this content
    Then the output should be:
      """
      "fallback"
      """

  Scenario: onNull with null value and calculation in lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null onNull (() -> 10 + 20)
      """
    When I run the application with this content
    Then the output should be:
      """
      30
      """

  Scenario: onNull validation - lambda parameter count
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null onNull ((x) -> "fail")
      """
    Then the application should fail with error containing "onNull lambda must have 0 parameters"

  Scenario: onNull validation - not a lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null onNull "not-a-lambda"
      """
    Then the application should fail with error containing "onNull expects a lambda function"

  # then Tests

  Scenario: then with non-null value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" then ((x) -> x ++ " world")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello world"
      """

  Scenario: then with null value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null then ((x) -> x ++ " world")
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: then validation - lambda parameter count
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" then (() -> "fail")
      """
    Then the application should fail with error containing "then lambda must have exactly 1 parameter"

  Scenario: then validation - not a lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" then "not-a-lambda"
      """
    Then the application should fail with error containing "then expects a lambda function"
