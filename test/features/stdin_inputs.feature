Feature: Stdin-backed CLI inputs
  In order to avoid silent null values from repeated stdin reads
  As a CLI user
  I want stdin-backed inputs to fail fast when more than one input consumes stdin

  Scenario: CLI rejects an explicit stdin-backed input when stdin is already auto-bound to payload
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {payload: payload, extra: extra}
      """
    And the following stdin content:
      """
      {"a":1}
      """
    When I run the application with this content and stdin-backed inputs and it fails:
      """
      extra=:json
      """
    Then the output should contain "multiple stdin-backed inputs are not supported"

  Scenario: CLI rejects oversized stdin-backed input content
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And the stdin content is "a" repeated 10485761 times
    When I run the application with this content and stdin-backed inputs and it fails:
      """
      """
    Then the output should contain "stdin input exceeds maximum size"
