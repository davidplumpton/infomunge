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
      <user><age>30</age><name>Alice</name></user>
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
