Feature: Math Functions
  In order to perform mathematical operations
  As a developer
  I want to use the math functions: floor, round, ceil, abs, pow, sqrt, min, max, sum, avg, mod

  # floor tests
  Scenario: floor basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      floor(3.7)
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: floor negative number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      floor(-3.2)
      """
    When I run the application with this content
    Then the output should be:
      """
      -4
      """

  Scenario: floor integer returns same value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      floor(5)
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: floor zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      floor(0)
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  # round tests
  Scenario: round up when >= 0.5
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      round(3.5)
      """
    When I run the application with this content
    Then the output should be:
      """
      4
      """

  Scenario: round down when < 0.5
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      round(3.4)
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: round negative number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      round(-3.5)
      """
    When I run the application with this content
    Then the output should be:
      """
      -4
      """

  Scenario: round integer returns same value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      round(7)
      """
    When I run the application with this content
    Then the output should be:
      """
      7
      """

  # ceil tests
  Scenario: ceil basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ceil(3.2)
      """
    When I run the application with this content
    Then the output should be:
      """
      4
      """

  Scenario: ceil negative number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ceil(-3.7)
      """
    When I run the application with this content
    Then the output should be:
      """
      -3
      """

  Scenario: ceil integer returns same value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ceil(5)
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  # abs tests
  Scenario: abs of positive number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      abs(5)
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: abs of negative number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      abs(-5)
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: abs of zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      abs(0)
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: abs of negative float
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      abs(-3.14)
      """
    When I run the application with this content
    Then the output should be valid JSON with number close to 3.14

  # pow tests
  Scenario: pow basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      pow(2, 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      8
      """

  Scenario: pow with zero exponent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      pow(5, 0)
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: pow with negative exponent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      pow(2, -2)
      """
    When I run the application with this content
    Then the output should be valid JSON with number close to 0.25

  Scenario: pow with fractional exponent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      pow(4, 0.5)
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: pow large exponent
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      pow(2, 10)
      """
    When I run the application with this content
    Then the output should be:
      """
      1024
      """

  Scenario: pow coerces boolean inputs to numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      pow(true, 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  # sqrt tests
  Scenario: sqrt basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sqrt(16)
      """
    When I run the application with this content
    Then the output should be:
      """
      4
      """

  Scenario: sqrt of zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sqrt(0)
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: sqrt of non-perfect square
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sqrt(2)
      """
    When I run the application with this content
    Then the output should be valid JSON with number close to 1.4142135623730951

  Scenario: sqrt of negative fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sqrt(-4)
      """
    When I run the application and it fails
    Then the output should contain "sqrt: cannot take square root of negative number"

  # min tests
  Scenario: min of multiple arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      min(5, 3, 8, 1, 9)
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: min of two arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      min(10, 5)
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: min of array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      min([5, 3, 8, 1, 9])
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: min with negative numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      min(-5, 3, -8, 1)
      """
    When I run the application with this content
    Then the output should be:
      """
      -8
      """

  Scenario: min of empty array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      min([])
      """
    When I run the application and it fails
    Then the output should contain "min function requires non-empty array"

  # max tests
  Scenario: max of multiple arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      max(5, 3, 8, 1, 9)
      """
    When I run the application with this content
    Then the output should be:
      """
      9
      """

  Scenario: max of two arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      max(10, 5)
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  Scenario: max of array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      max([5, 3, 8, 1, 9])
      """
    When I run the application with this content
    Then the output should be:
      """
      9
      """

  Scenario: max with negative numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      max(-5, -3, -8, -1)
      """
    When I run the application with this content
    Then the output should be:
      """
      -1
      """

  Scenario: max of empty array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      max([])
      """
    When I run the application and it fails
    Then the output should contain "max function requires non-empty array"

  # sum tests (basic usage in aggregation_functions.feature)
  Scenario: sum of empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sum([])
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: sum with negative numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sum([10, -3, 5, -2])
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  # avg tests (basic usage in aggregation_functions.feature)
  Scenario: avg of empty array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      avg([])
      """
    When I run the application and it fails
    Then the output should contain "avg: cannot calculate average of empty array"

  # mod tests (basic usage in utility_functions.feature)
  Scenario: mod with no remainder
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(12, 4)
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: mod division by zero fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(10, 0)
      """
    When I run the application and it fails
    Then the output should contain "mod: division by zero"

  # Combination scenarios
  Scenario: nested math functions
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      floor(sqrt(50))
      """
    When I run the application with this content
    Then the output should be:
      """
      7
      """

  Scenario: abs and min combined
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      min(abs(-5), abs(3), abs(-2))
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: sum and avg on same data
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "total": sum([10, 20, 30]),
        "average": avg([10, 20, 30])
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"total":60,"average":20}
      """

  Scenario: pow and sqrt inverse relationship
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sqrt(pow(7, 2))
      """
    When I run the application with this content
    Then the output should be:
      """
      7
      """
