Feature: Type Checking Functions
  In order to verify type predicates
  As a developer
  I want to use type checking functions: isBlank, isDecimal, isInteger, isEven, isOdd

  # isBlank tests
  Scenario: isBlank with empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank("")
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isBlank with whitespace string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank("   ")
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isBlank with null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank(null)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isBlank with non-empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank("hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isBlank with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank([])
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isBlank with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank({})
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isBlank with zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank(0)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isBlank with false boolean
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank(false)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  # isDecimal tests
  Scenario: isDecimal with decimal number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isDecimal(1.5)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isDecimal with negative decimal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isDecimal(-3.14)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isDecimal with zero point zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isDecimal(0.0)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isDecimal with integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isDecimal(5)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isDecimal with string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isDecimal("3.14")
      """
    When I run the application and it fails
    Then the output should contain "isDecimal expects a number"

  Scenario: isDecimal with null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isDecimal(null)
      """
    When I run the application and it fails
    Then the output should contain "isDecimal expects a number"

  # isInteger tests
  Scenario: isInteger with positive integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isInteger(42)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isInteger with negative integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isInteger(-7)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isInteger with zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isInteger(0)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isInteger with decimal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isInteger(3.14)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isInteger with string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isInteger("42")
      """
    When I run the application and it fails
    Then the output should contain "isInteger expects a number"

  # isEven tests
  Scenario: isEven with even positive integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEven(4)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isEven with even negative integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEven(-6)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isEven with zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEven(0)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isEven with odd positive integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEven(5)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isEven with odd negative integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEven(-3)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isEven with decimal fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEven(3.5)
      """
    When I run the application and it fails
    Then the output should contain "isEven expects an integer"

  Scenario: isEven with string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isEven("4")
      """
    When I run the application and it fails
    Then the output should contain "isEven expects an integer"

  # isOdd tests
  Scenario: isOdd with odd positive integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isOdd(7)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isOdd with odd negative integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isOdd(-5)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isOdd with even positive integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isOdd(8)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isOdd with even negative integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isOdd(-4)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isOdd with zero
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isOdd(0)
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: isOdd with decimal fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isOdd(2.5)
      """
    When I run the application and it fails
    Then the output should contain "isOdd expects an integer"

  Scenario: isOdd with string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isOdd("7")
      """
    When I run the application and it fails
    Then the output should contain "isOdd expects an integer"

  # Combination scenarios
  Scenario: isBlank in conditional
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (isBlank("")) "empty" else "not empty"
      """
    When I run the application with this content
    Then the output should be:
      """
      "empty"
      """

  Scenario: isInteger and isEven on collection
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] filter (x) -> isEven(x)
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,4]
      """

  Scenario: isOdd on collection
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4, 5] filter (x) -> isOdd(x)
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,3,5]
      """

  Scenario: isDecimal with scientific notation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isDecimal(1.23e-4)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isInteger with large number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isInteger(999999999)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: isBlank with string containing newlines
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      isBlank("\n\t  ")
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  # is operator tests
  Scenario: is operator with String type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" is String
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: is operator with Number type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      42 is Number
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: is operator with Boolean type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      true is Boolean
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: is operator with Array type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] is Array
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: is operator with Object type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {a: 1} is Object
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: is operator with Null type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null is Null
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: is operator returns false for wrong type
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" is Number
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: is operator with float Number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      3.14 is Number
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: is operator in conditional
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if ("test" is String) "yes" else "no"
      """
    When I run the application with this content
    Then the output should be:
      """
      "yes"
      """
