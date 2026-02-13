Feature: NDJSON Output
  In order to produce newline-delimited JSON data
  As a developer
  I want to write data in NDJSON format

  Scenario: Output array of objects as NDJSON
    Given the following script:
      """
      %im 0.1
      output application/x-ndjson
      ---
      [
        { name: "Alice", age: 30 },
        { name: "Bob", age: 25 }
      ]
      """
    When I run the script
    Then the output should be:
      """
      {"age":30,"name":"Alice"}
      {"age":25,"name":"Bob"}
      """

  Scenario: Output single-element array as NDJSON
    Given the following script:
      """
      %im 0.1
      output application/x-ndjson
      ---
      [{ id: 1, value: "only" }]
      """
    When I run the script
    Then the output should be:
      """
      {"id":1,"value":"only"}
      """

  Scenario: Output array of mixed types as NDJSON
    Given the following script:
      """
      %im 0.1
      output application/x-ndjson
      ---
      [{ x: 1 }, [1, 2, 3], "hello", 42, true, null]
      """
    When I run the script
    Then the output should be:
      """
      {"x":1}
      [1,2,3]
      "hello"
      42
      true
      null
      """

  Scenario: Round-trip NDJSON read and write
    Given the following NDJSON input:
      """
      {"name":"Alice","age":30}
      {"name":"Bob","age":25}
      """
    And the following script:
      """
      %im 0.1
      output application/x-ndjson
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      {"age":30,"name":"Alice"}
      {"age":25,"name":"Bob"}
      """

  Scenario: Error when NDJSON output is not an array
    Given the following input content:
      """
      %im 0.1
      output application/x-ndjson
      ---
      { name: "Alice" }
      """
    When I run the application and it fails
    Then the application should fail with error containing "NDJSON output expects an array"

  Scenario: Output NDJSON using application/ndjson alias
    Given the following script:
      """
      %im 0.1
      output application/ndjson
      ---
      [{ a: 1 }, { b: 2 }]
      """
    When I run the script
    Then the output should be:
      """
      {"a":1}
      {"b":2}
      """
