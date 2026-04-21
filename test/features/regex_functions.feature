Feature: Regex Functions
  In order to perform pattern matching and regex operations
  As a developer
  I want to use the regex functions: match, matches, scan

  Scenario: match with capture groups
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      match("foo123bar", "([a-z]+)([0-9]+)")
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2
    And the output should contain "foo"
    And the output should contain "123"

  Scenario: match no capture groups
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      match("hello world", "w[a-z]+")
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: match with flags case insensitive
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      match("Hello World", "([a-z]+)", "i")
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 1

  Scenario: match no match returns null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      match("hello", "[0-9]+")
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

  Scenario: match with optional group
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      match("test123", "([a-z]+)(\\d*)")
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2

  Scenario: matches basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      matches("hello world", "l")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: matches returns false
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      matches("hello", "[0-9]+")
      """
    When I run the application with this content
    Then the output should be:
      """
      false
      """

  Scenario: matches with pattern at start
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      matches("hello", "^h")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: matches case insensitive
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      matches("Hello", "hello", "i")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: matches empty pattern
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      matches("", "")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: scan basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      scan("hello world hello", "hello")
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2

  Scenario: scan with capture groups
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      scan("foo1 bar2 baz3", "([a-z]+)([0-9])")
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 3

  Scenario: scan no matches
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      scan("hello", "[0-9]+")
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: scan single match
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      scan("abc123def", "[0-9]+")
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 1
    And the output should contain "123"

  Scenario: scan overlapping patterns
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      scan("aaaa", "a+")
      """
    When I run the application with this content
    Then the output should contain "aaaa"

  Scenario: match with non-string input
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      match(123, "[0-9]+")
      """
    When I run the application and it fails
    Then the output should contain "match"

  Scenario: matches with non-string input
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      matches(null, "test")
      """
    When I run the application and it fails
    Then the output should contain "matches"

  Scenario: scan with non-string input
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      scan(456, "[0-9]")
      """
    When I run the application and it fails
    Then the output should contain "scan"

  Scenario: match with special characters
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      match("test@example.com", "([a-z]+)@([a-z]+)\\.([a-z]+)")
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 3
    And the output should contain "test"

  Scenario: matches word boundary
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      matches("hello world", "\\bhello\\b")
      """
    When I run the application with this content
    Then the output should be true

   Scenario: scan digits
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       scan("a1b2c3d4", "[0-9]")
       """
     When I run the application with this content
     Then the output should be valid JSON with array length of 4

   Scenario: matches operator basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       "hello world" matches "l"
       """
     When I run the application with this content
     Then the output should be true

   Scenario: matches operator returns false
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       "hello" matches "[0-9]+"
       """
     When I run the application with this content
     Then the output should be false

   Scenario: matches operator with pattern at start
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       "hello" matches "^h"
       """
     When I run the application with this content
     Then the output should be true


