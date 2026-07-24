Feature: Runtime Functions
  In order to handle errors gracefully
  As a developer
  I want to use fail, try, orElse, and orElseTry functions

  # fail() function tests

  Scenario: fail with custom message throws an exception
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      fail("Something went wrong")
      """
    When I run the application and it fails
    Then the output should contain "Something went wrong"

  Scenario: fail with default message throws an exception
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      fail()
      """
    When I run the application and it fails
    Then the output should contain "Error"

  Scenario: fail expects string argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      fail(123)
      """
    When I run the application and it fails
    Then the output should contain "fail expects a string argument"

  Scenario: fail accepts at most one argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      fail("msg1", "msg2")
      """
    When I run the application and it fails
    Then the output should contain "fail accepts at most 1 argument"

  Scenario: fail used conditionally
    Given the following input content:
      """
      %im 0.1
      output application/json
      var data = []
      ---
      if (sizeOf(data) == 0) fail("Data was empty") else data
      """
    When I run the application and it fails
    Then the output should contain "Data was empty"

  Scenario: fail not triggered when condition is false
    Given the following input content:
      """
      %im 0.1
      output application/json
      var data = [1, 2, 3]
      ---
      if (sizeOf(data) == 0) fail("Data was empty") else data
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  # try() function tests

  Scenario: try catches fail and returns error object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      try(() -> fail("Something went wrong")).success
      """
    When I run the application with this content
    Then the output should be false

  Scenario: try returns success true when expression succeeds
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      try(() -> 42).success
      """
    When I run the application with this content
    Then the output should be true

  Scenario: try returns result when expression succeeds
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      try(() -> 42).result
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: try returns error message when expression fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      try(() -> fail("Custom error")).error.message
      """
    When I run the application with this content
    Then the output should be "Custom error"

  Scenario: try returns error kind for user exceptions
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      try(() -> fail("test")).error.kind
      """
    When I run the application with this content
    Then the output should be "UserException"

  Scenario: try catches runtime errors
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      try(() -> (1 / 0)).success
      """
    When I run the application with this content
    Then the output should be false

  Scenario: try with complex expression that succeeds
    Given the following input content:
      """
      %im 0.1
      output application/json
      var items = [1, 2, 3]
      fun doubled() = items map $ * 2
      ---
      try(doubled)
      """
    When I run the application with this content
    Then the output should be:
      """
      {"result":[2,4,6],"success":true}
      """

  Scenario: try-family lambdas use their captured lexical scope
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      do {
        var offset = 5
        var captured = () -> offset
        ---
        do {
          var offset = 100
          ---
          [try(captured).result, orElse(try(() -> fail("first")), captured), orElseTry(try(() -> fail("second")), captured).result]
        }
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      [5,5,5]
      """

  Scenario: try requires exactly one argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      try()
      """
    When I run the application and it fails
    Then the output should contain "try requires exactly 1 argument"

  # orElse() function tests

  Scenario: orElse returns result when try succeeds
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> 42), "fallback")
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: orElse returns fallback when try fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> fail("error")), "fallback")
      """
    When I run the application with this content
    Then the output should be "fallback"

  Scenario: orElse with fallback function
    Given the following input content:
      """
      %im 0.1
      output application/json
      fun getDefault() = "computed default"
      ---
      orElse(try(() -> fail("error")), getDefault)
      """
    When I run the application with this content
    Then the output should be "computed default"

  Scenario: orElse does not evaluate fallback when try succeeds
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> "success"), fail("should not be called"))
      """
    When I run the application with this content
    Then the output should be "success"

  Scenario: orElse requires exactly two arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> 1))
      """
    When I run the application and it fails
    Then the output should contain "orElse requires exactly 2 arguments"

  Scenario: orElse requires TryResult as first argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElse("not a try result", "fallback")
      """
    When I run the application and it fails
    Then the output should contain "orElse: first argument must be a TryResult object"

  # orElseTry() function tests

  Scenario: orElseTry returns original TryResult when try succeeds
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> 42), () -> 99)
      """
    When I run the application with this content
    Then the output should be:
      """
      {"result":42,"success":true}
      """

  Scenario: orElseTry evaluates fallback when try fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> fail("error")), () -> 99)
      """
    When I run the application with this content
    Then the output should be:
      """
      {"result":99,"success":true}
      """

  Scenario: orElseTry captures fallback failure
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> fail("first")), () -> fail("second")).success
      """
    When I run the application with this content
    Then the output should be false

  Scenario: orElseTry chain with first success
    Given the following input content:
      """
      %im 0.1
      output application/json
      fun first() = 1
      fun second() = 2
      fun third() = 3
      ---
      orElse(orElseTry(orElseTry(try(first), second), third), 0)
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: orElseTry chain with fallback to second
    Given the following input content:
      """
      %im 0.1
      output application/json
      fun first() = fail("no first")
      fun second() = 2
      fun third() = 3
      ---
      orElse(orElseTry(orElseTry(try(first), second), third), 0)
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: orElseTry requires exactly two arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> 1))
      """
    When I run the application and it fails
    Then the output should contain "orElseTry requires exactly 2 arguments"
