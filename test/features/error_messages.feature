Feature: Error Messages
  In order to understand what went wrong
  As a developer
  I want to see clear and accurate error messages

  Scenario: Missing script file reports a clear read error
    When I run the application with "missing.im" and it fails
    Then the stderr should contain "Error: IOError: error reading script file: missing.im"
    And the output should not contain "open missing.im"

  Scenario: Unknown CLI flag exits through normal error handling
    When I run the application with arguments "--bogus" and it fails
    Then the exit status should be 1
    And the stderr should contain "flag provided but not defined: -bogus"

  Scenario: Extra positional script arguments are rejected
    When I run the application with arguments "first-script second-script" and it fails
    Then the exit status should be 1
    And the stderr should contain "Error: ValidationError: expected at most one inline script argument"

  Scenario: Script file and inline script cannot be combined
    When I run the application with arguments "-f missing.im inline-script" and it fails
    Then the exit status should be 1
    And the stderr should contain "Error: ValidationError: cannot combine -f with an inline script argument"
    And the output should not contain "error reading script file"

  Scenario: Division by zero with simple expression
    Given the following input content:
      """
      %im 0.1
      ---
      10 / 0
      """
    When I run the application and it fails
    Then the error should contain "division by zero"

  Scenario: Division by zero with expression
    Given the following input content:
      """
      %im 0.1
      ---
      100 / (5 - 5)
      """
    When I run the application and it fails
    Then the error should contain "division by zero"

  Scenario: Logical AND with non-boolean values
    Given the following input content:
      """
      %im 0.1
      ---
      1 and true
      """
    When I run the application and it fails
    Then the error should contain "logical AND requires booleans"

  Scenario: Logical OR with non-boolean values
    Given the following input content:
      """
      %im 0.1
      ---
      false or 1
      """
    When I run the application and it fails
    Then the error should contain "logical OR requires booleans"

  Scenario: Cannot negate non-numeric value
    Given the following input content:
      """
      %im 0.1
      ---
      -"hello"
      """
    When I run the application and it fails
    Then the error should contain "cannot negate"

  Scenario: Cannot apply NOT to non-boolean value
    Given the following input content:
      """
      %im 0.1
      ---
      not 42
      """
    When I run the application and it fails
    Then the error should contain "cannot apply NOT to"

  Scenario: Arithmetic operation with incompatible types
    Given the following input content:
      """
      %im 0.1
      ---
      1 - "two"
      """
    When I run the application and it fails
    Then the error should contain "cannot subtract"

  Scenario: Comparison with incompatible types
    Given the following input content:
      """
      %im 0.1
      ---
      1 < "two"
      """
    When I run the application and it fails
    Then the error should contain "cannot compare"

  Scenario: Array index out of bounds
    Given the following input content:
      """
      %im 0.1
      ---
      [1, 2, 3][5]
      """
    When I run the application and it fails
    Then the error should contain "array index out of bounds"

  Scenario: Array index with non-integer
    Given the following input content:
      """
      %im 0.1
      ---
      [1, 2, 3][3.14]
      """
    When I run the application and it fails
    Then the error should contain "array index must be an integer or string"

  Scenario: Object index out of bounds
    Given the following input content:
      """
      %im 0.1
      ---
      {a: 1, b: 2}[5]
      """
    When I run the application and it fails
    Then the error should contain "object index out of bounds"

  Scenario: Negative object index out of bounds
    Given the following input content:
      """
      %im 0.1
      ---
      {a: 1, b: 2}[-3]
      """
    When I run the application and it fails
    Then the error should contain "object index out of bounds: -3 (object has 2 keys)"

  Scenario: Map key with invalid type
    Given the following input content:
      """
      %im 0.1
      ---
      {a: 1, b: 2}[3.14]
      """
    When I run the application and it fails
    Then the error should contain "map key must be a string or int"

  Scenario: Can index into string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello"[0]
      """
    When I run the application with this content
    Then the output should be:
      """
      "h"
      """

  Scenario: String field access requires an integer index
    Given the following input content:
      """
      %im 0.1
      ---
      "hello".name
      """
    When I run the application and it fails
    Then the error should contain "string index must be an integer"

  Scenario: Presence selector
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {"user": {"name": "Alice"}}.user?
      """
    When I run the application with this content
    Then the output should be:
      """
      true
      """

  Scenario: Cannot index into unsupported type
    Given the following input content:
      """
      %im 0.1
      ---
      true[0]
      """
    When I run the application and it fails
    Then the error should contain "cannot index into"

  Scenario: Function call with wrong argument count
    Given the following input content:
      """
      %im 0.1
      ---
      sqrt(4, 9)
      """
    When I run the application and it fails
    Then the error should contain "requires exactly 1 argument"

  Scenario: Function call with no arguments when required
    Given the following input content:
      """
      %im 0.1
      ---
      ceil()
      """
    When I run the application and it fails
    Then the error should contain "requires exactly 1 argument"

  Scenario: Math function with non-numeric argument
    Given the following input content:
      """
      %im 0.1
      ---
      sqrt("hello")
      """
    When I run the application and it fails
    Then the error should contain "sqrt"

  Scenario: Sqrt of negative number
    Given the following input content:
      """
      %im 0.1
      ---
      sqrt(-9)
      """
    When I run the application and it fails
    Then the error should contain "cannot take square root of negative number"

  Scenario: If condition is not boolean
    Given the following input content:
      """
      %im 0.1
      ---
      if(42) "yes" else "no"
      """
    When I run the application and it fails
    Then the error should contain "condition must be a boolean"

  Scenario: SizeOf with unsupported type
    Given the following input content:
      """
      %im 0.1
      ---
      sizeOf(true)
      """
    When I run the application and it fails
    Then the error should contain "sizeOf: unsupported type"

  Scenario: Join with non-array first argument
    Given the following input content:
      """
      %im 0.1
      ---
      join(",", "hello")
      """
    When I run the application and it fails
    Then the error should contain "join: argument 1 expected array"

  Scenario: Join with non-string separator
    Given the following input content:
      """
      %im 0.1
      ---
      join(",", ["a", "b"], 123)
      """
    When I run the application and it fails
    Then the error should contain "requires exactly 2 argument"

  Scenario: Flatten with non-array argument
    Given the following input content:
      """
      %im 0.1
      ---
      flatten("hello")
      """
    When I run the application and it fails
    Then the error should contain "flatten: argument 1 expected array"

  Scenario: Merge with no arguments
    Given the following input content:
      """
      %im 0.1
      ---
      merge()
      """
    When I run the application and it fails
    Then the error should contain "merge requires at least 1 argument"

  Scenario: Merge with non-object argument
    Given the following input content:
      """
      %im 0.1
      ---
      merge({a: 1}, [1, 2, 3])
      """
    When I run the application and it fails
    Then the error should contain "merge expects objects"

  Scenario: Read function with missing arguments
    Given the following input content:
      """
      %im 0.1
      ---
      read('{"a": 1}')
      """
    When I run the application and it fails
    Then the error should contain "requires at least 2 arguments"

  Scenario: Read function with non-string mimeType
    Given the following input content:
      """
      %im 0.1
      ---
      read('{"a": 1}', 123)
      """
    When I run the application and it fails
    Then the error should contain "arguments must be strings"

  Scenario: Write function with wrong argument count
    Given the following input content:
      """
      %im 0.1
      ---
      write({a: 1})
      """
    When I run the application and it fails
    Then the error should contain "requires exactly 2 arguments"

  Scenario: Write function with non-string mimeType
    Given the following input content:
      """
      %im 0.1
      ---
      write({a: 1}, 123)
      """
    When I run the application and it fails
    Then the error should contain "expects mimeType to be a string"

  Scenario: Sort with unsupported array element type
    Given the following input content:
      """
      %im 0.1
      ---
      sort([{a: 1}, {b: 2}])
      """
    When I run the application and it fails
    Then the error should contain "sort only supports arrays of numbers or strings"

  Scenario: KeysOf with non-object argument
    Given the following input content:
      """
      %im 0.1
      ---
      keysOf([1, 2, 3])
      """
    When I run the application and it fails
    Then the error should contain "keysOf: argument 1 expected object"

  Scenario: EntriesOf with non-object argument
    Given the following input content:
      """
      %im 0.1
      ---
      entriesOf([1, 2, 3])
      """
    When I run the application and it fails
    Then the error should contain "entriesOf: argument 1 expected object"

  Scenario: Find on string with non-string search value
    Given the following input content:
      """
      %im 0.1
      ---
      find("hello world", 42)
      """
    When I run the application and it fails
    Then the error should contain "find on string expects search value to be a string"

  Scenario: Find with wrong argument count
    Given the following input content:
      """
      %im 0.1
      ---
      find([1, 2, 3])
      """
    When I run the application and it fails
    Then the error should contain "find requires 2 or 3 arguments"

  Scenario: Assert with wrong argument count
    Given the following input content:
      """
      %im 0.1
      ---
      assert(42)
      """
    When I run the application and it fails
    Then the error should contain "assert expects 2 or 3 arguments"

  Scenario: Assert with non-matcher second argument
    Given the following input content:
      """
      %im 0.1
      ---
      assert(42, "not a matcher")
      """
    When I run the application and it fails
    Then the error should contain "assert expects second argument to be a matcher"

  Scenario: Assert with non-string message
    Given the following input content:
      """
      %im 0.1
      ---
      assert(42, equalTo(42), 123)
      """
    When I run the application and it fails
    Then the error should contain "assert expects third argument (message) to be a string"

  Scenario: Update operator with non-object left side
    Given the following input content:
      """
      %im 0.1
      ---
      [1, 2, 3] ~ {a: 4}
      """
    When I run the application and it fails
    Then the error should contain "update operator (~) expects an object on the left"

  Scenario: Update operator with non-object right side
    Given the following input content:
      """
      %im 0.1
      ---
      {a: 1} ~ [1, 2, 3]
      """
    When I run the application and it fails
    Then the error should contain "update operator (~) expects an object on the right"

  Scenario: Mod function with wrong argument count
    Given the following input content:
      """
      %im 0.1
      ---
      mod(10)
      """
    When I run the application and it fails
    Then the error should contain "requires exactly 2 arguments"

  Scenario: Mod function with zero divisor
    Given the following input content:
      """
      %im 0.1
      ---
      mod(10, 0)
      """
    When I run the application and it fails
    Then the error should contain "division by zero"

  Scenario: Update case syntax error - missing case keyword
    Given the following input content:
      """
      %im 0.1
      ---
      {a: 1} update { at .a -> 2 }
      """
    When I run the application and it fails
    Then the error should contain "update case must start with 'case'"

  Scenario: Update case syntax error - missing at keyword
    Given the following input content:
      """
      %im 0.1
      ---
      {a: 1} update { case .a -> 2 }
      """
    When I run the application and it fails
    Then the error should contain "update case missing 'at' keyword"

  Scenario: Headerless script returns error
    Given the following input content:
      """
      42
      """
    When I run the application and it fails
    Then the error should contain "script must have a header with '---' separator"

  Scenario: Case without a match returns error
    Given the following input content:
      """
      %im 0.1
      ---
      1 case {
        2 -> "two"
      }
      """
    When I run the application and it fails
    Then the error should contain "no case matched the value"
