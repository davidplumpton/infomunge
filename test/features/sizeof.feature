Feature: Size Of Function
  In order to determine the length or count of collections
  As a developer
  I want to use the sizeOf function to get the size of strings, arrays, and objects

  Scenario: Get length of a string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf("hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: Get length of an empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf("")
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: Count Unicode characters instead of UTF-8 bytes
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [sizeOf("é"), sizeOf("é🙂"), "é"[0 to sizeOf("é") - 1]]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,"é"]
      """

  Scenario: Get length of an array literal
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf([1, 2, 3, 4, 5])
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: Get length of an empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf([])
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: Get key count of an object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf({"a": 1, "b": 2, "c": 3})
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: Get length of a variable string
    Given the following input content:
      """
      %im 0.1
      var message = "test"
      output application/json
      ---
      sizeOf(message)
      """
    When I run the application with this content
    Then the output should be:
      """
      4
      """

  Scenario: Get length of a variable array
    Given the following input content:
      """
      %im 0.1
      var items = ["apple", "banana", "cherry"]
      output application/json
      ---
      sizeOf(items)
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: Get length of a variable object
    Given the following input content:
      """
      %im 0.1
      var config = {"host": "localhost", "port": 8080}
      output application/json
      ---
      sizeOf(config)
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: Get size of nil value as null
    Given the following input content:
      """
      %im 0.1
      var empty = nil
      output application/json
      ---
      sizeOf(empty)
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: Use sizeOf in an arithmetic expression
    Given the following input content:
      """
      %im 0.1
      var arr = [1, 2, 3]
      output application/json
      ---
      sizeOf(arr) + 2
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: Use sizeOf in a conditional expression
    Given the following input content:
      """
      %im 0.1
      var list = ["a", "b"]
      output application/json
      ---
      sizeOf(list) == 2
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Nested sizeOf calls
    Given the following input content:
      """
      %im 0.1
      var data = [["a", "b"], ["c"], []]
      output application/json
      ---
      sizeOf(data)
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: sizeOf with no arguments should fail
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf()
      """
    When I run the application and it fails
    Then the error should contain "requires exactly 1 argument"

  Scenario: sizeOf with too many arguments should fail
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf("a", "b")
      """
    When I run the application and it fails
    Then the error should contain "requires exactly 1 argument"

  Scenario: sizeOf on integer should fail
    Given the following input content:
      """
      %im 0.1
      var num = 42
      output application/json
      ---
      sizeOf(num)
      """
    When I run the application and it fails
    Then the error should contain "unsupported type"

  Scenario: sizeOf on boolean should fail
    Given the following input content:
      """
      %im 0.1
      var flag = true
      output application/json
      ---
      sizeOf(flag)
      """
    When I run the application and it fails
    Then the error should contain "unsupported type"
