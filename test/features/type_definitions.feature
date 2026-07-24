Feature: Type Definitions
  In order to define reusable types for coercion and validation
  As a developer
  I want to declare custom types in the header using the type keyword

  Scenario: Simple type declaration is parsed
    Given the following input content:
      """
      %im 0.1
      type Currency = String
      output application/json
      ---
      "hello"
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: Type declaration with Number base type
    Given the following input content:
      """
      %im 0.1
      type Amount = Number
      output application/json
      ---
      42
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Multiple type declarations
    Given the following input content:
      """
      %im 0.1
      type Name = String
      type Age = Number
      type Active = Boolean
      output application/json
      ---
      {name: "Alice", age: 30, active: true}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"Alice","age":30,"active":true}
      """

  Scenario: Type declaration with format property
    Given the following input content:
      """
      %im 0.1
      type Currency = String { format: "##.00" }
      output application/json
      ---
      42.5 as Currency
      """
    When I run the application with this content
    Then the output should be:
      """
      "42.50"
      """

  Scenario: Type declaration with thousands separator format
    Given the following input content:
      """
      %im 0.1
      type Money = String { format: "###,###.00" }
      output application/json
      ---
      1234567.89 as Money
      """
    When I run the application with this content
    Then the output should be:
      """
      "1,234,567.89"
      """

  Scenario: Type declaration with multiple properties
    Given the following input content:
      """
      %im 0.1
      type Price = String { format: "##.00", prefix: "$" }
      output application/json
      ---
      99.9 as Price
      """
    When I run the application with this content
    Then the output should be:
      """
      "99.90"
      """

  Scenario: Type declaration with empty properties block
    Given the following input content:
      """
      %im 0.1
      type Plain = String { }
      output application/json
      ---
      "hello" as Plain
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """
