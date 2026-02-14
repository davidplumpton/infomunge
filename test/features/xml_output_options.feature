Feature: XML Output Options
  In order to match DataWeave XML output behavior
  As a developer
  I want to control XML declaration, namespace declarations, and null handling

  Scenario: Disable XML declaration
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaration=false
      ---
      root: "value"
      """
    When I run the script
    Then the output should be:
      """
      <root>value</root>
      """

  Scenario: Write nil on null
    Given the following script:
      """
      %im 0.1
      output application/xml writeNilOnNull=true
      ---
      root: { item: null }
      """
    When I run the script
    Then the output should contain:
      """
      <item xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:nil="true"></item>
      """

  Scenario: Skip null elements
    Given the following script:
      """
      %im 0.1
      output application/xml skipNullOn="elements"
      ---
      root: { item: null, keep: "yes" }
      """
    When I run the script
    Then the output should not contain "item"

  Scenario: Write declared namespaces on root
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces=All
      ns a http://example.com/a
      ns b http://example.com/b
      ---
      root: { child: "value" }
      """
    When I run the script
    Then the output should contain:
      """
      xmlns:a="http://example.com/a"
      """
    And the output should contain:
      """
      xmlns:b="http://example.com/b"
      """

  Scenario: Write declared namespaces by ids filter
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces="ids:a"
      ns a http://example.com/a
      ns b http://example.com/b
      ---
      root: { child: "value" }
      """
    When I run the script
    Then the output should contain:
      """
      xmlns:a="http://example.com/a"
      """
    And the output should not contain "xmlns:b="

  Scenario: Write declared namespaces by regex filter
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces="regex:^a$"
      ns a http://example.com/a
      ns b http://example.com/b
      ---
      root: { child: "value" }
      """
    When I run the script
    Then the output should contain:
      """
      xmlns:a="http://example.com/a"
      """
    And the output should not contain "xmlns:b="

  Scenario: Invalid boolean option value fails parsing
    Given the following input content:
      """
      %im 0.1
      output application/xml writeDeclaration=""
      ---
      root: "value"
      """
    When I run the application and it fails
    Then the application should fail with error containing "output option writeDeclaration"

  Scenario: Invalid output option format fails parsing
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaration
      ---
      root: "value"
      """
    Then running the script should fail with error containing "invalid output option"

  Scenario: Invalid writeDeclaredNamespaces value fails formatting
    Given the following input content:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces=invalid
      ns a http://example.com/a
      ---
      root: "value"
      """
    When I run the application and it fails
    Then the application should fail with error containing "invalid writeDeclaredNamespaces value"

  Scenario: Empty writeDeclaredNamespaces regex fails formatting
    Given the following input content:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces="regex:"
      ns a http://example.com/a
      ---
      root: "value"
      """
    When I run the application and it fails
    Then the application should fail with error containing "regex cannot be empty"

  Scenario: Invalid writeDeclaredNamespaces regex fails formatting
    Given the following input content:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces="regex:*"
      ns a http://example.com/a
      ---
      root: "value"
      """
    When I run the application and it fails
    Then the application should fail with error containing "invalid writeDeclaredNamespaces regex"

  Scenario: Skip null attributes only
    Given the following script:
      """
      %im 0.1
      output application/xml skipNullOn="attributes"
      ---
      root: {
        child @(drop: null, keep: "yes"): "value"
      }
      """
    When I run the script
    Then the output should contain:
      """
      <child keep="yes">value</child>
      """
    And the output should not contain "drop="

  Scenario: Skip null everywhere removes null element and null attributes
    Given the following script:
      """
      %im 0.1
      output application/xml skipNullOn="everywhere"
      ---
      root: {
        gone: null,
        keep @(drop: null): "yes"
      }
      """
    When I run the script
    Then the output should contain:
      """
      <keep>yes</keep>
      """
    And the output should not contain "<gone"
    And the output should not contain "drop="
