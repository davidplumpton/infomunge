Feature: CLI file size limits
  In order to avoid unbounded file reads
  As a CLI user
  I want oversized scripts and input files to fail before execution

  Scenario: CLI rejects oversized script files
    Given a file named "large-script.im" with "a" repeated 1048577 times
    When I run the application with "large-script.im" and it fails
    Then the output should contain "script file large-script.im exceeds maximum size"

  Scenario: CLI rejects oversized file-backed inputs
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And a file named "payload.json" with "a" repeated 10485761 times
    When I run the application with this content and inputs and it fails:
      """
      payload=payload.json
      """
    Then the output should contain "input file payload.json exceeds maximum size"
