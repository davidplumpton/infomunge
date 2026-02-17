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

  Scenario: uuid returns different values on successive calls
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      uuid() != uuid()
      """
    When I run the script
    Then the output should be true

  Scenario: uuid takes no arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      uuid("bad")
      """
    Then running the script should fail with error containing "uuid takes no arguments"

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

  Scenario: evaluateCompatibilityFlag requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag()
      """
    Then running the script should fail with error containing "evaluateCompatibilityFlag requires exactly 1 argument"

  Scenario: evaluateCompatibilityFlag expects string argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      evaluateCompatibilityFlag(123)
      """
    Then running the script should fail with error containing "evaluateCompatibilityFlag expects flagName to be a string"

  # --- envVar ---

  Scenario: envVar returns value for existing variable
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      envVar("PATH") != null
      """
    When I run the script
    Then the output should be true

  Scenario: envVar returns null for non-existent variable
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      envVar("INFOMUNGE_NONEXISTENT_VAR_12345")
      """
    When I run the script
    Then the output should be null

  Scenario: envVar requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      envVar()
      """
    Then running the script should fail with error containing "envVar requires exactly 1 argument"

  Scenario: envVar expects string argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      envVar(123)
      """
    Then running the script should fail with error containing "envVar expects a string argument"

  # --- envVars ---

  Scenario: envVars returns an object
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(envVars())
      """
    When I run the script
    Then the output should be "Object"

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

  Scenario: envVars takes no arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      envVars("unexpected")
      """
    Then running the script should fail with error containing "envVars takes no arguments"

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

  Scenario: log requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      log()
      """
    Then running the script should fail with error containing "log requires exactly 1 argument"

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

  Scenario: logWith requires exactly two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logWith(1)
      """
    Then running the script should fail with error containing "logWith requires exactly 2 arguments"

  Scenario: logWith prefix must be string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      logWith("value", 123)
      """
    Then running the script should fail with error containing "logWith expects a string"

  # --- fail (in-process) ---

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

  Scenario: fail expects string argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      fail(123)
      """
    Then running the script should fail with error containing "fail expects a string argument"

  Scenario: fail accepts at most one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      fail("a", "b")
      """
    Then running the script should fail with error containing "fail accepts at most 1 argument"

  # --- try (in-process) ---

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

  Scenario: try returns error kind for user exceptions
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try(() -> fail("test")).error.kind
      """
    When I run the script
    Then the output should be "UserException"

  Scenario: try catches runtime errors
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try(() -> (1 / 0)).success
      """
    When I run the script
    Then the output should be false

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

  Scenario: try requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      try()
      """
    Then running the script should fail with error containing "try requires exactly 1 argument"

  # --- orElse (in-process) ---

  Scenario: orElse returns result when try succeeds
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> 42), "fallback")
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: orElse returns fallback when try fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> fail("error")), "fallback")
      """
    When I run the script
    Then the output should be "fallback"

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

  Scenario: orElse requires exactly two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElse(try(() -> 1))
      """
    Then running the script should fail with error containing "orElse requires exactly 2 arguments"

  Scenario: orElse requires TryResult as first argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElse("not a try result", "fallback")
      """
    Then running the script should fail with error containing "orElse: first argument must be a TryResult object"

  # --- orElseTry (in-process) ---

  Scenario: orElseTry returns original TryResult when try succeeds
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> 42), () -> 99).result
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: orElseTry evaluates fallback when try fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> fail("error")), () -> 99).result
      """
    When I run the script
    Then the output should be:
      """
      99
      """

  Scenario: orElseTry captures fallback failure
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> fail("first")), () -> fail("second")).success
      """
    When I run the script
    Then the output should be false

  Scenario: orElseTry chain with first success
    Given the following script:
      """
      %im 0.1
      output application/json
      fun first() = 1
      fun second() = 2
      fun third() = 3
      ---
      orElse(orElseTry(orElseTry(try(first), second), third), 0)
      """
    When I run the script
    Then the output should be:
      """
      1
      """

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

  Scenario: orElseTry requires exactly two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElseTry(try(() -> 1))
      """
    Then running the script should fail with error containing "orElseTry requires exactly 2 arguments"

  Scenario: orElseTry requires TryResult as first argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      orElseTry("not a try result", () -> 1)
      """
    Then running the script should fail with error containing "orElseTry: first argument must be a TryResult object"
