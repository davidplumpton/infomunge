Feature: README command examples

  Scenario: DataWeave comparison uses the CLI input contract
    Then the README DataWeave example should use a named file input and be executable

  Scenario: Server curl example submits a valid multiline script
    Given the server is running with API key "your-secret"
    When I post the README server example
    Then the response status should be 200
    And the output should be:
      """
      3
      """
