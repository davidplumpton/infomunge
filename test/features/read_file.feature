Feature: Read Input File
  In order to process data
  As a user
  I want to provide a file path as an argument and see its content

  Scenario: Read an existing file
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "Hello, Infomunge!"
      """
    When I run the application with this content
    Then the output should contain "Hello, Infomunge!"