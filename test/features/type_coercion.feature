Feature: Type Coercion with as Operator
  In order to convert values between types
  As a developer
  I want to use the 'as' operator to coerce values to specific types

  Scenario: Coerce number to String
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      42 as String
      """
    When I run the application with this content
    Then the output should be:
      """
      "42"
      """

  Scenario: Coerce string to Number (integer)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "123" as Number
      """
    When I run the application with this content
    Then the output should be:
      """
      123
      """

  Scenario: Coerce string to Number (float)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "3.14" as Number
      """
    When I run the application with this content
    Then the output should be:
      """
      3.14
      """

  Scenario: Coerce string to Boolean (true)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "true" as Boolean
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Coerce string to Boolean (false)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "false" as Boolean
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Coerce number to Boolean
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      1 as Boolean
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Coerce zero to Boolean
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      0 as Boolean
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Coerce boolean to String
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      true as String
      """
    When I run the application with this content
    Then the output should be:
      """
      "true"
      """

  Scenario: Coerce boolean to Number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      true as Number
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: Use custom type for coercion
    Given the following input content:
      """
      %im 0.1
      type Amount = Number
      output application/json
      ---
      "42" as Amount
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Use custom type aliasing String
    Given the following input content:
      """
      %im 0.1
      type Name = String
      output application/json
      ---
      123 as Name
      """
    When I run the application with this content
    Then the output should be:
      """
      "123"
      """

  Scenario: Coerce in object value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {price: "99.99" as Number}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"price":99.99}
      """

  Scenario: Chain coercions
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      (42 as String) as Number
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Coerce null to String
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null as String
      """
    When I run the application with this content
    Then the output should be:
      """
      "null"
      """

  Scenario: Coerce null to Number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null as Number
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: Coerce null to Boolean
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null as Boolean
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: Invalid string to Number fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "42abc" as Number
      """
    Then the application should fail with error containing "cannot coerce"

  Scenario: Invalid string to Boolean fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "maybe" as Boolean
      """
    Then the application should fail with error containing "cannot coerce"

  Scenario: Unknown type fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      42 as UnknownType
      """
    Then the application should fail with error containing "unknown type: UnknownType"

  Scenario: Coerce with whitespace in string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      " 42 " as Number
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """
