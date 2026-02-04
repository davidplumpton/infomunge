Feature: Selector Features
  In order to support DataWeave-like selectors
  As a developer
  I want filter, metadata, namespace, presence, and assert selectors

  Scenario: Filter selector on array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{name: "Ana", age: 17}, {name: "Bob", age: 21}][?($.age >= 18)].name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Bob"]
      """

  Scenario: Filter selector on object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2, c: 3}[?($ > 1)]
      """
    When I run the application with this content
    Then the output should be:
      """
      {"b":2,"c":3}
      """

  Scenario: Metadata selector size
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1, b: 2}.^size
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: Presence selector returns boolean
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1}.a?
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Assert selector returns value when key exists
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1}.a!
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: Assert selector fails when key is missing
    Given the following input content:
      """
      %im 0.1
      ---
      {a: 1}.b!
      """
    When I run the application and it fails
    Then the error should contain "assert selector failed"

  Scenario: Namespace selector returns namespace URI
    Given the following XML input:
      """
      <root><h:item xmlns:h="http://example.com/h">value</h:item></root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root.h#item.#
      """
    When I run the script
    Then the output should be:
      """
      "http://example.com/h"
      """
