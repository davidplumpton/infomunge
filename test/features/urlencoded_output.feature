Feature: URL-encoded Output
  In order to produce form-encoded data
  As a developer
  I want to write data in URL-encoded format

  Scenario: Output object as URL-encoded
    Given the following script:
      """
      %im 0.1
      output application/x-www-form-urlencoded
      ---
      { name: "Alice", age: "30" }
      """
    When I run the script
    Then the output should be:
      """
      name=Alice&age=30
      """

  Scenario: Output object with special characters
    Given the following script:
      """
      %im 0.1
      output application/x-www-form-urlencoded
      ---
      { greeting: "hello world", path: "/foo/bar" }
      """
    When I run the script
    Then the output should be:
      """
      greeting=hello+world&path=%2Ffoo%2Fbar
      """

  Scenario: Round-trip URL-encoded read and write
    Given the following URL-encoded input:
      """
      name=Alice&city=New+York
      """
    And the following script:
      """
      %im 0.1
      output application/x-www-form-urlencoded
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      name=Alice&city=New+York
      """

  Scenario: Error when URL-encoded output is not an object
    Given the following input content:
      """
      %im 0.1
      output application/x-www-form-urlencoded
      ---
      [1, 2, 3]
      """
    When I run the application and it fails
    Then the application should fail with error containing "urlencoded output expects an object"

  Scenario: Output object with numeric values
    Given the following script:
      """
      %im 0.1
      output application/x-www-form-urlencoded
      ---
      { count: 42, rate: 3.14 }
      """
    When I run the script
    Then the output should be:
      """
      count=42&rate=3.14
      """

  Scenario: Output object with boolean values
    Given the following script:
      """
      %im 0.1
      output application/x-www-form-urlencoded
      ---
      { active: true, deleted: false }
      """
    When I run the script
    Then the output should be:
      """
      active=true&deleted=false
      """
