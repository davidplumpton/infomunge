Feature: Strings Module
  In order to use DataWeave string helpers from a module
  As a developer
  I want the dw::core::Strings module to expose missing functions

  Scenario: appendIfMissing appends a missing suffix
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Strings
      ---
      appendIfMissing("abc", ".txt")
      """
    When I run the application with this content
    Then the output should be:
      """
      "abc.txt"
      """

  Scenario: appendIfMissing does not duplicate suffix
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Strings
      ---
      appendIfMissing("abc.txt", ".txt")
      """
    When I run the application with this content
    Then the output should be:
      """
      "abc.txt"
      """

  Scenario: prependIfMissing adds a missing prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Strings
      ---
      prependIfMissing("world", "hello ")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello world"
      """

  Scenario: charCodeAt reads unicode code point
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Strings
      ---
      charCodeAt("Aé", 1)
      """
    When I run the application with this content
    Then the output should be:
      """
      233
      """

  Scenario: fromCharCode builds characters from code points
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Strings
      ---
      fromCharCode(9731)
      """
    When I run the application with this content
    Then the output should be:
      """
      "☃"
      """

  Scenario: Import module and call with namespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      import dw::core::Strings
      ---
      Strings::charCodeAt("AZ", 1)
      """
    When I run the application with this content
    Then the output should be:
      """
      90
      """
