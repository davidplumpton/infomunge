Feature: String Interpolation
  In order to build dynamic strings with embedded expressions
  As a developer
  I want to use $(expression) syntax within strings

  Scenario: Simple variable interpolation
    Given the following input content:
      """
      %im 0.1
      output application/json
      var name = "World"
      ---
      "Hello $(name)!"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello World!"
      """

  Scenario: Multiple interpolations
    Given the following input content:
      """
      %im 0.1
      output application/json
      var first = "John"
      var last = "Doe"
      ---
      "$(first) $(last)"
      """
    When I run the application with this content
    Then the output should be:
      """
      "John Doe"
      """

  Scenario: Interpolation with numeric expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      var x = 5
      var y = 3
      ---
      "$(x) + $(y) = $(x + y)"
      """
    When I run the application with this content
    Then the output should be:
      """
      "5 + 3 = 8"
      """

  Scenario: Interpolation with object field access
    Given the following input content:
      """
      %im 0.1
      output application/json
      var user = {name: "Alice", age: 30}
      ---
      "$(user.name) is $(user.age)"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Alice is 30"
      """

  Scenario: String without interpolation unchanged
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "No interpolation here"
      """
    When I run the application with this content
    Then the output should be:
      """
      "No interpolation here"
      """

  Scenario: Interpolation at start of string
    Given the following input content:
      """
      %im 0.1
      output application/json
      var greeting = "Hello"
      ---
      "$(greeting) World"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello World"
      """

  Scenario: Interpolation at end of string
    Given the following input content:
      """
      %im 0.1
      output application/json
      var name = "Alice"
      ---
      "Hello $(name)"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello Alice"
      """

  Scenario: Only interpolation (no literal text)
    Given the following input content:
      """
      %im 0.1
      output application/json
      var msg = "Hello World"
      ---
      "$(msg)"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello World"
      """

  Scenario: Interpolation with function call
    Given the following input content:
      """
      %im 0.1
      output application/json
      var items = [1, 2, 3]
      ---
      "Array has $(sizeOf(items)) items"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Array has 3 items"
      """

  Scenario: Interpolation with comparison
    Given the following input content:
      """
      %im 0.1
      output application/json
      var count = 5
      ---
      "Is five: $(count == 5)"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Is five: true"
      """

  Scenario: Interpolation with nested parentheses
    Given the following input content:
      """
      %im 0.1
      output application/json
      var a = 2
      var b = 3
      ---
      "Result: $((a + b) * 2)"
      """
    When I run the application with this content
    Then the output should be:
      """
      "Result: 10"
      """

  Scenario: Integer to string coercion
    Given the following input content:
      """
      %im 0.1
      output application/json
      var count = 42
      ---
      "Count: " + count
      """
    When I run the application with this content
    Then the output should be:
      """
      "Count: 42"
      """

  Scenario: Boolean to string coercion
    Given the following input content:
      """
      %im 0.1
      output application/json
      var active = true
      ---
      "Active: " + active
      """
    When I run the application with this content
    Then the output should be:
      """
      "Active: true"
      """

  Scenario: Float to string coercion
    Given the following input content:
      """
      %im 0.1
      output application/json
      var price = 19.99
      ---
      "Price: $" + price
      """
    When I run the application with this content
    Then the output should be:
      """
      "Price: $19.99"
      """
