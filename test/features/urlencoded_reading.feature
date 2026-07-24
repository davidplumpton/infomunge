Feature: URL-encoded Reading
  In order to process form-encoded data
  As a developer
  I want to read and parse URL-encoded content

  Scenario: Read simple URL-encoded key-value pairs
    Given the following URL-encoded input:
      """
      name=Alice&age=30
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      {"name":"Alice","age":"30"}
      """

  Scenario: Read URL-encoded with percent-encoded values
    Given the following URL-encoded input:
      """
      greeting=hello+world&path=%2Ffoo%2Fbar
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      {"greeting":"hello world","path":"/foo/bar"}
      """

  Scenario: Read URL-encoded with multiple values for same key
    Given the following URL-encoded input:
      """
      color=red&color=blue&color=green
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      {"color":["red","blue","green"]}
      """

  Scenario: Access individual URL-encoded field
    Given the following URL-encoded input:
      """
      username=bob&email=bob%40example.com
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.email
      """
    When I run the script
    Then the output should be:
      """
      "bob@example.com"
      """

  Scenario: Read URL-encoded with empty value
    Given the following URL-encoded input:
      """
      key=&other=value
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      {"key":"","other":"value"}
      """
