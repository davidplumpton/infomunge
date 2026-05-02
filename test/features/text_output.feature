Feature: Text output formatting
  In order to emit plain text from scripts
  As a developer
  I want CLI and server output paths to support text/plain

  Scenario: CLI outputs text/plain strings without JSON quotes
    Given the following input content:
      """
      %im 0.1
      output text/plain
      ---
      "hello text"
      """
    When I run the application with this content
    Then the output should be:
      """
      hello text
      """

  Scenario: Run endpoint honors text/plain output from the script header
    Given the server is running
    And the following script:
      """
      %im 0.1
      output text/plain
      ---
      "server text"
      """
    When I run the server script without specifying output
    Then the response status should be 200
    And the output should be:
      """
      server text
      """
