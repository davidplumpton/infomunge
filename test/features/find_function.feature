Feature: Find Function
  As a developer
  I want to use the 'find' function to locate items in arrays and substrings or patterns in strings

  Scenario: Find items in an array
    When I evaluate "[1, 2, 3, 2] find 2"
    Then the output should be "[1,3]"

  Scenario: Find substring in a string
    When I evaluate "'abbbc' find 'b'"
    Then the output should be "[1,2,3]"

  Scenario: Find regex matches in a string
    When I evaluate "'123-456-7890' find '[0-9]+'"
    Then the output should be "[[0,3],[4,7],[8,12]]"

  Scenario: Find regex matches with flags
    When I evaluate "find('Hello World', 'hello', 'i')"
    Then the output should be "[[0,5]]"

  Scenario: Find no matches in array
    When I evaluate "[1, 2, 3] find 4"
    Then the output should be "[]"

  Scenario: Find no matches in string
    When I evaluate "'hello' find 'world'"
    Then the output should be "[]"

  Scenario: Find with special characters as literal (non-regex heuristic)
    When I evaluate "'a.b' find '.'"
    Then the output should be "[1]"

  Scenario: Find with special characters as regex (heuristic)
    When I evaluate "'a.b' find '\.'"
    Then the output should be "[[1,2]]"
