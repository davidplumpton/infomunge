Feature: Runner output formatting paths
  In order to verify the runner's built-in output formatting logic
  As a developer
  I want to exercise formatOutputWithContext and related helper functions in-process

  Scenario: JSON output through runner output path
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      { name: "Alice", age: 30 }
      """
    When I run the script through the runner output path
    Then the output should contain "Alice"
    And the output should contain "30"

  Scenario: Large JSON output through runner output path
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(100000)
      """
    When I run the script through the runner output path
    Then the output should be valid JSON with array length of 100000

  Scenario: XML output with declaration through runner output path
    Given the following script:
      """
      %im 0.1
      output application/xml
      ---
      root: "hello"
      """
    When I run the script through the runner output path
    Then the output should contain "<?xml"
    And the output should contain "<root>hello</root>"

  Scenario: XML output with writeDeclaration=false through runner output path
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaration=false
      ---
      root: "value"
      """
    When I run the script through the runner output path
    Then the output should not contain "<?xml"
    And the output should contain "<root>value</root>"

  Scenario: XML output with writeNilOnNull=true through runner output path
    Given the following script:
      """
      %im 0.1
      output application/xml writeNilOnNull=true
      ---
      root: { item: null }
      """
    When I run the script through the runner output path
    Then the output should contain "xsi:nil"

  Scenario: XML output with skipNullOn through runner output path
    Given the following script:
      """
      %im 0.1
      output application/xml skipNullOn="elements"
      ---
      root: { item: null, keep: "yes" }
      """
    When I run the script through the runner output path
    Then the output should not contain "item"
    And the output should contain "<keep>yes</keep>"

  Scenario: CSV output through runner output path
    Given the following script:
      """
      %im 0.1
      output application/csv
      ---
      [{ name: "Bob", age: 25 }, { name: "Eve", age: 31 }]
      """
    When I run the script through the runner output path
    Then the output should contain "Bob"
    And the output should contain "Eve"

  Scenario: Invalid boolean option value fails through runner output path
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaration=""
      ---
      root: "value"
      """
    Then running the script through the runner output path should fail with error containing "output option writeDeclaration"

  Scenario: No header prints raw result through runner output path
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      42
      """
    When I run the script through the runner output path
    Then the output should contain "42"

  Scenario: XML output with namespace declarations through runner output path
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces=All
      ns a http://example.com/a
      ---
      root: { child: "value" }
      """
    When I run the script through the runner output path
    Then the output should contain "xmlns:a"

  Scenario: Output options with brace syntax through runner output path
    Given the following script:
      """
      %im 0.1
      output application/xml {writeDeclaration=false}
      ---
      item: "test"
      """
    When I run the script through the runner output path
    Then the output should not contain "<?xml"
    And the output should contain "<item>test</item>"
