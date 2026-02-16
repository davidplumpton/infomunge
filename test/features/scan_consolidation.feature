Feature: Consolidated scanning behavior
  Verify that string/bracket-aware splitting and scanning works
  correctly after consolidation of scanning logic.

  Scenario: Object with comma in string value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      { key: "hello, world", num: 42 }
      """
    When I run the script
    Then the output should be:
      """
      {"key":"hello, world","num":42}
      """

  Scenario: Nested brackets in expression
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [[1, 2], [3, 4]][0]
      """
    When I run the script
    Then the output should be:
      """
      [1,2]
      """

  Scenario: Single-line header with quoted string containing keyword
    Given the following script:
      """
      %im 0.1 var msg = "output this var fun" output application/json --- msg
      """
    When I run the script
    Then the output should be:
      """
      "output this var fun"
      """

  Scenario: Variable with balanced brackets
    Given the following script:
      """
      %im 0.1
      var data = { items: [1, (2 + 3)] }
      output application/json
      ---
      data
      """
    When I run the script
    Then the output should be:
      """
      {"items":[1,5]}
      """

  Scenario: Comment after string containing slashes
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "http://example.com" // this is a comment
      """
    When I run the script
    Then the output should be:
      """
      "http://example.com"
      """

  Scenario: Escaped quote inside string with comma
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      { a: "he said \"hi, there\"", b: 1 }
      """
    When I run the script
    Then the output should be:
      """
      {"a":"he said \"hi, there\"","b":1}
      """

  Scenario: Single-quoted string with brackets inside
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      { msg: 'array is [1,2]', val: 3 }
      """
    When I run the script
    Then the output should be:
      """
      {"msg":"array is [1,2]","val":3}
      """
