Feature: XML Output
  In order to emit XML data
  As a developer
  I want to specify application/xml as output format

  Scenario: Output as XML
    Given the following input content:
      """
      %im 0.1
      output application/xml
      ---
      { user: { name: "Alice", age: 30 } }
      """
    When I run the application with this content
    Then the output should be:
      """
      <user><age>30</age><name>Alice</name></user>
      """
