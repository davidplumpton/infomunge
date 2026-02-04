Feature: Binaries Module
  In order to use DataWeave binary helpers from a module
  As a developer
  I want the dw::core::Binaries module to expose missing functions

  Scenario: concatWith joins values with a separator
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Binaries
      ---
      concatWith(["a", "b", "c"], "|")
      """
    When I run the application with this content
    Then the output should be:
      """
      "a|b|c"
      """

  Scenario: readLinesWith splits content using the separator
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Binaries
      ---
      readLinesWith("alpha\nbeta\ngamma", "\n")
      """
    When I run the application with this content
    Then the output should be:
      """
      ["alpha","beta","gamma"]
      """

  Scenario: writeLinesWith joins lines using the separator
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::Binaries
      ---
      writeLinesWith(["first", "second"], "\n")
      """
    When I run the application with this content
    Then the output should be:
      """
      "first\nsecond"
      """

  Scenario: Import module and call with namespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      import dw::core::Binaries
      ---
      Binaries::concatWith(["x", "y"], ",")
      """
    When I run the application with this content
    Then the output should be:
      """
      "x,y"
      """
