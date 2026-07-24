Feature: YAML Output Formatting
  In order to output data as YAML
  As a developer
  I want to format data as YAML content

  # Simple objects
  Scenario: Output simple object as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {name: "Alice", age: 30}
      """
    When I run the application with this content
    Then the output should contain "name: Alice"
    And the output should contain "age: 30"

  Scenario: Output object with string values
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {greeting: "Hello", target: "World"}
      """
    When I run the application with this content
    Then the output should contain "greeting: Hello"
    And the output should contain "target: World"

  # Arrays
  Scenario: Output simple array as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      ["apple", "banana", "cherry"]
      """
    When I run the application with this content
    Then the output should contain "- apple"
    And the output should contain "- banana"
    And the output should contain "- cherry"

  Scenario: Output array of numbers
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      [1, 2, 3, 4, 5]
      """
    When I run the application with this content
    Then the output should contain "- 1"
    And the output should contain "- 5"

  # Nested objects
  Scenario: Output nested object as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {person: {name: "Alice", address: {city: "NYC", zip: "10001"}}}
      """
    When I run the application with this content
    Then the output should contain "person:"
    And the output should contain "name: Alice"
    And the output should contain "address:"
    And the output should contain "city: NYC"

  # Mixed content
  Scenario: Output object with array value
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {items: ["a", "b", "c"]}
      """
    When I run the application with this content
    Then the output should contain "items:"
    And the output should contain "- a"
    And the output should contain "- b"
    And the output should contain "- c"

  Scenario: Output array of objects
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      [{name: "Alice", age: 30}, {name: "Bob", age: 25}]
      """
    When I run the application with this content
    Then the output should contain "- name: Alice"
    And the output should contain "age: 30"
    And the output should contain "- name: Bob"
    And the output should contain "age: 25"

  # Primitive values
  Scenario: Output boolean true as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {enabled: true}
      """
    When I run the application with this content
    Then the output should contain "enabled: true"

  Scenario: Output boolean false as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {disabled: false}
      """
    When I run the application with this content
    Then the output should contain "disabled: false"

  Scenario: Output null as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {value: null}
      """
    When I run the application with this content
    Then the output should contain "value: null"

  Scenario: Output float number as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {pi: 3.14159}
      """
    When I run the application with this content
    Then the output should contain "pi: 3.14159"

  # Edge cases
  Scenario: Output empty object as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {}
      """
    When I run the application with this content
    Then the output should contain "{}"

  Scenario: Output empty array as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      []
      """
    When I run the application with this content
    Then the output should contain "[]"

  # Special characters
  Scenario: Output string with colon as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {time: "12:30"}
      """
    When I run the application with this content
    Then the output should contain "time:"

  Scenario: Output string with special characters
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {special: "hello#world"}
      """
    When I run the application with this content
    Then the output should contain "special:"

  # Complex scenarios
  Scenario: Output deeply nested structure as YAML
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {a: {b: {c: {d: "deep"}}}}
      """
    When I run the application with this content
    Then the output should contain "a:"
    And the output should contain "b:"
    And the output should contain "c:"
    And the output should contain "d: deep"

  Scenario: Output mixed types in object
    Given the following input content:
      """
      %im 0.1
      output application/yaml
      ---
      {name: "Test", count: 42, active: true, items: [1, 2, 3]}
      """
    When I run the application with this content
    Then the output should contain "name: Test"
    And the output should contain "count: 42"
    And the output should contain "active: true"
    And the output should contain "items:"
