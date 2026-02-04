Feature: Objects Module
  In order to use DataWeave object helpers
  As a developer
  I want the dw::core::Objects module to be available

  Scenario: entrySet returns object entries
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Objects
      ---
      entrySet({name: "Alice", age: 30})
      """
    When I run the application with this content
    Then the output should contain "\"key\":\"age\""
    And the output should contain "\"value\":30"
    And the output should contain "\"key\":\"name\""
    And the output should contain "\"value\":\"Alice\""

  Scenario: keySet returns object keys
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Objects
      ---
      keySet({z: 1, a: 2, m: 3})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","m","z"]
      """

  Scenario: valueSet returns object values by sorted key
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Objects
      ---
      valueSet({z: "last", a: "first", m: "middle"})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["first","middle","last"]
      """

  Scenario: mergeWith merges objects using right precedence
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Objects
      ---
      mergeWith({a: 1, b: 2}, {b: 20, c: 3})
      """
    When I run the application with this content
    Then the output should contain "\"a\":1"
    And the output should contain "\"b\":20"
    And the output should contain "\"c\":3"

  Scenario: Import module and call with namespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      import dw::core::Objects
      ---
      Objects::keySet({b: 1, a: 2})
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","b"]
      """
