Feature: Array Indexing
  In order to access specific elements in a list
  As a developer
  I want to use index notation to retrieve values from an array

  Scenario: Access first element of an array literal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [10, 20, 30][0]
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  Scenario: Access element using a variable
    Given the following input content:
      """
      %im 0.1
      var list = ["a", "b", "c"]
      output application/json
      ---
      list[1]
      """
    When I run the application with this content
    Then the output should be:
      """
      "b"
      """

  Scenario: Access element using an expression index
    Given the following input content:
      """
      %im 0.1
      var list = [100, 200, 300]
      output application/json
      ---
      list[1 + 1]
      """
    When I run the application with this content
    Then the output should be:
      """
      300
      """

  Scenario: Extract field from all array elements using dot notation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [
        { name: "Alice", age: 30 },
        { name: "Bob", age: 25 },
        { name: "Charlie", age: 35 }
      ].name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice","Bob","Charlie"]
      """

  Scenario: Extract field from array elements with different field names
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "names": [{"name": "Alice"}, {"name": "Bob"}].name,
        "ages": [{"age": 30}, {"age": 25}].age
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"names":["Alice","Bob"],"ages":[30,25]}
      """

  Scenario: Extract field from mixed array elements skips elements without the field
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [
        { name: "Alice", age: 30 },
        { age: 25 },
        { name: "Charlie" }
      ].name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice","Charlie"]
      """

  Scenario: Keyed selector on an array of scalars returns null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [-80.45]["score"]
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: Keyed selector preserves explicit null values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{score: null}, {other: 2}]["score"]
      """
    When I run the application with this content
    Then the output should be:
      """
      [null]
      """

  Scenario: Array field presence reports whether any element has the field
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "present": [{score: 1}, 2].score?,
        "missing": [{other: 1}, 2].score?
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"present":true,"missing":false}
      """

  Scenario: Array field assertion skips elements without the field
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{score: 1}, {other: 2}, 3, {score: 4}].score!
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,4]
      """

  Scenario: Array field assertion fails when no element has the field
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{other: 2}, 3].score!
      """
    When I run the application and it fails
    Then the error should contain "assert selector failed"

  Scenario: Dot notation and keyed notation select immediate array fields
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "dot": [{"name": "A"}, {"child": {"name": "nested"}}, {"name": "B"}].name,
        "keyed": [{"name": "A"}, {"child": {"name": "nested"}}, {"name": "B"}]["name"]
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"dot":["A","B"],"keyed":["A","B"]}
      """

  Scenario: Slice array with range index
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [10, 20, 30, 40, 50][1 to 3]
      """
    When I run the application with this content
    Then the output should be:
      """
      [20,30,40]
      """

  Scenario: Range index inside function call arguments
    Given the following JSON input:
      """
      {"name": "hello"}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(payload.name[0 to 0])
      """
    When I run the script
    Then the output should be:
      """
      "String"
      """

  Scenario: Multiple range indexes in one expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["hello"[0 to 0], "world"[1 to 2]]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["h","or"]
      """

  Scenario: Negative range index bounds
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3][-2 to -1]
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,3]
      """

  Scenario: Descending range index bounds
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        array: [0, 1, 2][-1 to 0],
        text: "é🙂x"[-1 to 0]
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"array":[2,1,0],"text":"x🙂é"}
      """

  Scenario: Out-of-range bounds make ascending and descending range indexes null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        arrayEnd: [0, 1, 2][1 to 5],
        arrayNegativeStart: [0, 1, 2][-5 to 1],
        arrayStart: [0, 1, 2][5 to 7],
        arrayDescending: [0, 1, 2][2 to -5],
        textEnd: "é🙂x"[1 to 5],
        textNegativeStart: "é🙂x"[-5 to 1],
        textStart: "é🙂x"[5 to 7],
        textDescending: "é🙂x"[2 to -5]
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"arrayEnd":null,"arrayNegativeStart":null,"arrayStart":null,"arrayDescending":null,"textEnd":null,"textNegativeStart":null,"textStart":null,"textDescending":null}
      """

  Scenario: Computed range index end
    Given the following input content:
      """
      %im 0.1
      output application/json
      var last = 2
      ---
      [10, 20, 30, 40][1 to last]
      """
    When I run the application with this content
    Then the output should be:
      """
      [20,30]
      """
