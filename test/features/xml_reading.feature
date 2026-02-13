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

  Scenario: Error when XML element has too many attributes
    Given the following XML input:
      """
      <root a0="v" a1="v" a2="v" a3="v" a4="v" a5="v" a6="v" a7="v" a8="v" a9="v" a10="v" a11="v" a12="v" a13="v" a14="v" a15="v" a16="v" a17="v" a18="v" a19="v" a20="v" a21="v" a22="v" a23="v" a24="v" a25="v" a26="v" a27="v" a28="v" a29="v" a30="v" a31="v" a32="v" a33="v" a34="v" a35="v" a36="v" a37="v" a38="v" a39="v" a40="v" a41="v" a42="v" a43="v" a44="v" a45="v" a46="v" a47="v" a48="v" a49="v" a50="v" a51="v" a52="v" a53="v" a54="v" a55="v" a56="v" a57="v" a58="v" a59="v" a60="v" a61="v" a62="v" a63="v" a64="v" a65="v" a66="v" a67="v" a68="v" a69="v" a70="v" a71="v" a72="v" a73="v" a74="v" a75="v" a76="v" a77="v" a78="v" a79="v" a80="v" a81="v" a82="v" a83="v" a84="v" a85="v" a86="v" a87="v" a88="v" a89="v" a90="v" a91="v" a92="v" a93="v" a94="v" a95="v" a96="v" a97="v" a98="v" a99="v" a100="v" a101="v" a102="v" a103="v" a104="v" a105="v" a106="v" a107="v" a108="v" a109="v" a110="v" a111="v" a112="v" a113="v" a114="v" a115="v" a116="v" a117="v" a118="v" a119="v" a120="v" a121="v" a122="v" a123="v" a124="v" a125="v" a126="v" a127="v" a128="v"></root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "attribute count exceeded"

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
