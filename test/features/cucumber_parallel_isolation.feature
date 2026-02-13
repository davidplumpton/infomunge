Feature: Cucumber parallel scenario isolation
  Scenarios that use the same relative file names should remain isolated when godog runs in parallel.

  Scenario: first scenario uses input.txt
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      101
      """
    When I run the application with this content
    Then the output should contain "101"

  Scenario: second scenario uses input.txt
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      102
      """
    When I run the application with this content
    Then the output should contain "102"

  Scenario: third scenario uses input.txt
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      103
      """
    When I run the application with this content
    Then the output should contain "103"

  Scenario: fourth scenario uses input.txt
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      104
      """
    When I run the application with this content
    Then the output should contain "104"
