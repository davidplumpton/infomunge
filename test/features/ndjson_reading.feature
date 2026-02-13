Feature: NDJSON Reading
  In order to process newline-delimited JSON data
  As a developer
  I want to read and parse NDJSON content

  Scenario: Read simple NDJSON records
    Given the following NDJSON input:
      """
      {"name":"Alice","age":30}
      {"name":"Bob","age":25}
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
      [{"age":30,"name":"Alice"},{"age":25,"name":"Bob"}]
      """

  Scenario: Read NDJSON with single record
    Given the following NDJSON input:
      """
      {"id":1,"value":"only"}
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
      [{"id":1,"value":"only"}]
      """

  Scenario: Read NDJSON with blank lines between records
    Given the following NDJSON input:
      """
      {"a":1}

      {"b":2}
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
      [{"a":1},{"b":2}]
      """

  Scenario: Read NDJSON with mixed value types per line
    Given the following NDJSON input:
      """
      {"x":1}
      [1,2,3]
      "hello"
      42
      true
      null
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
      [{"x":1},[1,2,3],"hello",42,true,null]
      """

  Scenario: Access individual NDJSON record by index
    Given the following NDJSON input:
      """
      {"name":"Alice"}
      {"name":"Bob"}
      {"name":"Charlie"}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload[1].name
      """
    When I run the script
    Then the output should be:
      """
      "Bob"
      """

  Scenario: Count NDJSON records
    Given the following NDJSON input:
      """
      {"id":1}
      {"id":2}
      {"id":3}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(payload)
      """
    When I run the script
    Then the output should be:
      """
      3
      """

  Scenario: Error on malformed NDJSON line
    Given the following NDJSON input:
      """
      {"valid":true}
      {invalid json}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "NDJSON parse error"
