Feature: Input name validation
  In order to avoid ambiguous or shadowed variables
  As a user
  I want CLI and /run input names validated consistently

  Scenario: CLI rejects duplicate input names
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And a file named "payload.json" with content "{\"first\":1}"
    And a file named "payload2.json" with content "{\"second\":2}"
    When I run the application with this content and inputs and it fails:
      """
      payload=payload.json
      payload=payload2.json
      """
    Then the output should contain "duplicate input name \"payload\""

  Scenario: CLI rejects empty input names
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And a file named "payload.json" with content "{\"first\":1}"
    When I run the application with this content and inputs and it fails:
      """
      =payload.json
      """
    Then the output should contain "input name is required"

  Scenario: CLI rejects whitespace-only input names
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And a file named "payload.json" with content "{\"first\":1}"
    When I run the application with this content and inputs and it fails:
      """
         =payload.json
      """
    Then the output should contain "input name is required"

  Scenario: CLI rejects non-identifier input names
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And a file named "payload.json" with content "{\"first\":1}"
    When I run the application with this content and inputs and it fails:
      """
      payload-name=payload.json
      """
    Then the output should contain "invalid input name \"payload-name\""

  Scenario: Run endpoint rejects duplicate input names
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And the server run inputs are:
      """
      [
        {"name":"payload","format":"json","content":"{\"first\":1}"},
        {"name":"payload","format":"json","content":"{\"second\":2}"}
      ]
      """
    When I run the server script with configured inputs
    Then the response status should be 400
    And the output should contain "duplicate input name \"payload\""

  Scenario: Run endpoint rejects empty input names
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And the server run inputs are:
      """
      [
        {"name":"","format":"json","content":"{\"first\":1}"}
      ]
      """
    When I run the server script with configured inputs
    Then the response status should be 400
    And the output should contain "input name is required"

  Scenario: Run endpoint rejects whitespace-only input names
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And the server run inputs are:
      """
      [
        {"name":"   ","format":"json","content":"{\"first\":1}"}
      ]
      """
    When I run the server script with configured inputs
    Then the response status should be 400
    And the output should contain "input name is required"

  Scenario: Run endpoint rejects non-identifier input names
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And the server run inputs are:
      """
      [
        {"name":"payload-name","format":"json","content":"{\"first\":1}"}
      ]
      """
    When I run the server script with configured inputs
    Then the response status should be 400
    And the output should contain "invalid input name \"payload-name\""

  Scenario: Run endpoint rejects too many inputs
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    And I configure 101 server JSON inputs
    When I run the server script with configured inputs
    Then the response status should be 400
    And the output should contain "too many inputs: maximum 100"
