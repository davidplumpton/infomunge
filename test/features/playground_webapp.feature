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
    And the output should contain "pretty-print"
    And the output should contain "Pretty print"

  Scenario: Playground page supports standalone runner hook
    Given the server is running
    When I request the playground page
    Then the response status should be 200
    And the output should contain "infomungeRun"

  Scenario: Playground input cards avoid HTML interpolation of input names
    Given the server is running
    When I request the playground page
    Then the response status should be 200
    And the output should contain "nameInput.value = defaultName"
    And the output should not contain "value=\"' + defaultName + '\""

  Scenario: Playground page lists examples
    Given the server is running
    When I request the playground page
    Then the response status should be 200
    And the output should contain "example-picker"
    And the output should contain "Revenue dashboard"
    And the output should contain "Inventory export to CSV"
    And the output should contain "Customer 360"

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

  Scenario: Run endpoint returns timeout when request context is canceled
    Given the server is running
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      "ignored"
      """
    When I run the server script with a canceled request context
    Then the response status should be 408
    And the output should contain "Request timeout"

  Scenario: Run endpoint rejects oversized request bodies
    Given the server is running
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the server script with an oversized request body
    Then the response status should be 413
    And the output should contain "request body exceeds"

  Scenario: Run endpoint rejects requests without API key when configured
    Given the server is running with API key "secret-token"
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      {ok: true}
      """
    When I run the server script without an API key
    Then the response status should be 401
    And the output should contain "unauthorized"

  Scenario: Run endpoint accepts valid API key when configured
    Given the server is running with API key "secret-token"
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      {ok: true}
      """
    When I run the server script with API key "secret-token"
    Then the response status should be 200
    And the output should contain "\"ok\":true"

  Scenario: Run endpoint accepts valid bearer token when API key is configured
    Given the server is running with API key "secret-token"
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      {ok: true}
      """
    When I run the server script with bearer token "secret-token"
    Then the response status should be 200
    And the output should contain "\"ok\":true"
