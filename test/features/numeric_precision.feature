Feature: Exact numeric behavior
  InfoMunge must not silently truncate, overflow, or collapse distinct numbers.

  Scenario: Integer division preserves a fractional result
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      5 / 2
      """
    When I run the application with this content
    Then the output should be:
      """
      2.5
      """

  Scenario: Integer overflow produces an explicit error
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      9223372036854775807 + 1
      """
    When I run the application and it fails
    Then the error should contain "integer overflow during addition"

  Scenario: Minimum signed integer literal is accepted
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      -9223372036854775808
      """
    When I run the application with this content
    Then the output should be:
      """
      -9223372036854775808
      """

  Scenario: Equality preserves integers above float precision
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      9007199254740993 == 9007199254740992.0
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Coercion equality preserves integer strings above float precision
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "9007199254740993" ~= 9007199254740992
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Ordering preserves integers above float precision
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      9007199254740993 > 9007199254740992.0
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Decimal arithmetic preserves an exact large integer result
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      9007199254740993 + 0.0
      """
    When I run the application with this content
    Then the output should be:
      """
      9007199254740993
      """
