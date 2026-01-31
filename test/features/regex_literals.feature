Feature: Regex Literals
  In order to write cleaner regex patterns
  As a developer
  I want to use the /pattern/ syntax for regexes

  Scenario: Regex literal with match function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      match("hello 123", /([a-z]+) ([0-9]+)/)
      """
    When I run the script
    Then the output should be:
      """
      ["hello","123"]
      """

  Scenario: Regex literal with matches function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      matches("hello", /^[a-z]+$/)
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Regex literal with scan function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      scan("foo1 bar2", /([a-z]+)([0-9])/)
      """
    When I run the script
    Then the output should be:
      """
      [["foo","1"],["bar","2"]]
      """

  Scenario: Regex literal with find function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      find("foo bar baz", /ba./)
      """
    When I run the script
    Then the output should be:
      """
      [[4,7],[8,11]]
      """

  Scenario: Regex literal with contains function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      contains("hello world", /world/)
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Regex literal with splitBy function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      splitBy("a,b;c:d", /[,;:]/)
      """
    When I run the script
    Then the output should be:
      """
      ["a","b","c","d"]
      """

  Scenario: Regex literal with replace function
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      replace("foo bar baz", /b(..)/, "X")
      """
    When I run the script
    Then the output should be:
      """
      "foo X X"
      """

  Scenario: Regex literal with flags (case insensitive)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      matches("HELLO", /hello/i)
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Regex literal with slashes in pattern
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      matches("https://example.com", /https?:\/\/.+/)
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Infix match operator with regex literal
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" match /([a-z]+)/
      """
    When I run the script
    Then the output should be:
      """
      ["hello"]
      """

  Scenario: Infix matches operator with regex literal
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" matches /^[a-z]+$/
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Infix matches operator with regex literal and flags
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "HELLO" matches /hello/i
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Replace operator with regex literal
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "foo123bar" replace /[0-9]+/ with "NUM"
      """
    When I run the script
    Then the output should be:
      """
      "fooNUMbar"
      """

  Scenario: Type of regex literal
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(/abc/)
      """
    When I run the script
    Then the output should be:
      """
      "Regex"
      """

  Scenario: Is Regex type check
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      /abc/ is Regex
      """
    When I run the script
    Then the output should be:
      """
      true
      """
