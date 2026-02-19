Feature: Math builtins runner coverage
  In order to measure cucumber coverage for math builtin internals
  As a maintainer
  I want math builtins in builtins_math.go exercised in-process

  Scenario: Math builtin success paths execute via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        ceilValue: ceil(3.2),
        floorValue: floor(3.8),
        roundValue: round(3.5),
        sqrtValue: sqrt(16),
        maxValue: max([1, 5, 3]),
        minValue: min([1, 5, 3]),
        avgValue: avg([2, 4, 6]),
        decimalCheck: isDecimal(1.5),
        integerCheck: isInteger(4.0),
        evenCheck: isEven(8),
        oddCheck: isOdd(7),
        randomValue: random(),
        randomIntValue: randomInt(10)
      }
      """
    When I run the script
    Then the output should be valid JSON with 13 keys

  Scenario: Math builtin error paths execute via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ceil()
      """
    Then running the script should fail with error containing "ceil function requires exactly 1 argument"

  Scenario: randomInt validates max argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      randomInt(0)
      """
    Then running the script should fail with error containing "randomInt max must be greater than 0"

  Scenario: math builtin type validation errors are reported consistently
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      floor("x")
      """
    Then running the script should fail with error containing "floor: cannot convert string to number"
