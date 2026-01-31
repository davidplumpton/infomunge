Feature: Filtering and Transformation Functions
  In order to filter and transform collections
  As a developer
  I want to use the functions: distinctBy, filterObject, find, filter, map

  # Existing filter tests are in filter_operator.feature
  # This file covers distinctBy, filterObject, find, and additional usage patterns

  Scenario: distinctBy basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"id": 1, "type": "A"}, {"id": 2, "type": "A"}, {"id": 3, "type": "B"}, {"id": 4, "type": "B"}] distinctBy (item) -> item.type
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2

  Scenario: distinctBy with numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": 1}, {"val": 2}, {"val": 1}, {"val": 3}, {"val": 2}] distinctBy (item) -> item.val
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 3

  Scenario: distinctBy empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] distinctBy (x) -> x
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: distinctBy single element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice"}] distinctBy (person) -> person.name
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 1

  Scenario: distinctBy with null values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": null}, {"val": 1}, {"val": null}] distinctBy (item) -> item.val
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2

  Scenario: distinctBy preserves order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"order": 3, "cat": "A"}, {"order": 1, "cat": "B"}, {"order": 2, "cat": "A"}] distinctBy (item) -> item.cat
      """
    When I run the application with this content
    Then the output should contain "\"order\":3"

  Scenario: filterObject basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2, "c": 3, "d": 4}, (k, v) -> v > 2)
      """
    When I run the application with this content
    Then the output should be valid JSON with 2 keys
    And the output should contain "\"c\":3"
    And the output should contain "\"d\":4"

  Scenario: filterObject with value filtering
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2, "c": 3, "d": 4}, (k, v) -> v >= 3)
      """
    When I run the application with this content
    Then the output should be valid JSON with 2 keys
    And the output should contain "\"c\":3"

  Scenario: filterObject empty result
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2}, (k, v) -> v > 10)
      """
    When I run the application with this content
    Then the output should be:
      """
      {}
      """

  Scenario: filterObject all pass
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 1, "b": 2, "c": 3}, (k, v) -> v > 0)
      """
    When I run the application with this content
    Then the output should be valid JSON with 3 keys

  Scenario: filterObject with key filtering
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"aa": 5, "bb": 10, "cc": 15}, (k, v) -> k == "bb")
      """
    When I run the application with this content
    Then the output should be valid JSON with 1 keys
    And the output should contain "\"bb\":10"

  Scenario: find basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      find([1, 2, 2, 3], 2)
      """
    When I run the application with this content
    Then the output should contain "1"

  Scenario: find returns indices of matching elements
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      find([10, 20, 30, 20], 20)
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2

  Scenario: find with object value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"id": 1}, {"id": 2}, {"id": 1}] find {"id": 1}
      """
    When I run the application with this content
    Then the output should be:
      """
      [0,2]
      """

  Scenario: find with no match
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      find([1, 2, 3], 99)
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: find with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      find([], 5)
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: find returns all matching indices
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      find([1, 1, 1, 2], 1)
      """
    When I run the application with this content
    Then the output should contain "0"
    And the output should contain "1"
    And the output should contain "2"

  Scenario: map basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4] map (x) -> x * 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,4,6,8]
      """

  Scenario: map with field extraction
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}] map (person) -> person.name
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice","Bob"]
      """

  Scenario: map to objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] map (x) -> {"val": x, "squared": x * x}
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 3
    And the output should contain "\"squared\":4"

  Scenario: map with index
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["a", "b", "c"] map (item, index) -> {"pos": index, "char": item}
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 3
    And the output should contain "\"pos\":0"

  Scenario: map empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] map (x) -> x * 2
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: map with conditional
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] map (x) -> if (x > 2) x * 10 else x
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,30,40,50]
      """

  Scenario: filter basic usage (existing test validation)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] filter (x) -> x > 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,4,5]
      """

  Scenario: distinctBy requires lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] distinctBy "not a lambda"
      """
    When I run the application and it fails
    Then the output should contain "distinctBy expects a lambda function"

  Scenario: distinctBy on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "not an array" distinctBy (x) -> x
      """
    When I run the application and it fails
    Then the output should contain "distinctBy expects an array"

  Scenario: filterObject on non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] filterObject (x) -> x > 1
      """
    When I run the application and it fails
    Then the output should contain "filterObject"

  Scenario: find on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      find({"a": 1}, 1)
      """
    When I run the application and it fails
    Then the output should contain "find expects an array"

  Scenario: map on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "not an array" map (x) -> x * 2
      """
    When I run the application and it fails
    Then the output should contain "map expects an array"

  Scenario: distinctBy with multiple same values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"type": "X", "id": 1}, {"type": "X", "id": 2}, {"type": "X", "id": 3}] distinctBy (x) -> x.type
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 1

  Scenario: filterObject with numeric check
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      filterObject({"a": 5, "b": 15, "c": 10}, (k, v) -> v >= 10)
      """
    When I run the application with this content
    Then the output should be valid JSON with 2 keys
    And the output should contain "\"b\":15"
    And the output should contain "\"c\":10"
