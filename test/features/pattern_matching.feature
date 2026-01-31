Feature: Pattern Matching
  Scenario: Match against literal values
    Given the following script:
      """
      ---
      "foo" case {
        "bar" -> "it is bar",
        "foo" -> "it is foo",
        else -> "it is something else"
      }
      """
    When I run the script
    Then the output should be:
      """
      "it is foo"
      """

  Scenario: Match against types
    Given the following script:
      """
      ---
      123 case {
        is String -> "it is a string",
        is Number -> "it is a number",
        else -> "unknown type"
      }
      """
    When I run the script
    Then the output should be:
      """
      "it is a number"
      """

  Scenario: Match with variable capture
    Given the following script:
      """
      ---
      "hello" case {
        s is String -> "string length is " + sizeOf(s),
        n is Number -> "number is " + n,
        else -> "unknown"
      }
      """
    When I run the script
    Then the output should be:
      """
      "string length is 5"
      """

  Scenario: Match with conditions
    Given the following script:
      """
      ---
      15 case {
        n is Number if n > 10 -> "greater than 10",
        n is Number -> "10 or less",
        else -> "not a number"
      }
      """
    When I run the script
    Then the output should be:
      """
      "greater than 10"
      """

  Scenario: Match keyword with conditions
    Given the following script:
      """
      ---
      15 match {
        case n is Number if n > 10 -> "greater than 10"
        else -> "not a number"
      }
      """
    When I run the script
    Then the output should be:
      """
      "greater than 10"
      """

  Scenario: Match keyword with case labels
    Given the following script:
      """
      ---
      "hello" match {
        case is String -> "it is a string"
        else -> "not a string"
      }
      """
    When I run the script
    Then the output should be:
      """
      "it is a string"
      """

  Scenario: Match keyword with multiple case labels without commas
    Given the following script:
      """
      ---
      123 match {
        case is String -> "string"
        case is Number -> "number"
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "number"
      """

  # Default case handling
  Scenario: Default case when no match found
    Given the following script:
      """
      ---
      true case {
        "foo" -> "matched foo",
        "bar" -> "matched bar",
        else -> "no match"
      }
      """
    When I run the script
    Then the output should be:
      """
      "no match"
      """

  Scenario: Match with only else branch
    Given the following script:
      """
      ---
      "anything" case {
        else -> "always this"
      }
      """
    When I run the script
    Then the output should be:
      """
      "always this"
      """

  # Type guards
  Scenario: Match against Boolean type
    Given the following script:
      """
      ---
      true case {
        is String -> "string",
        is Number -> "number",
        is Boolean -> "boolean",
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "boolean"
      """

  Scenario: Match against Array type
    Given the following script:
      """
      ---
      [1, 2, 3] case {
        is String -> "string",
        is Array -> "array",
        is Object -> "object",
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "array"
      """

  Scenario: Match against Object type
    Given the following script:
      """
      ---
      {a: 1} case {
        is String -> "string",
        is Array -> "array",
        is Object -> "object",
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "object"
      """

  Scenario: Match against Null type
    Given the following script:
      """
      ---
      null case {
        is String -> "string",
        is Null -> "null value",
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "null value"
      """

  # Multiple conditions
  Scenario: Match with multiple guards on same type
    Given the following script:
      """
      ---
      5 case {
        n is Number if n > 10 -> "large",
        n is Number if n > 5 -> "medium",
        n is Number if n > 0 -> "small",
        else -> "zero or negative"
      }
      """
    When I run the script
    Then the output should be:
      """
      "small"
      """

  Scenario: Match first matching condition
    Given the following script:
      """
      ---
      15 case {
        n is Number if n > 5 -> "over 5",
        n is Number if n > 10 -> "over 10",
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "over 5"
      """

  # Variable capture with expressions
  Scenario: Variable capture used in expression
    Given the following script:
      """
      ---
      10 case {
        n is Number -> n * 2,
        else -> 0
      }
      """
    When I run the script
    Then the output should be:
      """
      20
      """

  Scenario: Variable capture with string concatenation
    Given the following script:
      """
      ---
      "world" case {
        s is String -> "hello " ++ s,
        else -> "not a string"
      }
      """
    When I run the script
    Then the output should be:
      """
      "hello world"
      """

  # Literal matching
  Scenario: Match exact number literal
    Given the following script:
      """
      ---
      42 case {
        0 -> "zero",
        1 -> "one",
        42 -> "the answer",
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "the answer"
      """

  Scenario: Match exact boolean literal
    Given the following script:
      """
      ---
      false case {
        true -> "yes",
        false -> "no",
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "no"
      """

  Scenario: Match null literal
    Given the following script:
      """
      ---
      null case {
        null -> "was null",
        else -> "not null"
      }
      """
    When I run the script
    Then the output should be:
      """
      "was null"
      """

  # Complex expressions
  Scenario: Match result of expression
    Given the following script:
      """
      ---
      (5 + 5) case {
        5 -> "five",
        10 -> "ten",
        else -> "other"
      }
      """
    When I run the script
    Then the output should be:
      """
      "ten"
      """

  Scenario: Match with complex guard expression
    Given the following script:
      """
      ---
      [1, 2, 3] case {
        arr is Array if sizeOf(arr) > 2 -> "large array",
        arr is Array -> "small array",
        else -> "not an array"
      }
      """
    When I run the script
    Then the output should be:
      """
      "large array"
      """

  # Match keyword variations
  Scenario: Match keyword with literal
    Given the following script:
      """
      ---
      "test" match {
        case "test" -> "matched test"
        else -> "no match"
      }
      """
    When I run the script
    Then the output should be:
      """
      "matched test"
      """

  Scenario: Match keyword with variable and guard
    Given the following script:
      """
      ---
      [1, 2, 3, 4, 5] match {
        case arr is Array if sizeOf(arr) == 5 -> "five items"
        case arr is Array -> "some items"
        else -> "not array"
      }
      """
    When I run the script
    Then the output should be:
      """
      "five items"
      """
