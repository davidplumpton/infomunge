Feature: Java Properties Reading
  In order to process Java properties data
  As a developer
  I want to read and parse Java properties content

  Scenario: Read a simple Java properties string
    Given the following properties input:
      """
      name=Alice
      age=30
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      {"name":"Alice","age":"30"}
      """

  Scenario: Read properties with escapes continuation and empty key
    Given the following properties input:
      """
      path\=name\:id=value\=1\:2
      multiline=hello\
       world
      =blank
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      {"path=name:id":"value=1:2","multiline":"helloworld","":"blank"}
      """
