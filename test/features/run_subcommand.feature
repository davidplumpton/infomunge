Feature: Optional run subcommand
  In order to be compatible with DataWeave CLI conventions
  As a user
  I want to optionally use "run" as a subcommand

  Scenario: Run subcommand works the same as without it
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "Hello from run!"
      """
    When I run the application with the run subcommand
    Then the output should contain "Hello from run!"

  Scenario: Run subcommand works with expressions
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      1 + 2
      """
    When I run the application with the run subcommand
    Then the output should contain "3"
