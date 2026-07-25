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
      ["hello 123","hello","123"]
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
      [["foo1","foo","1"],["bar2","bar","2"]]
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

  Scenario: Contains treats regex metacharacters in strings literally
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [contains("a|b", "|"), contains("a[]b", "[]"), contains("a\\b", "\\"), contains("a+b", "+"), contains("a?b", "?")]
      """
    When I run the script
    Then the output should be:
      """
      [true,true,true,true,true]
      """

  Scenario: Contains supports explicit regex equivalents for literal metacharacters
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [contains("a|b", regex("\\|")), contains("a[]b", regex("\\[\\]")), contains("a\\b", regex("\\\\")), contains("a+b", regex("\\+")), contains("a?b", regex("\\?"))]
      """
    When I run the script
    Then the output should be:
      """
      [true,true,true,true,true]
      """

  Scenario: Find treats regex metacharacters in strings literally
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [find("a|b", "|"), find("a[]b", "[]"), find("a\\b", "\\"), find("a+b", "+"), find("a?b", "?")]
      """
    When I run the script
    Then the output should be:
      """
      [[1],[1],[1],[1],[1]]
      """

  Scenario: Find supports explicit regex equivalents for literal metacharacters
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [find("a|b", regex("\\|")), find("a[]b", regex("\\[\\]")), find("a\\b", regex("\\\\")), find("a+b", regex("\\+")), find("a?b", regex("\\?"))]
      """
    When I run the script
    Then the output should be:
      """
      [[[1,2]],[[1,3]],[[1,2]],[[1,2]],[[1,2]]]
      """

  Scenario: Replace treats regex metacharacters in strings literally
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [replace("a|b", "|", "X"), replace("a[]b", "[]", "X"), replace("a\\b", "\\", "X"), replace("a+b", "+", "X"), replace("a?b", "?", "X")]
      """
    When I run the script
    Then the output should be:
      """
      ["aXb","aXb","aXb","aXb","aXb"]
      """

  Scenario: Replace supports explicit regex equivalents for literal metacharacters
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [replace("a|b", regex("\\|"), "X"), replace("a[]b", regex("\\[\\]"), "X"), replace("a\\b", regex("\\\\"), "X"), replace("a+b", regex("\\+"), "X"), replace("a?b", regex("\\?"), "X")]
      """
    When I run the script
    Then the output should be:
      """
      ["aXb","aXb","aXb","aXb","aXb"]
      """

  Scenario: SplitBy treats regex metacharacters in strings literally
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [splitBy("a|b", "|"), splitBy("a[]b", "[]"), splitBy("a\\b", "\\"), splitBy("a+b", "+"), splitBy("a?b", "?")]
      """
    When I run the script
    Then the output should be:
      """
      [["a","b"],["a","b"],["a","b"],["a","b"],["a","b"]]
      """

  Scenario: SplitBy supports explicit regex equivalents for literal metacharacters
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [splitBy("a|b", regex("\\|")), splitBy("a[]b", regex("\\[\\]")), splitBy("a\\b", regex("\\\\")), splitBy("a+b", regex("\\+")), splitBy("a?b", regex("\\?"))]
      """
    When I run the script
    Then the output should be:
      """
      [["a","b"],["a","b"],["a","b"],["a","b"],["a","b"]]
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
      ["hello","hello"]
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

  Scenario: Infix scan operator with regex literal
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "abc123def" scan /[0-9]+/
      """
    When I run the script
    Then the output should be:
      """
      [["123"]]
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

  Scenario: Infix matches after string ending in escaped backslashes
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "a\\\\" matches /^a\\\\$/
      """
    When I run the script
    Then the output should be:
      """
      true
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

  Scenario: Division operator after array indexing
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3][1] / 2
      """
    When I run the script
    Then the output should be:
      """
      1
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
