Feature: Namespace Support
  In order to work with XML namespaces
  As a developer
  I want to declare and use XML namespace prefixes

  Scenario: XML output with namespace prefixes and attributes
    Given the following script:
      """
      %im 0.1
      output application/xml
      ns ns0 http://example.com/ns0
      ---
      root: {
        ns0#child @(attr: "val"): "content"
      }
      """
    When I run the script
    Then the output should contain:
      """
      <root xmlns:ns0="http://example.com/ns0"><ns0:child attr="val">content</ns0:child></root>
      """
