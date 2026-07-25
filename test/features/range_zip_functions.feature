Feature: Range and Zip Functions
  In order to work with sequences and combine arrays
  As a developer
  I want to use the range, to, zip, and unzip functions

  # Range function tests
  Scenario: range basic usage
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(5)
      """
    When I run the script
    Then the output should be:
      """
      [0,1,2,3,4]
      """

  Scenario: range with zero
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(0)
      """
    When I run the script
    Then the output should be:
      """
      []
      """

  Scenario: range with one
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(1)
      """
    When I run the script
    Then the output should be:
      """
      [0]
      """

  Scenario: range with large number
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(100)
      """
    When I run the script
    Then the output should be valid JSON
    And the output should contain "[0,1,2"

  Scenario: range with negative number fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(-5)
      """
    Then running the script should fail with error containing "range end must be non-negative"

  Scenario: range preserves an exact bound immediately above 2^53 before validation
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(-9007199254740993)
      """
    Then running the script should fail with error containing "range end must be non-negative"

  Scenario: range rejects fractional bounds instead of rounding
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(3.5)
      """
    Then running the script should fail with error containing "numeric precision loss"

  Scenario: range rejects a float beyond the int64 boundary
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(9223372036854775808.0)
      """
    Then running the script should fail with error containing "outside the supported integer range"

  Scenario: range requires numeric argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range("not a number")
      """
    Then running the script should fail with error containing "range expects a number"

  Scenario: range requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(5, 10)
      """
    Then running the script should fail with error containing "range requires exactly 1 argument"

  Scenario: range used in map
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(3) map (x) -> x * 2
      """
    When I run the script
    Then the output should be:
      """
      [0,2,4]
      """

  Scenario: range is not rewritten inside other identifiers
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      strange(1)
      """
    Then running the script should fail with error containing "undefined function: strange"

  # To function tests
  Scenario: to basic usage ascending
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(1, 5)
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3,4,5]
      """

  Scenario: to with same start and end
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(5, 5)
      """
    When I run the script
    Then the output should be:
      """
      [5]
      """

  Scenario: to descending
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(5, 1)
      """
    When I run the script
    Then the output should be:
      """
      [5,4,3,2,1]
      """

  Scenario: to preserves fractional bounds ascending
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(1.9, 3.1)
      """
    When I run the script
    Then the output should be:
      """
      [1.9,2.9]
      """

  Scenario: to preserves fractional bounds descending
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(3.1, 1.9)
      """
    When I run the script
    Then the output should be:
      """
      [3.1,2.1]
      """

  Scenario: to with negative numbers
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(-2, 2)
      """
    When I run the script
    Then the output should be:
      """
      [-2,-1,0,1,2]
      """

  Scenario: to with both negative numbers
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(-5, -1)
      """
    When I run the script
    Then the output should be:
      """
      [-5,-4,-3,-2,-1]
      """

  Scenario: to descending with negative numbers
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(2, -2)
      """
    When I run the script
    Then the output should be:
      """
      [2,1,0,-1,-2]
      """

  Scenario: to requires two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(5)
      """
    Then running the script should fail with error containing "to requires exactly 2 arguments"

  Scenario: to with non-numeric start
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to("start", 5)
      """
    Then running the script should fail with error containing "to expects start to be a number"

  Scenario: to with non-numeric end
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(1, "end")
      """
    Then running the script should fail with error containing "to expects end to be a number"

  Scenario: to used in filter
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      to(1, 10) filter (x) -> x > 5
      """
    When I run the script
    Then the output should be:
      """
      [6,7,8,9,10]
      """

  # Infix to operator tests (DataWeave syntax: 1 to 5)
  Scenario: infix to basic usage ascending
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 to 5
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3,4,5]
      """

  Scenario: infix to with parentheses
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (1 to 5)
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3,4,5]
      """

  Scenario: infix to descending
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      5 to 1
      """
    When I run the script
    Then the output should be:
      """
      [5,4,3,2,1]
      """

  Scenario: infix to preserves fractional bounds ascending
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1.9 to 3.1
      """
    When I run the script
    Then the output should be:
      """
      [1.9,2.9]
      """

  Scenario: infix to preserves fractional bounds descending
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      3.1 to 1.9
      """
    When I run the script
    Then the output should be:
      """
      [3.1,2.1]
      """

  Scenario: infix to with negative numbers
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      -2 to 2
      """
    When I run the script
    Then the output should be:
      """
      [-2,-1,0,1,2]
      """

  Scenario: infix to used in map
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (1 to 5) map (x) -> x * 2
      """
    When I run the script
    Then the output should be:
      """
      [2,4,6,8,10]
      """

  Scenario: infix to used in filter
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      (1 to 10) filter (x) -> x > 5
      """
    When I run the script
    Then the output should be:
      """
      [6,7,8,9,10]
      """

  Scenario: infix to with expressions
    Given the following script:
      """
      %im 0.1
      output application/json
      var start = 1
      var end = 5
      ---
      start to end
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3,4,5]
      """

  Scenario: infix to large range
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      1 to 100
      """
    When I run the script
    Then the output should be valid JSON
    And the output should contain "[1,2,3"
    And the output should contain "98,99,100]"

  # Zip function tests
  Scenario: zip basic usage
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([1, 2, 3], ["a", "b", "c"])
      """
    When I run the script
    Then the output should be:
      """
      [[1,"a"],[2,"b"],[3,"c"]]
      """

  Scenario: zip with different length arrays - shorter first
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([1, 2], ["a", "b", "c", "d"])
      """
    When I run the script
    Then the output should be:
      """
      [[1,"a"],[2,"b"]]
      """

  Scenario: zip with different length arrays - shorter second
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([1, 2, 3, 4], ["a", "b"])
      """
    When I run the script
    Then the output should be:
      """
      [[1,"a"],[2,"b"]]
      """

  Scenario: zip with empty first array
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([], ["a", "b", "c"])
      """
    When I run the script
    Then the output should be:
      """
      []
      """

  Scenario: zip with empty second array
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([1, 2, 3], [])
      """
    When I run the script
    Then the output should be:
      """
      []
      """

  Scenario: zip with both empty arrays
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([], [])
      """
    When I run the script
    Then the output should be:
      """
      []
      """

  Scenario: zip with mixed types
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([1, 2, 3], [true, false, null])
      """
    When I run the script
    Then the output should be:
      """
      [[1,true],[2,false],[3,null]]
      """

  Scenario: zip with objects
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([{"x": 1}, {"x": 2}], [{"y": 10}, {"y": 20}])
      """
    When I run the script
    Then the output should contain "[{\"x\":1},{\"y\":10}]"
    And the output should contain "[{\"x\":2},{\"y\":20}]"

  Scenario: zip requires two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([1, 2, 3])
      """
    Then running the script should fail with error containing "zip requires exactly 2 arguments"

  Scenario: zip with non-array first argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip("not an array", [1, 2, 3])
      """
    Then running the script should fail with error containing "zip expects first argument to be an array"

  Scenario: zip with non-array second argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([1, 2, 3], "not an array")
      """
    Then running the script should fail with error containing "zip expects second argument to be an array"

  Scenario: zip combined with map
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip([1, 2, 3], [10, 20, 30]) map (pair) -> pair[0] + pair[1]
      """
    When I run the script
    Then the output should be:
      """
      [11,22,33]
      """

  # Unzip function tests
  Scenario: unzip basic usage
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip([[1, "a"], [2, "b"], [3, "c"]])
      """
    When I run the script
    Then the output should be:
      """
      [[1,2,3],["a","b","c"]]
      """

  Scenario: unzip with single pair
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip([[5, "x"]])
      """
    When I run the script
    Then the output should be:
      """
      [[5],["x"]]
      """

  Scenario: unzip with empty array
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip([])
      """
    When I run the script
    Then the output should be:
      """
      [[],[]]
      """

  Scenario: unzip with mixed types
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip([[1, true], [2, false], [3, null]])
      """
    When I run the script
    Then the output should be:
      """
      [[1,2,3],[true,false,null]]
      """

  Scenario: unzip requires array of pairs
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip([1, 2, 3])
      """
    Then running the script should fail with error containing "unzip expects array of pairs"

  Scenario: unzip requires pairs with exactly 2 elements
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip([[1, 2, 3], [4, 5, 6]])
      """
    Then running the script should fail with error containing "unzip expects array of pairs"

  Scenario: unzip with non-pair element
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip([[1, "a"], "not a pair"])
      """
    Then running the script should fail with error containing "unzip expects array of pairs"

  Scenario: unzip requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip([1, 2], [3, 4])
      """
    Then running the script should fail with error containing "unzip requires exactly 1 argument"

  Scenario: unzip with non-array argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip("not an array")
      """
    Then running the script should fail with error containing "unzip: argument 1 expected array"

  # Zip and unzip roundtrip
  Scenario: zip and unzip roundtrip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip(zip([1, 2, 3], ["a", "b", "c"]))
      """
    When I run the script
    Then the output should be:
      """
      [[1,2,3],["a","b","c"]]
      """

  # Combined usage scenarios
  Scenario: range combined with zip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      zip(range(3), ["a", "b", "c"])
      """
    When I run the script
    Then the output should be:
      """
      [[0,"a"],[1,"b"],[2,"c"]]
      """

  Scenario: to combined with unzip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      unzip(to(1, 3) map (x) -> [x, x * 10])
      """
    When I run the script
    Then the output should be:
      """
      [[1,2,3],[10,20,30]]
      """
