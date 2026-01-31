Feature: Union Types
  In order to accept multiple types for a value
  As a developer
  I want to use union type syntax with the 'as' operator

  Scenario: Coerce string to String | Number (matches String)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" as String | Number
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: Coerce number to String | Number (matches Number)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      42 as String | Number
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Coerce null to String | Number | Null (matches Null)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null as String | Number | Null
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: String value matches String in Number | String union
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "42" as Number | String
      """
    When I run the application with this content
    Then the output should be:
      """
      "42"
      """

  Scenario: Coerce boolean fails for String | Number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      true as String | Number
      """
    When I run the application with this content
    Then the output should be:
      """
      "true"
      """

  Scenario: Union type in object value (string value matches String)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {value: "99.99" as Number | String}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"value":"99.99"}
      """

  Scenario: Union type in object value with actual number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {value: 99.99 as Number | String}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"value":99.99}
      """

  Scenario: Custom type with union base type
    Given the following input content:
      """
      %im 0.1
      type Numeric = String | Number
      output application/json
      ---
      42 as Numeric
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Union with three types
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [
        "hello" as String | Number | Boolean,
        42 as String | Number | Boolean,
        true as String | Number | Boolean
      ]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["hello",42,true]
      """
