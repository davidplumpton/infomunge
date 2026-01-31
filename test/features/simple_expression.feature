Feature: Simple Expression Evaluation
  In order to verify the basic evaluation engine
  As a developer
  I want to evaluate a simple integer constant expression

  Scenario: Evaluate integer constant 42
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      42
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: using expression binds local variables
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      using(x = 5, y = 10) x + y
      """
    When I run the application with this content
    Then the output should be:
      """
      15
      """

  Scenario: using expression with single-quoted string containing a paren
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      using(x = 'a)') x
      """
    When I run the application with this content
    Then the output should be:
      """
      "a)"
      """

  Scenario: Large script does not trip the rewriter guard
    Given a script with 3000 repeated lines
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: Evaluate simple addition
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      1 + 2
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """
