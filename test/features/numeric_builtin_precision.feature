Feature: Exact numeric builtin behavior
  In order to transform large identifiers and counters safely
  As an InfoMunge user
  I want numeric builtins to preserve exact integer values

  Scenario: Numeric builtins preserve values immediately above 2^53
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [
        sum([9007199254740993, 1]),
        avg([9007199254740993, 9007199254740993]),
        max([9007199254740992, 9007199254740993]),
        abs(9007199254740993),
        mod(9007199254740993, 2)
      ]
      """
    When I run the script
    Then the output should be:
      """
      [9007199254740994,9007199254740993,9007199254740993,9007199254740993,1]
      """

  Scenario: Integer predicates classify an exact float at the int64 boundary
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [
        isInteger(9223372036854775808.0),
        isEven(9223372036854775808.0),
        isOdd(9223372036854775808.0)
      ]
      """
    When I run the script
    Then the output should be:
      """
      [true,true,false]
      """

  Scenario: Numeric builtins report integer overflow
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      sum([9223372036854775807, 1])
      """
    Then running the script should fail with error containing "integer overflow during sum"

  Scenario: Numeric builtins report unsupported precision loss
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      avg([9007199254740992, 9007199254740993])
      """
    Then running the script should fail with error containing "numeric precision loss during avg"
