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

  Scenario: Filter selector stays inside an implicit map lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{a: [1, 2]}, {a: [0, 3]}] map $.a[?($ == 2)]
      """
    When I run the application with this content
    Then the output should be:
      """
      [[2],[]]
      """

  Scenario: Range selector stays inside an implicit map lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{a: [1, 2]}, {a: [0, 3]}] map $.a[0 to 0]
      """
    When I run the application with this content
    Then the output should be:
      """
      [[1],[0]]
      """

  Scenario: Recursive selector stays inside an implicit map lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{child: {name: "a"}}, {child: {name: "b"}}] map $..name
      """
    When I run the application with this content
    Then the output should be:
      """
      [["a"],["b"]]
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

  Scenario: Metadata selector counts Unicode characters
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "é🙂".^size
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

  Scenario: Quoted selector punctuation remains part of literal keys
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "arrayQuestion": [{"score?": 7}]["score?"],
        "arrayBang": [{"score!": 8}]["score!"],
        "objectQuestion": {"score?": 9}["score?"],
        "objectBang": {"score!": 10}["score!"]
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"arrayQuestion":[7],"arrayBang":[8],"objectQuestion":9,"objectBang":10}
      """

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
