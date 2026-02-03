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
