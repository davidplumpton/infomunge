Feature: Utility Functions
  In order to perform utility operations
  As a developer
  I want to use the with, xsiType, mod, and evaluateCompatibilityFlag functions

  # Mod function tests
  Scenario: mod basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(10, 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: mod with zero remainder
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(10, 5)
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: mod with floating point numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(10.5, 3)
      """
    When I run the application with this content
    Then the output should contain "1.5"

  Scenario: mod with negative dividend
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(-10, 3)
      """
    When I run the application with this content
    Then the output should contain "-1"

  Scenario: mod with negative divisor
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(10, -3)
      """
    When I run the application with this content
    Then the output should contain "1"

  Scenario: mod with both negative
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(-10, -3)
      """
    When I run the application with this content
    Then the output should contain "-1"

  Scenario: mod with small divisor
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(1, 0.3)
      """
    When I run the application with this content
    Then the output should be valid JSON with number close to 0.1

  Scenario: mod requires exactly two arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(10)
      """
    When I run the application and it fails
    Then the output should contain "requires exactly 2 arguments"

  Scenario: mod with non-numeric dividend fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod("ten", 3)
      """
    When I run the application and it fails
    Then the output should contain "expects dividend to be a number"

  Scenario: mod with non-numeric divisor fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(10, "three")
      """
    When I run the application and it fails
    Then the output should contain "expects divisor to be a number"

   Scenario: mod with zero divisor fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       mod(10, 0)
       """
     When I run the application and it fails
     Then the output should contain "division by zero"

   Scenario: mod operator basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       17 mod 5
       """
     When I run the application with this content
     Then the output should be:
       """
       2
       """

   Scenario: mod operator with zero remainder
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       10 mod 5
       """
     When I run the application with this content
     Then the output should be:
       """
       0
       """

   Scenario: mod operator with floating point
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       10.5 mod 3
       """
     When I run the application with this content
     Then the output should contain "1.5"

   Scenario: mod operator with negative numbers
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       -10 mod 3
       """
     When I run the application with this content
     Then the output should be:
       """
       -1
       """

   # With function tests
  Scenario: with basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      with("original", "replacement")
      """
    When I run the application with this content
    Then the output should be:
      """
      "replacement"
      """

  Scenario: with returns second argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      with(null, "new value")
      """
    When I run the application with this content
    Then the output should be:
      """
      "new value"
      """

  Scenario: with with numeric values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      with(5, 10)
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  Scenario: with with objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      with({"a": 1}, {"b": 2})
      """
    When I run the application with this content
    Then the output should contain "b"
    And the output should contain "2"

  Scenario: with with arrays
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      with([1, 2], [3, 4])
      """
    When I run the application with this content
    Then the output should be:
      """
      [3,4]
      """

  Scenario: with with boolean
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      with(false, true)
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: with ignores first argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      with({"ignored": true}, {"used": true})
      """
    When I run the application with this content
    Then the output should contain "used"
    And the output should not contain "ignored"

  Scenario: with requires exactly two arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      with("only one")
      """
    When I run the application and it fails
    Then the output should contain "requires exactly 2 arguments"

  # XsiType function tests
  Scenario: xsiType basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      xsiType("string")
      """
    When I run the application with this content
    Then the output should contain "xsi:type"
    And the output should contain "string"

  Scenario: xsiType with namespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      xsiType("xs:string")
      """
    When I run the application with this content
    Then the output should contain "xs:string"

  Scenario: xsiType with complex type name
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      xsiType("CustomType")
      """
    When I run the application with this content
    Then the output should be:
      """
      {"xsi:type":"CustomType"}
      """

  Scenario: xsiType with integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      xsiType("int")
      """
    When I run the application with this content
    Then the output should contain "int"

  Scenario: xsiType returns object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      typeOf(xsiType("myType"))
      """
    When I run the application with this content
    Then the output should be:
      """
      "Object"
      """

  Scenario: xsiType creates single key object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(xsiType("type"))
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: xsiType requires exactly one argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      xsiType()
      """
    When I run the application and it fails
    Then the output should contain "requires exactly 1 argument"

  Scenario: xsiType with non-string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      xsiType(123)
      """
    When I run the application and it fails
    Then the output should contain "expects typeName to be a string"

  # EvaluateCompatibilityFlag function tests
  Scenario: evaluateCompatibilityFlag with allowUndefinedProperties
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("allowUndefinedProperties")
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: evaluateCompatibilityFlag with allowNullArithmetic
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("allowNullArithmetic")
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: evaluateCompatibilityFlag with allowImplicitConversion
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("allowImplicitConversion")
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: evaluateCompatibilityFlag with strictTypeChecking
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("strictTypeChecking")
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: evaluateCompatibilityFlag with allowDynamicProperties
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("allowDynamicProperties")
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: evaluateCompatibilityFlag with unknown flag returns null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("unknownFlag")
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: evaluateCompatibilityFlag requires exactly one argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag()
      """
    When I run the application and it fails
    Then the output should contain "requires exactly 1 argument"

  Scenario: evaluateCompatibilityFlag with non-string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag(123)
      """
    When I run the application and it fails
    Then the output should contain "expects flagName to be a string"

  # Combined scenarios
  Scenario: use mod to check if even
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      mod(4, 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: use with for conditional replacement
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      if (true) with("old", "new") else "default"
      """
    When I run the application with this content
    Then the output should be:
      """
      "new"
      """

  Scenario: combine xsiType with object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      xsiType("MyType") ++ {"value": 42}
      """
    When I run the application with this content
    Then the output should contain "xsi:type"
    And the output should contain "value"
    And the output should contain "42"

  Scenario: check multiple compatibility flags
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [
        evaluateCompatibilityFlag("allowUndefinedProperties"),
        evaluateCompatibilityFlag("strictTypeChecking")
      ]
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,false]
      """

  Scenario: mod in array transformation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4] map (x) -> mod(x, 2)
      """
     When I run the application with this content
     Then the output should be:
       """
       [1,0,1,0]
       """

   Scenario: indexOf with array finds element
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       indexOf([10, 20, 30], 20)
       """
     When I run the application with this content
     Then the output should be:
       """
       1
       """

   Scenario: indexOf with array element not found
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       indexOf([10, 20, 30], 40)
       """
     When I run the application with this content
     Then the output should be:
       """
       -1
       """

   Scenario: indexOf with array first element
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       indexOf([10, 20, 30], 10)
       """
     When I run the application with this content
     Then the output should be:
       """
       0
       """

   Scenario: indexOf with string substring
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       indexOf("hello world", "world")
       """
     When I run the application with this content
     Then the output should be:
       """
       6
       """

   Scenario: indexOf with string substring not found
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       indexOf("hello world", "notfound")
       """
     When I run the application with this content
     Then the output should be:
       """
       -1
       """

   Scenario: indexOf with empty string
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       indexOf("", "a")
       """
     When I run the application with this content
     Then the output should be:
       """
       -1
       """

   Scenario: indexOf with empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       indexOf([], 10)
       """
     When I run the application with this content
     Then the output should be:
       """
       -1
       """

   Scenario: indexOf requires exactly two arguments
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       indexOf([1, 2, 3])
       """
     When I run the application and it fails
     Then the output should contain "requires exactly 2 arguments"
