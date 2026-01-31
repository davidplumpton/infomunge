Feature: Header Variable Declarations
  In order to reuse values and simplify expressions
  As a developer
  I want to declare variables in the script header and use them in the body

  Scenario: Use a simple string variable
    Given the following input content:
      """
      %im 0.1
      var name = "InfoMunge"
      output application/json
      ---
      { tool: name }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"tool":"InfoMunge"}
      """

  Scenario: Use a number variable in an expression
    Given the following input content:
      """
      %im 0.1
      var base = 10
      output application/json
      ---
      { result: base * 2 }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"result":20}
      """

  Scenario: Use multiple variables
    Given the following input content:
      """
      %im 0.1
      var firstName = "John"
      var lastName = "Doe"
      output application/json
      ---
      { fullName: firstName + " " + lastName }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"fullName":"John Doe"}
      """

  Scenario: Use a multiline object variable
    Given the following input content:
      """
      %im 0.1
      var config =
        {
          name: "InfoMunge",
          version: 1
        }
      output application/json
      ---
      { tool: config.name, version: config.version }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"tool":"InfoMunge","version":1}
      """
