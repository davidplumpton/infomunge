Feature: Numbers Module
  In order to use DataWeave number conversion helpers
  As a developer
  I want the dw::core::Numbers module to expose radix conversion functions

  Scenario: toBinary converts a positive integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Numbers
      ---
      toBinary(13)
      """
    When I run the application with this content
    Then the output should be:
      """
      "1101"
      """

  Scenario: fromBinary converts to an integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Numbers
      ---
      fromBinary("1101")
      """
    When I run the application with this content
    Then the output should be:
      """
      13
      """

  Scenario: toRadix converts to base 16
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Numbers
      ---
      toRadix(255, 16)
      """
    When I run the application with this content
    Then the output should be:
      """
      "FF"
      """

  Scenario: fromRadix converts from base 16
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Numbers
      ---
      fromRadix("FF", 16)
      """
    When I run the application with this content
    Then the output should be:
      """
      255
      """

  Scenario: Import module and call with namespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      import dw::core::Numbers
      ---
      Numbers::toRadix(-10, 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      "-1010"
      """
