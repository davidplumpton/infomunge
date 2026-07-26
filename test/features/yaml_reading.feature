Feature: YAML Reading
  In order to process YAML data
  As a developer
  I want to read and parse YAML content

  Scenario: Read a simple YAML string
    Given the following YAML input:
      """
      name: Alice
      age: 30
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
      {"name":"Alice","age":30}
      """

  # Basic Types
  Scenario: Read YAML array
    Given the following YAML input:
      """
      items:
        - apple
        - banana
        - cherry
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.items
      """
    When I run the script
    Then the output should be:
      """
      ["apple","banana","cherry"]
      """

  Scenario: Read YAML string - plain style
    Given the following YAML input:
      """
      message: Hello World
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.message
      """
    When I run the script
    Then the output should be:
      """
      "Hello World"
      """

  Scenario: Read YAML string - quoted style
    Given the following YAML input:
      """
      message: "Hello \"World\""
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.message
      """
    When I run the script
    Then the output should be:
      """
      "Hello \"World\""
      """

  Scenario: Read YAML number - integer vs float distinction
    Given the following YAML input:
      """
      integer: 42
      float_value: 3.14
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should contain "42"

  Scenario: Read YAML boolean and null
    Given the following YAML input:
      """
      enabled: true
      disabled: false
      empty_value: null
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should contain "true"

  # Complex Structures
  Scenario: Read deeply nested YAML object
    Given the following YAML input:
      """
      level1:
        level2:
          level3:
            level4:
              value: "nested"
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.level1.level2.level3.level4.value
      """
    When I run the script
    Then the output should be:
      """
      "nested"
      """

  Scenario: Read YAML array of objects
    Given the following YAML input:
      """
      people:
        - name: Alice
          age: 30
        - name: Bob
          age: 25
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.people[1].name
      """
    When I run the script
    Then the output should be:
      """
      "Bob"
      """

  Scenario: Read YAML with mixed nested structures
    Given the following YAML input:
      """
      config:
        database:
          host: localhost
          ports:
            - 5432
            - 5433
          options:
            timeout: 30
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.config.database.ports
      """
    When I run the script
    Then the output should be:
      """
      [5432,5433]
      """

  # YAML-Specific Features
  Scenario: Reject YAML with multiple documents
    Given the following YAML input:
      """
      ---
      doc1: first
      ---
      doc2: second
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "multiple documents are not supported"

  Scenario: Read YAML with anchors and aliases
    Given the following YAML input:
      """
      template: &default_template
        name: default
        version: 1.0
      production:
        <<: *default_template
        environment: prod
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.production.name
      """
    When I run the script
    Then the output should be:
      """
      "default"
      """

  Scenario: Reject a recursive YAML alias
    Given the following YAML input:
      """
      root: &root
        self: *root
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "recursive alias cycle"

  Scenario: Read YAML with literal scalar
    Given the following YAML input:
      """
      description: |
        This is a literal block.
        It preserves newlines.
        And indentation.
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.description
      """
    When I run the script
    Then the output should contain "literal block"

  Scenario: Read YAML with folded scalar
    Given the following YAML input:
      """
      summary: >
        This is a folded
        block that collapses
        newlines into spaces.
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.summary
      """
    When I run the script
    Then the output should contain "folded"

  # Edge Cases
  Scenario: Read YAML with complex keys
    Given the following YAML input:
      """
      ? complex-key-1
      : value1
      ? complex-key-2
      : value2
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should contain "complex-key-1"

  Scenario: Read empty YAML document
    Given the following YAML input:
      """
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
      null
      """

  # Error Handling
  Scenario: Error on malformed YAML - duplicate keys
    Given the following YAML input:
      """
      key: value1
      key: value2
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "YAML parse error"

  Scenario: Error on malformed YAML - invalid syntax
    Given the following YAML input:
      """
      invalid: [unclosed array
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "YAML parse error"
