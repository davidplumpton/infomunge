Feature: mapObject with XML Duplicate Keys
  In order to correctly transform XML data with repeated elements
  As a developer
  I want mapObject to iterate over each repeated element as a separate key-value pair

  Scenario: mapObject iterates over repeated XML elements (JSON output)
    Given the following XML input:
      """
      <root>
        <item>1</item>
        <item>2</item>
        <other>3</other>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root mapObject (v, k) -> {
        (k ++ "_transformed"): v
      }
      """
    When I run the script
    Then the output should be:
      """
      {"item_transformed":["1","2"],"other_transformed":"3"}
      """

  Scenario: mapObject iterates over repeated XML elements (XML output)
    Given the following XML input:
      """
      <root>
        <item>1</item>
        <item>2</item>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/xml
      ---
      {
        res: payload.root mapObject (v, k) -> {
          (k ++ "_transformed"): v
        }
      }
      """
    When I run the script
    Then the output should contain "<item_transformed>1</item_transformed>"
    And the output should contain "<item_transformed>2</item_transformed>"

  Scenario: mapObject treats JSON arrays as single entries
    Given the following JSON input:
      """
      {
        "item": [1, 2]
      }
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload mapObject (v, k) -> {
        (k): sizeOf(v)
      }
      """
    When I run the script
    Then the output should be:
      """
      {"item":2}
      """

  Scenario: filterObject iterates over repeated XML elements
    Given the following XML input:
      """
      <root>
        <item>1</item>
        <item>2</item>
        <item>3</item>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root filterObject (v, k) -> v != "2"
      """
    When I run the script
    Then the output should be:
      """
      {"item":["1","3"]}
      """

  Scenario: pluck iterates over repeated XML elements
    Given the following XML input:
      """
      <root>
        <item>1</item>
        <item>2</item>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root pluck (v, k) -> k ++ ":" ++ v
      """
    When I run the script
    Then the output should be:
      """
      ["item:1","item:2"]
      """

  Scenario: keysOf iterates over repeated XML elements
    Given the following XML input:
      """
      <root>
        <item>1</item>
        <item>2</item>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      keysOf(payload.root)
      """
    When I run the script
    Then the output should be:
      """
      ["item","item"]
      """

  Scenario: valuesOf iterates over repeated XML elements
    Given the following XML input:
      """
      <root>
        <item>1</item>
        <item>2</item>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      valuesOf(payload.root)
      """
    When I run the script
    Then the output should be:
      """
      ["1","2"]
      """

  Scenario: entriesOf iterates over repeated XML elements
    Given the following XML input:
      """
      <root>
        <item>1</item>
        <item>2</item>
      </root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      entriesOf(payload.root)
      """
    When I run the script
    Then the output should be:
      """
      [{"key":"item","value":"1"},{"key":"item","value":"2"}]
      """
