Feature: Collection callback arity
  As a DataWeave author
  I want collection callbacks to accept the parameter prefix they need
  So that callbacks can intentionally ignore supplied values

  Scenario Outline: Array collection callbacks may declare no parameters
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
      <expected>
      """

    Examples:
      | expression                         | expected        |
      | [1,2] map () -> 9                  | [9,9]           |
      | [1,2] filter () -> true            | [1,2]           |
      | [1,2] flatMap () -> [9]            | [9,9]           |
      | [2,1] orderBy () -> 0              | [2,1]           |
      | [1,2] distinctBy () -> 0           | [1]             |
      | [1,2] groupBy () -> "all"          | {"all":[1,2]}   |
      | [1,2] maxBy () -> 0                | 1               |
      | [1,2] minBy () -> 0                | 1               |
      | takeWhile([1,2], () -> false)      | []              |
      | dropWhile([1,2], () -> false)      | [1,2]           |
      | some([1,2], () -> true)            | true            |
      | every([1,2], () -> false)          | false           |

  Scenario: Object collection callbacks may declare no parameters
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        mapped: {a: 1} mapObject () -> {fixed: 9},
        filtered: {a: 1, b: 2} filterObject () -> true,
        plucked: {a: 1, b: 2} pluck () -> 9
      }
      """
    When I execute the script
    Then the output should be:
      """
      {"mapped":{"fixed":9},"filtered":{"a":1,"b":2},"plucked":[9,9]}
      """

  Scenario: Object collection callbacks may declare only the value parameter
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        mapped: {a: 1, b: 2} mapObject (v) -> {((v as String)): v},
        filtered: {a: 1, b: 2} filterObject (v) -> v > 1,
        plucked: {a: 1, b: 2} pluck (v) -> v * 10
      }
      """
    When I execute the script
    Then the output should be:
      """
      {"mapped":{"1":1,"2":2},"filtered":{"b":2},"plucked":[10,20]}
      """

  Scenario: Object collection callbacks expose value key and index
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        mapped: {a: 1, b: 2} mapObject (v, k, i) -> {(k): i},
        filtered: {a: 1, b: 2} filterObject (v, k, i) -> i == 1,
        plucked: {a: 1, b: 2} pluck (v, k, i) -> i
      }
      """
    When I execute the script
    Then the output should be:
      """
      {"mapped":{"a":0,"b":1},"filtered":{"b":2},"plucked":[0,1]}
      """
