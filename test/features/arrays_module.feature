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
