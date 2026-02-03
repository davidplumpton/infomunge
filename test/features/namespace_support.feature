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
      <root><ns0:child xmlns:ns0="http://example.com/ns0" attr="val">content</ns0:child></root>
      """

  Scenario: Namespace type variable in element and attribute
    Given the following script:
      """
      %im 0.1
      output application/xml
      var myNS = { uri: "http://example.com/dyn", prefix: "ns-0" } as Namespace
      ---
      root: {
        myNS#child @(myNS#attr: "yes"): "content"
      }
      """
    When I run the script
    Then the output should contain:
      """
      <ns-0:child xmlns:ns-0="http://example.com/dyn" ns-0:attr="yes">content</ns-0:child>
      """
