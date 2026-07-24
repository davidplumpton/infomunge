Feature: String Indexing
  In order to select text without corrupting Unicode
  As a developer
  I want string indexes and ranges to count characters instead of UTF-8 bytes

  Scenario: Direct string indexes are Unicode-safe
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["éx"[0], "éx"[1], "éx"[-2], "a🙂b"[1], "a🙂b"[-2]]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["é","x","é","🙂","🙂"]
      """

  Scenario: String range indexes are Unicode-safe
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "é🙂x"[0 to 1]
      """
    When I run the application with this content
    Then the output should be:
      """
      "é🙂"
      """

  Scenario: Direct string indexes remain strict when out of range
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "é🙂x"[5]
      """
    When I run the application and it fails
    Then the error should contain "string index out of bounds"
