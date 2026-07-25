Feature: Null collection operation propagation
  As a transformation author
  I want supported collection helpers to preserve null sources
  So that optional collection pipelines do not fail or change meaning

  Scenario Outline: Array collection operation propagates null
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      <expression>
      """
    When I execute the script
    Then the output should be:
      """
      null
      """

    Examples:
      | expression                      |
      | null map (x) -> x               |
      | null filter (x) -> true         |
      | null flatMap (x) -> [x]         |
      | null orderBy (x) -> x           |
      | null distinctBy (x) -> x        |
      | null groupBy (x) -> x           |

  Scenario Outline: Object collection operation propagates null
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      <expression>
      """
    When I execute the script
    Then the output should be:
      """
      null
      """

    Examples:
      | expression                                    |
      | null pluck (value, key) -> value              |
      | null mapObject (value, key) -> {(key): value} |
      | null filterObject (value, key) -> true        |

  Scenario: Flatten propagates null
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      flatten(null)
      """
    When I execute the script
    Then the output should be:
      """
      null
      """

  Scenario Outline: Unsupported extremum helper continues to reject null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null <builtin> (value) -> value
      """
    When I run the application and it fails
    Then the output should contain "<builtin> expects an array"

    Examples:
      | builtin |
      | maxBy   |
      | minBy   |
