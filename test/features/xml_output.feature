Feature: XML Output
  In order to emit XML data
  As a developer
  I want to specify application/xml as output format

  Scenario: Output as XML
    Given the following input content:
      """
      %im 0.1
      output application/xml
      ---
      { user: { name: "Alice", age: 30 } }
      """
    When I run the application with this content
    Then the output should be:
      """
      <?xml version='1.0' encoding='UTF-8'?>
      <user><name>Alice</name><age>30</age></user>
      """

  Scenario: Escape XML text and attribute values
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaration=false
      ---
      root: {
        item @(label: "A&B \"quoted\" <tag>"): "5 < 6 & 7 > 4"
      }
      """
    When I run the script
    Then the output should be:
      """
      <root><item label="A&amp;B &#34;quoted&#34; &lt;tag&gt;">5 &lt; 6 &amp; 7 &gt; 4</item></root>
      """

  Scenario: Reject multiple XML document roots
    Given the following input content:
      """
      %im 0.1
      output application/xml
      ---
      { first: 1, second: 2 }
      """
    When I run the application and it fails
    Then the error should contain "XML output expects exactly one root element, got 2"

  Scenario: Reject invalid XML element names
    Given the following input content:
      """
      %im 0.1
      output application/xml
      ---
      { "bad key": 1 }
      """
    When I run the application and it fails
    Then the error should contain "invalid XML element name"

  Scenario: Reject unbound XML namespace prefixes
    Given the following input content:
      """
      %im 0.1
      output application/xml
      ---
      { "missing:root": 1 }
      """
    When I run the application and it fails
    Then the error should contain "namespace prefix"
    And the error should contain "is not declared"

  Scenario: Preserve child order with a valid namespace
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaration=false
      ns p urn:example
      ---
      p#root: {
        p#b: 2,
        p#a: 1
      }
      """
    When I run the script
    Then the output should be:
      """
      <p:root xmlns:p="urn:example"><p:b xmlns:p="urn:example">2</p:b><p:a xmlns:p="urn:example">1</p:a></p:root>
      """
