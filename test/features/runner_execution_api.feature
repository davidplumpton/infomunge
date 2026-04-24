Feature: Runner execution API
  In order to keep evaluation separate from output adapters
  As a developer
  I want structured runner execution and CLI output to stay behavior-compatible

  Scenario: Structured runner API returns value and output metadata
    Given the following JSON input:
      """
      {"value": 41}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      { answer: payload.value + 1 }
      """
    When I execute the script through the structured runner API
    Then the output should be:
      """
      {"answer":42}
      """
    And the runner output MIME type should be "application/json"

  Scenario: CLI output adapter prints formatted execution result
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      { city: "Wellington", count: 2 }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"city":"Wellington","count":2}
      """
