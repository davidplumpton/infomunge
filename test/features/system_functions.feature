Feature: System Functions
  In order to access environment variables
  As a developer
  I want to use the envVar and envVars functions

  Scenario: envVar requires exactly one argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      envVar()
      """
    When I run the application and it fails
    Then the output should contain "envVar requires exactly 1 argument"

  Scenario: envVar with too many arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      envVar("HOME", "extra")
      """
    When I run the application and it fails
    Then the output should contain "envVar requires exactly 1 argument"

  Scenario: envVar expects string argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      envVar(123)
      """
    When I run the application and it fails
    Then the output should contain "envVar expects a string argument"

  Scenario: envVar returns value for existing variable
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      envVar("PATH") != null
      """
    When I run the application with this content
    Then the output should be true

  Scenario: envVar returns null for non-existent variable
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      envVar("INFOMUNGE_NONEXISTENT_VAR_12345")
      """
    When I run the application with this content
    Then the output should be null

  Scenario: envVars takes no arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      envVars("unexpected")
      """
    When I run the application and it fails
    Then the output should contain "envVars takes no arguments"

  Scenario: envVars returns an object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      typeOf(envVars())
      """
    When I run the application with this content
    Then the output should be "Object"

  Scenario: envVars contains PATH
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      envVars().PATH != null
      """
    When I run the application with this content
    Then the output should be true

  Scenario: envVar and envVars return consistent values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      envVar("PATH") == envVars().PATH
      """
    When I run the application with this content
    Then the output should be true
