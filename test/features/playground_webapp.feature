Feature: Playground webapp

  Scenario: Server exposes the playground page
    Given the server is running
    When I request the playground page
    Then the response status should be 200
    And the output should contain "inputs-panel"
    And the output should contain "script-panel"
    And the output should contain "result-panel"
    And the output should contain "Add input"
    And the output should contain "Run Script"

  Scenario: Run endpoint defaults to header output when output is omitted
    Given the server is running
    And the following script:
      """
      %im 0.1
      output application/yaml
      ---
      {name: "Alice"}
      """
    When I run the server script without specifying output
    Then the response status should be 200
    And the output should contain "name: Alice"

  Scenario: Run endpoint honors explicit output override
    Given the server is running
    And the following script:
      """
      %im 0.1
      output application/yaml
      ---
      {name: "Alice"}
      """
    When I run the server script with output "json"
    Then the response status should be 200
    And the output should contain "\"name\""
