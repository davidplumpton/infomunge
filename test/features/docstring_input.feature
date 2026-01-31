Feature: Docstring Input
  In order to test the application with complex or multi-line input
  As a developer
  I want to provide input content directly in the feature file using docstrings

  Scenario: Process content from docstring
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "Hello, world!"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello, world!"
      """
