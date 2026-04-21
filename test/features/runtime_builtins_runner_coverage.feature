Feature: Runtime builtins runner coverage
  In order to measure cucumber coverage for runtime introspection builtins
  As a maintainer
  I want runtime builtins exercised in-process

  # --- uuid ---

  Scenario: uuid returns a non-null string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      uuid() != null
      """
    When I run the script
    Then the output should be true

  # --- evaluateCompatibilityFlag ---

  Scenario: evaluateCompatibilityFlag returns true for allowUndefinedProperties
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("allowUndefinedProperties")
      """
    When I run the script
    Then the output should be true

  Scenario: evaluateCompatibilityFlag returns false for allowNullArithmetic
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("allowNullArithmetic")
      """
    When I run the script
    Then the output should be false

  Scenario: evaluateCompatibilityFlag returns true for allowImplicitConversion
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("allowImplicitConversion")
      """
    When I run the script
    Then the output should be true

  Scenario: evaluateCompatibilityFlag returns false for strictTypeChecking
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("strictTypeChecking")
      """
    When I run the script
    Then the output should be false

  Scenario: evaluateCompatibilityFlag returns true for allowDynamicProperties
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("allowDynamicProperties")
      """
    When I run the script
    Then the output should be true

  Scenario: evaluateCompatibilityFlag returns null for unknown flag
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag("unknownFlag")
      """
    When I run the script
    Then the output should be null

  Scenario: evaluateCompatibilityFlag expects string argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag(123)
      """
    Then running the script should fail with error containing "evaluateCompatibilityFlag expects flagName to be a string"

  # --- envVar/envVars (unique in-process scenarios) ---

  Scenario: envVars contains PATH key
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      envVars().PATH != null
      """
    When I run the script
    Then the output should be true

  Scenario: envVar and envVars are consistent
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      envVar("PATH") == envVars().PATH
      """
    When I run the script
    Then the output should be true

  # --- log ---

  Scenario: log returns its input value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      log(42)
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: log returns string unchanged
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      log("hello")
      """
    When I run the script
    Then the output should be "hello"

  # --- logDebug ---

  Scenario: logDebug returns its input value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logDebug(100)
      """
    When I run the script
    Then the output should be:
      """
      100
      """

  Scenario: logDebug requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logDebug()
      """
    Then running the script should fail with error containing "logDebug requires exactly 1 argument"

  # --- logInfo ---

  Scenario: logInfo returns its input value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logInfo(200)
      """
    When I run the script
    Then the output should be:
      """
      200
      """

  Scenario: logInfo requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logInfo()
      """
    Then running the script should fail with error containing "logInfo requires exactly 1 argument"

  # --- logWarn ---

  Scenario: logWarn returns its input value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logWarn(300)
      """
    When I run the script
    Then the output should be:
      """
      300
      """

  Scenario: logWarn requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logWarn()
      """
    Then running the script should fail with error containing "logWarn requires exactly 1 argument"

  # --- logError ---

  Scenario: logError returns its input value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logError(400)
      """
    When I run the script
    Then the output should be:
      """
      400
      """

  Scenario: logError requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logError()
      """
    Then running the script should fail with error containing "logError requires exactly 1 argument"

  # --- logWith ---

  Scenario: logWith returns its input value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logWith(500, "CUSTOM")
      """
    When I run the script
    Then the output should be:
      """
      500
      """

  # --- fail (unique in-process scenarios) ---

  Scenario: fail with custom message
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      fail("Something went wrong")
      """
    Then running the script should fail with error containing "Something went wrong"

  Scenario: fail with default message
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      fail()
      """
    Then running the script should fail with error containing "Error"

  # --- try (unique in-process scenarios) ---

  Scenario: try catches fail and returns success false
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try(() -> fail("oops")).success
      """
    When I run the script
    Then the output should be false

  Scenario: try returns success true for successful expression
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try(() -> 42).success
      """
    When I run the script
    Then the output should be true

  Scenario: try returns result for successful expression
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try(() -> 42).result
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: try returns error message on failure
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try(() -> fail("Custom error")).error.message
      """
    When I run the script
    Then the output should be "Custom error"

  Scenario: try with named zero-arg function
    Given the following script:
      """
      %im 0.1
      output application/json
      var items = [1, 2, 3]
      fun doubled() = items map $ * 2
      ---
      try(doubled).success
      """
    When I run the script
    Then the output should be true

  # --- orElse (unique in-process scenarios) ---

  Scenario: orElse with fallback lambda
    Given the following script:
      """
      %im 0.1
      output application/json
      fun getDefault() = "computed default"
      ---
      orElse(try(() -> fail("error")), getDefault)
      """
    When I run the script
    Then the output should be "computed default"

  # --- orElseTry (unique in-process scenarios) ---

  Scenario: orElseTry chain falls through to second
    Given the following script:
      """
      %im 0.1
      output application/json
      fun first() = fail("no first")
      fun second() = 2
      fun third() = 3
      ---
      orElse(orElseTry(orElseTry(try(first), second), third), 0)
      """
    When I run the script
    Then the output should be:
      """
      2
      """

  Scenario: orElseTry requires TryResult as first argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElseTry("not a try result", () -> 1)
      """
    Then running the script should fail with error containing "orElseTry: first argument must be a TryResult object"
