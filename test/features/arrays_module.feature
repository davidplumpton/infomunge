Feature: Arrays Module
  In order to use DataWeave array helpers
  As a developer
  I want the dw::core::Arrays module to be available

  Scenario: countBy counts matching values
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      countBy([1, 2, 3, 4, 5], (x) -> x > 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: Import module and call with namespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      import dw::core::Arrays
      ---
      Arrays::countBy([1, 2, 3, 4, 5], (x) -> x > 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: divideBy splits arrays into chunks
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      divideBy([1, 2, 3, 4, 5], 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      [[1,2],[3,4],[5]]
      """

  Scenario: indexWhere returns the first matching index
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      indexWhere(["a", "b", "c"], (x) -> x == "b")
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: Arrays join combines matches
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      Arrays::join([1, 2], [1, 3], (l) -> l, (r) -> r)
      """
    When I run the application with this content
    Then the output should be valid JSON
    And the output should contain "\"l\":1"
    And the output should contain "\"r\":1"

  Scenario: drop removes leading elements
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      drop([1, 2, 3, 4], 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,4]
      """

  Scenario: dropWhile removes matching leading elements
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      dropWhile([1, 2, 3, 0, 4], (x) -> x < 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,0,4]
      """

  Scenario: every returns true when all items match
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      every([2, 4, 6], (x) -> x > 1)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: firstWith returns the first matching element
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      firstWith([1, 2, 3, 4], (x) -> x > 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: indexOf returns the index of the first match
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      indexOf(["a", "b", "c"], "b")
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: leftJoin includes unmatched left rows
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      leftJoin([{id: 1}, {id: 2}], [{ownerId: 1}, {ownerId: 3}], (l) -> l.id, (r) -> r.ownerId)
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"l":{"id":1},"r":{"ownerId":1}},{"l":{"id":2},"r":null}]
      """

  Scenario: outerJoin includes unmatched right rows
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      outerJoin([{id: 1}, {id: 2}], [{ownerId: 1}, {ownerId: 3}], (l) -> l.id, (r) -> r.ownerId)
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"l":{"id":1},"r":{"ownerId":1}},{"l":{"id":2},"r":null},{"l":null,"r":{"ownerId":3}}]
      """

  Scenario: Import a specific Arrays symbol
    Given the following input content:
      """
      %im 0.1
      output application/json
      import countBy from dw::core::Arrays
      ---
      countBy([1, 2, 3, 4], (x) -> x > 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: partition splits items by criteria
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      partition([1, 2, 3, 4], (x) -> isEven(x))
      """
    When I run the application with this content
    Then the output should be valid JSON
    And the output should contain "\"success\":[2,4]"
    And the output should contain "\"failure\":[1,3]"

  Scenario: slice returns a subarray
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      slice([1, 2, 3, 4, 5], 1, 4)
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,3,4]
      """

  Scenario: some returns true when any item matches
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      some([1, 2, 3], (x) -> x == 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: splitAt divides the array at an index
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      splitAt([1, 2, 3, 4], 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      {"l":[1,2],"r":[3,4]}
      """

  Scenario: splitWhere divides the array at the first match
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      splitWhere([1, 2, 3, 4], (x) -> x > 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      {"l":[1,2],"r":[3,4]}
      """

  Scenario: sumBy sums mapped values
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      sumBy([1, 2, 3], (x) -> x * 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      12
      """

  Scenario: take returns the leading elements
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      take([1, 2, 3, 4], 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2]
      """

  Scenario: takeWhile returns leading matches
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Arrays
      ---
      takeWhile([1, 2, 3, 1], (x) -> x < 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2]
      """
