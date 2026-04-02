Feature: Import Directive
  In order to reuse code across scripts
  As a developer
  I want to import modules and use their functions

  Scenario: Import module and call function with namespace
    Given the following input content:
      """
      %im 0.1
      import modules::MathUtils
      output application/json
      ---
      MathUtils::double(21)
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Import all from module with star
    Given the following input content:
      """
      %im 0.1
      import * from modules::MathUtils
      output application/json
      ---
      double(21)
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Import specific functions from module
    Given the following input content:
      """
      %im 0.1
      import double, triple from modules::MathUtils
      output application/json
      ---
      {doubled: double(7), tripled: triple(7)}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"doubled":14,"tripled":21}
      """

  Scenario: Module function uses module variable
    Given the following input content:
      """
      %im 0.1
      import modules::MathUtils
      output application/json
      ---
      MathUtils::addOffset(10)
      """
    When I run the application with this content
    Then the output should be:
      """
      15
      """

  Scenario: Module function uses do block separator
    Given the following input content:
      """
      %im 0.1
      import modules::DoBlock
      output application/json
      ---
      DoBlock::addOne(2)
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: Module function uses multiple module variables
    Given the following input content:
      """
      %im 0.1
      import modules::MathUtils
      output application/json
      ---
      MathUtils::scaleAndAdd(3)
      """
    When I run the application with this content
    Then the output should be:
      """
      35
      """

  Scenario: Module function uses multiline module variable
    Given the following input content:
      """
      %im 0.1
      import modules::MathUtils
      output application/json
      ---
      MathUtils::scaleAndAddSettings(3)
      """
    When I run the application with this content
    Then the output should be:
      """
      35
      """

  Scenario: Import module variable with star
    Given the following input content:
      """
      %im 0.1
      import * from modules::MathUtils
      output application/json
      ---
      scale + offset
      """
    When I run the application with this content
    Then the output should be:
      """
      15
      """

  Scenario: Mix imported and local functions
    Given the following input content:
      """
      %im 0.1
      import double from modules::MathUtils
      fun quadruple(x) = double(double(x))
      output application/json
      ---
      quadruple(5)
      """
    When I run the application with this content
    Then the output should be:
      """
      20
      """

  Scenario: Multiple module imports
    Given the following input content:
      """
      %im 0.1
      import modules::MathUtils
      import modules::StringUtils
      output application/json
      ---
      {math: MathUtils::double(5), greeting: StringUtils::greet("World")}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"greeting":"Hello, World","math":10}
      """

  Scenario: Import with star and call directly
    Given the following input content:
      """
      %im 0.1
      import * from modules::StringUtils
      output application/json
      ---
      greet("Alice")
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello, Alice"
      """

  Scenario: Module function calls builtin
    Given the following input content:
      """
      %im 0.1
      import shout from modules::StringUtils
      output application/json
      ---
      shout("hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      "HELLO"
      """

  Scenario: Import functions from multiple dw core modules
    Given the following input content:
      """
      %im 0.1
      import keySet from dw::core::Objects
      import concatWith from dw::core::Binaries
      output application/json
      ---
      {keys: keySet({b: 2, a: 1}), joined: concatWith(["x", "y"], ":")}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"joined":"x:y","keys":["a","b"]}
      """

  Scenario: Reject module import path traversal
    Given the following input content:
      """
      %im 0.1
      import ..::..::tmp::evil
      output application/json
      ---
      1
      """
    When I run the application and it fails
    Then the error should contain "invalid module spec"

  Scenario: Script file imports resolve relative to the script directory
    Given a file named "nested/modules/MathUtils.im" with content:
      """
      %im 0.1
      fun double(x) = x * 2
      """
    And a file named "nested/script.im" with content:
      """
      %im 0.1
      import modules::MathUtils
      output application/json
      ---
      MathUtils::double(21)
      """
    When I run the application with "nested/script.im"
    Then the output should be:
      """
      42
      """
