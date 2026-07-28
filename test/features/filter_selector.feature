Feature: Filter selector
  In order to select collection members compatibly with DataWeave
  As a transformation author
  I want filter selectors to distinguish no matches from unsupported sources

  Scenario: An array filter selector with no matches allows default
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(([1][?($ > 9)] default [7]))
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario: An object filter selector with no matches allows default
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(({a: 1}[?($ > 9)] default {fallback: 0}))
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  Scenario Outline: Filter selectors reject unsupported source types
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      <source>[?($ == <candidate>)]
      """
    When I run the application and it fails
    Then the error should contain "selector filter expects an array or object"

    Examples:
      | source | candidate |
      | "abc"  | "b"       |
      | 123    | "2"       |
      | null   | 1         |
