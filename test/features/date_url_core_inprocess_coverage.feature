Feature: In-process coverage for date, URL, and core builtins
  In order to measure in-process coverage for these builtin families
  As a maintainer
  I want date/URL/core functions exercised through the in-process script runner

  # --- URL builtins (in-process via module import) ---

  Scenario: parseURI returns URL parts in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      parseURI("https://example.com/path?q=hello#frag")
      """
    When I run the script
    Then the output should contain "example.com"
    And the output should contain "/path"
    And the output should contain "frag"

  Scenario: encodeURIComponent encodes a string in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      encodeURIComponent("hello world")
      """
    When I run the script
    Then the output should be:
      """
      "hello%20world"
      """

  Scenario: decodeURIComponent decodes a component in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      decodeURIComponent("hello%20world")
      """
    When I run the script
    Then the output should be:
      """
      "hello world"
      """

  Scenario: encodeURI encodes a URI in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      encodeURI("https://example.com/a path")
      """
    When I run the script
    Then the output should contain "example.com"
    And the output should contain "a%20path"

  Scenario: decodeURI decodes a URI in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      decodeURI("https://example.com/a%20path")
      """
    When I run the script
    Then the output should be:
      """
      "https://example.com/a path"
      """

  Scenario: compose builds a simple URI in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      compose({scheme: "https", host: "example.com", path: "/test"})
      """
    When I run the script
    Then the output should contain "https://example.com/test"

  Scenario: parseURI with user and password credentials
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      parseURI("https://alice:secret@example.com/path")
      """
    When I run the script
    Then the output should contain "alice"
    And the output should contain "secret"

  Scenario: parseURI with port number
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      parseURI("https://example.com:9090/api")
      """
    When I run the script
    Then the output should contain "9090"
    And the output should contain "/api"

  Scenario: parseURI with query string
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      parseURI("https://example.com/search?q=hello&lang=en")
      """
    When I run the script
    Then the output should contain "hello"
    And the output should contain "en"

  Scenario: parseURI with non-string argument fails
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      parseURI(42)
      """
    Then running the script should fail with error containing "parseURI"

  Scenario: parseURI requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      parseURI()
      """
    Then running the script should fail with error containing "parseURI requires exactly 1 argument"

  Scenario: compose with port number
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      compose({scheme: "https", host: "example.com", port: 8443, path: "/api"})
      """
    When I run the script
    Then the output should contain "example.com:8443"
    And the output should contain "/api"

  Scenario: compose with user credentials
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      compose({scheme: "ftp", user: "admin", password: "pass", host: "files.example.com", path: "/data"})
      """
    When I run the script
    Then the output should contain "admin:pass"
    And the output should contain "files.example.com"

  Scenario: compose with query string
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      compose({scheme: "https", host: "example.com", path: "/search", query: {q: "hello world"}})
      """
    When I run the script
    Then the output should contain "example.com"
    And the output should contain "hello"

  Scenario: compose requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      compose()
      """
    Then running the script should fail with error containing "compose requires exactly 1 argument"

  Scenario: encodeURIComponent requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      encodeURIComponent()
      """
    Then running the script should fail with error containing "encodeURIComponent requires exactly 1 argument"

  Scenario: decodeURIComponent requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      decodeURIComponent()
      """
    Then running the script should fail with error containing "decodeURIComponent requires exactly 1 argument"

  # --- Core: parseGoSyntaxObjectLiteral / parseGoSyntaxArrayLiteral via reduce defaults ---
  # The preprocessor converts object literals like {count: 0} to Go map syntax before
  # embedding in __lambda param strings; parseGoSyntaxObjectLiteral handles these.
  # The reduce operator supports default accumulator values.

  Scenario: reduce with object literal default accumulator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] reduce (acc = {count: 0}, x) -> {count: acc.count + x}
      """
    When I run the script
    Then the output should be:
      """
      {"count":6}
      """

  Scenario: reduce with nested object literal default accumulator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [10, 20] reduce (acc = {sum: 0, count: 0}, x) -> {sum: acc.sum + x, count: acc.count + 1}
      """
    When I run the script
    Then the output should be:
      """
      {"sum":30,"count":2}
      """

  Scenario: reduce with array literal default accumulator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] reduce (acc = [0], x) -> acc ++ [x]
      """
    When I run the script
    Then the output should be:
      """
      [0,1,2,3]
      """

  Scenario: reduce with numeric default accumulator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4] reduce (acc = 0, x) -> acc + x
      """
    When I run the script
    Then the output should be:
      """
      10
      """

  # --- Core: callBuiltinForceEval error paths ---

  Scenario: force_eval with wrong argument count fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      force_eval()
      """
    Then running the script should fail with error containing "force_eval requires exactly 1 argument"

  Scenario: force_eval with non-lazy value fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      force_eval(42)
      """
    Then running the script should fail with error containing "force_eval argument must be a lazy value"

  # --- Core: callBuiltinOnNull in-process (non-null path and error paths) ---

  Scenario: onNull with non-null value returns value in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "valid" onNull (() -> "fallback")
      """
    When I run the script
    Then the output should be:
      """
      "valid"
      """

  Scenario: onNull with wrong lambda arity fails in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null onNull ((x) -> "fail")
      """
    Then running the script should fail with error containing "onNull lambda must have 0 parameters"

  Scenario: onNull with non-lambda fails in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null onNull "not-a-lambda"
      """
    Then running the script should fail with error containing "onNull expects a lambda function"

  # --- Core: callBuiltinThen in-process (null path and error paths) ---

  Scenario: then with null value returns null in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null then ((x) -> x ++ " world")
      """
    When I run the script
    Then the output should be null

  Scenario: then with wrong lambda arity fails in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" then (() -> "fail")
      """
    Then running the script should fail with error containing "then lambda must have exactly 1 parameter"

  Scenario: then with non-lambda fails in-process
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" then "not-a-lambda"
      """
    Then running the script should fail with error containing "then expects a lambda function"

  # --- Core: callBuiltinCoerce error and alternative paths ---

  Scenario: as operator with wrong argument count fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      __coerce(42)
      """
    Then running the script should fail with error containing "as operator requires 2 or 3 arguments"

  Scenario: as operator coerces number to string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      42 as String
      """
    When I run the script
    Then the output should be:
      """
      "42"
      """

  Scenario: as operator coerces string to number
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "3.14" as Number
      """
    When I run the script
    Then the output should be valid JSON with number close to 3.14

  Scenario: as operator with string literal type name
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      __coerce(42, "String")
      """
    When I run the script
    Then the output should be:
      """
      "42"
      """
