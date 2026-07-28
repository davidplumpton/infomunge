Feature: Number Indexing
  In order to use DataWeave-compatible ordinal selectors
  As a developer
  I want to index the standard string rendering of numbers

  Scenario: Index positive, negative, and decimal numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        "integer": [(123)[0], (123)[-1]],
        "negative": [(-123)[0], (-123)[1]],
        "decimal": [(12.5)[2], (0.0000001)[1]]
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"integer":["1","3"],"negative":["-","1"],"decimal":[".","E"]}
      """

  Scenario: Number indexes outside either bound return null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [(123)[3], (123)[-4]]
      """
    When I run the application with this content
    Then the output should be:
      """
      [null,null]
      """

  Scenario: Number ordinal selector strings support numeric negation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      -((12.5)[-3])
      """
    When I run the application with this content
    Then the output should be:
      """
      -2
      """
