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

  Scenario: Import Strings module and call with namespace
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

  Scenario Outline: Strings null-helper overloads preserve null
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Strings
      ---
      <expression>
      """
    When I run the application with this content
    Then the output should be:
      """
      null
      """

    Examples:
      | expression                         |
      | appendIfMissing(null, ".txt")      |
      | prependIfMissing(null, "prefix")   |
      | charCodeAt(null, 0)                |
      | fromCharCode(null)                 |

  Scenario: Direct appendIfMissing builtin remains strict for null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      appendIfMissing(null, ".txt")
      """
    When I run the application and it fails
    Then the output should contain "appendIfMissing expects a string as argument 1"
