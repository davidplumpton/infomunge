Feature: Optional Types
  In order to accept a value or null
  As a developer
  I want to use optional type syntax (Type?) with the 'as' operator

  Scenario: Coerce string to String? (matches String)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello" as String?
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: Coerce null to String? (matches Null)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null as String?
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: Coerce number to Number? (matches Number)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      42 as Number?
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Coerce null to Number? (matches Null)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      null as Number?
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: Optional type in object value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {name: "Alice" as String?, age: null as Number?}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"Alice","age":null}
      """

  Scenario: Custom type with optional base type
    Given the following input content:
      """
      %im 0.1
      type OptionalNumber = Number?
      output application/json
      ---
      [42 as OptionalNumber, null as OptionalNumber]
      """
    When I run the application with this content
    Then the output should be:
      """
      [42,null]
      """

  Scenario: Coerce number to String? (coerces to String)
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      42 as String?
      """
    When I run the application with this content
    Then the output should be:
      """
      "42"
      """

  Scenario: Optional with custom type
    Given the following input content:
      """
      %im 0.1
      type Currency = String { format: "##.00" }
      output application/json
      ---
      [1234.5 as Currency?, null as Currency?]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["1234.50",null]
      """
