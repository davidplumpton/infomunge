Feature: CSV Output
  In order to emit tabular data
  As a developer
  I want to specify application/csv as output format

  Scenario: Output as CSV
    Given the following input content:
      """
      %im 0.1
      output application/csv
      ---
      [
        { name: "Alice", age: 30 },
        { name: "Bob", age: 25 }
      ]
      """
    When I run the application with this content
    Then the output should be:
      """
      name,age
      Alice,30
      Bob,25
      """

  Scenario: Output heterogeneous objects with different keys
    Given the following input content:
      """
      %im 0.1
      output application/csv
      ---
      [
        { name: "Alice", age: 30 },
        { name: "Bob", city: "NYC" }
      ]
      """
    When I run the application with this content
    Then the output should be:
      """
      name,age,city
      Alice,30,
      Bob,,NYC
      """

  Scenario: Output objects with missing values
    Given the following input content:
      """
      %im 0.1
      output application/csv
      ---
      [
        { name: "Alice", age: 30, city: "LA" },
        { name: "Bob", age: 25 }
      ]
      """
    When I run the application with this content
    Then the output should be:
      """
      name,age,city
      Alice,30,LA
      Bob,25,
      """

  Scenario: Output empty array
    Given the following input content:
      """
      %im 0.1
      output application/csv
      ---
      []
      """
    When I run the application with this content
    Then the output should be:
      """
      """

  Scenario: Output non-array as CSV raises error
    Given the following input content:
      """
      %im 0.1
      output application/csv
      ---
      { name: "Alice" }
      """
    When I run the application and it fails
    Then the application should fail with error containing "CSV output expects an array of objects"

  Scenario: Output array with non-object rows raises error
    Given the following input content:
      """
      %im 0.1
      output application/csv
      ---
      [
        { name: "Alice" },
        "Bob"
      ]
      """
    When I run the application and it fails
    Then the application should fail with error containing "CSV output expects array elements to be objects"
