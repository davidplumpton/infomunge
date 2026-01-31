Feature: If/Else Expression
  In order to handle conditional logic in expressions
  As a developer
  I want to use if/else expressions to return different values based on a condition

  Scenario: Simple if/else with literal condition true
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (true) "yes" else "no"
      """
    When I run the application with this content
    Then the output should be:
      """
      "yes"
      """

  Scenario: Simple if/else with literal condition false
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (false) "yes" else "no"
      """
    When I run the application with this content
    Then the output should be:
      """
      "no"
      """

  Scenario: If/else with equality comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (5 == 5) "equal" else "not equal"
      """
    When I run the application with this content
    Then the output should be:
      """
      "equal"
      """

  Scenario: If/else with inequality comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (5 == 10) "equal" else "not equal"
      """
    When I run the application with this content
    Then the output should be:
      """
      "not equal"
      """

  Scenario: If/else with greater than comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (10 > 5) "greater" else "not greater"
      """
    When I run the application with this content
    Then the output should be:
      """
      "greater"
      """

  Scenario: If/else with less than comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (3 < 5) "less" else "not less"
      """
    When I run the application with this content
    Then the output should be:
      """
      "less"
      """

  Scenario: If/else with variable condition
    Given the following input content:
      """
      %im 0.1
      var active = true
      output application/json
      ---
      if (active) "enabled" else "disabled"
      """
    When I run the application with this content
    Then the output should be:
      """
      "enabled"
      """

  Scenario: If/else with variable in condition
    Given the following input content:
      """
      %im 0.1
      var x = 42
      output application/json
      ---
      if (x > 40) "high" else "low"
      """
    When I run the application with this content
    Then the output should be:
      """
      "high"
      """

  Scenario: If/else returning numeric values
    Given the following input content:
      """
      %im 0.1
      var score = 85
      output application/json
      ---
      if (score >= 80) 100 else 50
      """
    When I run the application with this content
    Then the output should be:
      """
      100
      """

  Scenario: Nested if/else expressions
    Given the following input content:
      """
      %im 0.1
      var x = 15
      output application/json
      ---
      if (x > 20) "high" else if (x > 10) "medium" else "low"
      """
    When I run the application with this content
    Then the output should be:
      """
      "medium"
      """

  Scenario: If/else in arithmetic expression
    Given the following input content:
      """
      %im 0.1
      var active = true
      output application/json
      ---
      if (active) 10 else 5 + 3
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  Scenario: If/else with string concatenation result
    Given the following input content:
      """
      %im 0.1
      var user = "Alice"
      output application/json
      ---
      if (user == "Alice") "Hello Alice" else "Hello stranger"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello Alice"
      """

  Scenario: If/else with string containing else and escaped quotes
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (true) "value with else and \"quotes\"" else "fallback"
      """
    When I run the application with this content
    Then the output should be:
      """
      "value with else and \"quotes\""
      """

  Scenario: If/else with single-quoted string containing else
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (true) 'value with else )' else 'fallback'
      """
    When I run the application with this content
    Then the output should be:
      """
      "value with else )"
      """

  Scenario: If/else with no space after if
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if(true) "yes" else "no"
      """
    When I run the application with this content
    Then the output should be:
      """
      "yes"
      """

  Scenario: If/else with logical AND condition
    Given the following input content:
      """
      %im 0.1
      var a = true
      var b = false
      output application/json
      ---
      if (a and b) "both true" else "not both"
      """
    When I run the application with this content
    Then the output should be:
      """
      "not both"
      """

  Scenario: If/else with complex nested condition
    Given the following input content:
      """
      %im 0.1
      var x = 50
      var y = 30
      output application/json
      ---
      if (x > y and x > 40) "x is greater and high" else "condition false"
      """
    When I run the application with this content
    Then the output should be:
      """
      "x is greater and high"
      """

  Scenario: Deeply nested if/else three levels
    Given the following input content:
      """
      %im 0.1
      var x = 15
      output application/json
      ---
      if (x > 20) "too big" else if (x > 10) "medium" else "too small"
      """
    When I run the application with this content
    Then the output should be:
      """
      "medium"
      """

  Scenario: If/else with multiple conditions
    Given the following input content:
      """
      %im 0.1
      var a = true
      var b = true
      output application/json
      ---
      if (a and b) "both true" else "not both"
      """
    When I run the application with this content
    Then the output should be:
      """
      "both true"
      """

  Scenario: If/else with variable in branches
    Given the following input content:
      """
      %im 0.1
      var score = 85
      var message = "pass"
      output application/json
      ---
      if (score >= 80) message else "fail"
      """
    When I run the application with this content
    Then the output should be:
      """
      "pass"
      """

  Scenario: If/else with function call in branches
    Given the following input content:
      """
      %im 0.1
      var flag = true
      output application/json
      ---
      if (flag) sizeOf("hello") else sizeOf("hi")
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: If/else with bracketed expression in branch
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (true) ([1,2,3][1]) else 0
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """
