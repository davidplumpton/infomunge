Feature: Logging Functions
  In order to debug and trace execution
  As a developer
  I want to use logging functions that return their input value

  # log function - basic logging
  Scenario: log returns the input value unchanged
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log(42)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      42
      """
    And the stderr should contain "42"

  Scenario: log with string value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log("hello")
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      "hello"
      """

  Scenario: log with array value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log([1, 2, 3])
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      [1,2,3]
      """

  Scenario: log with null value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log(null)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      null
      """

  Scenario: log with boolean value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log(true)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      true
      """

  # logDebug function
  Scenario: logDebug returns the input value and logs with DEBUG prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logDebug(100)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      100
      """
    And the stderr should contain "[DEBUG]"

  Scenario: logDebug with string value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logDebug("debug message")
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      "debug message"
      """

  # logInfo function
  Scenario: logInfo returns the input value and logs with INFO prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logInfo(200)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      200
      """
    And the stderr should contain "[INFO]"

  Scenario: logInfo with string value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logInfo("info message")
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      "info message"
      """

  # logWarn function
  Scenario: logWarn returns the input value and logs with WARN prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logWarn(300)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      300
      """
    And the stderr should contain "[WARN]"

  Scenario: logWarn with string value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logWarn("warning message")
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      "warning message"
      """

  # logError function
  Scenario: logError returns the input value and logs with ERROR prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logError(400)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      400
      """
    And the stderr should contain "[ERROR]"

  Scenario: logError with string value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logError("error message")
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      "error message"
      """

  # logWith function - custom prefix
  Scenario: logWith returns the input value and logs with custom prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logWith(500, "CUSTOM")
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      500
      """
    And the stderr should contain "[CUSTOM]"

  Scenario: logWith with string value and custom prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logWith("custom message", "MY_TAG")
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      "custom message"
      """

  # Error handling
  Scenario: log requires exactly one argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log()
      """
    When I run the application and it fails
    Then the output should contain "log requires exactly 1 argument"

  Scenario: log with too many arguments fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log(1, 2)
      """
    When I run the application and it fails
    Then the output should contain "log requires exactly 1 argument"

  Scenario: logWith requires exactly two arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logWith(1)
      """
    When I run the application and it fails
    Then the output should contain "logWith requires exactly 2 arguments"

  Scenario: logWith prefix must be string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      logWith("value", 123)
      """
    When I run the application and it fails
    Then the output should contain "logWith expects a string"

  # Logging in expressions
  Scenario: log can be chained in expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log(5) + log(3)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      8
      """

  Scenario: log with computed expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      log(10 * 5)
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      50
      """

  # Different log levels work correctly
  Scenario: All log levels output with correct prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        debug: logDebug("d"),
        info: logInfo("i"),
        warn: logWarn("w"),
        error: logError("e")
      }
      """
    When I run the application with this content
    Then the stdout should be valid JSON equal to:
      """
      {"debug":"d","info":"i","warn":"w","error":"e"}
      """
    And the stderr should contain "[DEBUG]"
    And the stderr should contain "[INFO]"
    And the stderr should contain "[WARN]"
    And the stderr should contain "[ERROR]"
