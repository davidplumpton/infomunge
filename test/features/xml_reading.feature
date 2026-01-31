Feature: XML Reading
  In order to process XML data
  As a developer
  I want to read and parse XML content

  Scenario: Read a simple XML string
    Given the following XML input:
      """
      <user><name>Alice</name><age>30</age></user>
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
      {"user":{"age":"30","name":"Alice"}}
      """

  Scenario: Read XML with declaration header
    Given the following XML input:
      """
      <?xml version="1.0" encoding="UTF-8"?>
      <user><name>Alice</name></user>
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
      {"user":{"name":"Alice"}}
      """

  Scenario: Access XML element by integer index
    Given the following XML input:
      """
      <language><name>DataWeave</name><version>2.0</version></language>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      { version : payload.language[1] }
      """
    When I run the script
    Then the output should be:
      """
      {"version":"2.0"}
      """

  Scenario: Access repeated XML element by index
    Given the following XML input:
      """
      <root><element><subelement1>SE1</subelement1></element><element>E2</element></root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root.element[0]
      """
    When I run the script
    Then the output should be:
      """
      {"subelement1":"SE1"}
      """

  Scenario: Access XML child elements using ordinal indexing
    Given the following XML input:
      """
      <root><first>1</first><second>2</second><third>3</third></root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      { a: payload.root[0], b: payload.root[1], c: payload.root[2] }
      """
    When I run the script
    Then the output should be:
      """
      {"a":"1","b":"2","c":"3"}
      """

  Scenario: Access XML attributes using @ selector
    Given the following XML input:
      """
      <root version="1.0">
        <element attr="value">content</element>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        version: payload.root.@version,
        attr: payload.root.element.@attr
      }
      """
    When I run the script
    Then the output should be:
      """
      {"attr":"value","version":"1.0"}
      """

  Scenario: Access nested XML attributes
    Given the following XML input:
      """
      <order id="123">
        <customer id="456">John Doe</customer>
      </order>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        orderId: payload.order.@id,
        customerId: payload.order.customer.@id
      }
      """
    When I run the script
    Then the output should be:
      """
      {"customerId":"456","orderId":"123"}
      """

  # Basic Structure
  Scenario: Read XML with self-closing tags
    Given the following XML input:
      """
      <root>
        <empty/>
        <item>value</item>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root
      """
    When I run the script
    Then the output should contain "empty"

  Scenario: Read XML with mixed content (text and child elements)
    Given the following XML input:
      """
      <root>
        Some text here
        <child>Child content</child>
        More text here
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root
      """
    When I run the script
    Then the output should contain "Some text"

  Scenario: Read XML with numeric content
    Given the following XML input:
      """
      <root>
        <values>
          <item>100</item>
          <item>200</item>
          <item>300</item>
        </values>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(payload.root.values.item)
      """
    When I run the script
    Then the output should be:
      """
      3
      """

  # Nested and Complex
  Scenario: Read deeply nested XML (10+ levels)
    Given the following XML input:
      """
      <l1><l2><l3><l4><l5><l6><l7><l8><l9><l10>Deep value</l10></l9></l8></l7></l6></l5></l4></l3></l2></l1>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.l1.l2.l3.l4.l5.l6.l7.l8.l9.l10
      """
    When I run the script
    Then the output should be:
      """
      "Deep value"
      """

  Scenario: Read XML with repeated nested elements (arrays within arrays)
    Given the following XML input:
      """
      <root>
        <section id="1">
          <item>A1</item>
          <item>A2</item>
        </section>
        <section id="2">
          <item>B1</item>
          <item>B2</item>
        </section>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root.section[1].item[0]
      """
    When I run the script
    Then the output should be:
      """
      "B1"
      """

  Scenario: Read XML with attributes on complex element
    Given the following XML input:
      """
      <root>
        <response status="200" type="success">
          <data>content</data>
        </response>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        status: payload.root.response.@status,
        type: payload.root.response.@type,
        data: payload.root.response.data
      }
      """
    When I run the script
    Then the output should contain "200"

  # Namespaces
  Scenario: Read XML with default namespace
    Given the following XML input:
      """
      <root xmlns="http://example.com/ns">
        <element>value</element>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root
      """
    When I run the script
    Then the output should contain "element"

  Scenario: Read XML with multiple namespace prefixes
    Given the following XML input:
      """
      <soap:Envelope xmlns:soap="http://example.com/soap" xmlns:app="http://example.com/app">
        <soap:Body>
          <app:Message>Hello</app:Message>
        </soap:Body>
      </soap:Envelope>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should contain "Envelope"

  Scenario: Read XML with namespace-qualified attributes
    Given the following XML input:
      """
      <root xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance" xsi:type="custom">
        <value>test</value>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root
      """
    When I run the script
    Then the output should contain "value"

  # Edge Cases
  Scenario: Read XML with comments (ignored)
    Given the following XML input:
      """
      <root>
        <!-- This is a comment -->
        <element>value</element>
        <!-- Another comment -->
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root.element
      """
    When I run the script
    Then the output should be:
      """
      "value"
      """

  Scenario: Read XML with processing instructions
    Given the following XML input:
      """
      <?xml version="1.0"?>
      <?xml-stylesheet type="text/xsl" href="style.xsl"?>
      <root>
        <element>content</element>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root.element
      """
    When I run the script
    Then the output should be:
      """
      "content"
      """

  Scenario: Read empty element vs element with empty text
    Given the following XML input:
      """
      <root>
        <empty_self_closing/>
        <empty_with_tags></empty_with_tags>
        <with_whitespace>   </with_whitespace>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root
      """
    When I run the script
    Then the output should contain "empty_self_closing"

  # Error Handling
  Scenario: Error on unclosed XML tags
    Given the following XML input:
      """
      <root>
        <element>value
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "XML validation error"

  Scenario: Error on mismatched XML tags
    Given the following XML input:
      """
      <root>
        <element>value</wrong_tag>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "XML validation error"

  Scenario: Read XML with DOCTYPE declaration
    Given the following XML input:
      """
      <?xml version="1.0"?>
      <!DOCTYPE root [
        <!ELEMENT root (item)>
        <!ELEMENT item (#PCDATA)>
      ]>
      <root>
        <item>content</item>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root.item
      """
    When I run the script
    Then the output should be:
      """
      "content"
      """