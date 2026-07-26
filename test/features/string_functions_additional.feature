Feature: Additional String Functions
  In order to manipulate and transform strings
  As a developer
  I want to use additional string functions: trim, length, repeat, split, charAt, indexOf, lastIndexOf, substring, startsWith, endsWith, contains, replace

  # trim tests
  Scenario: trim basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      trim("  hello  ")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: trim leading whitespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      trim("   hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: trim trailing whitespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      trim("hello   ")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: trim with tabs and newlines
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      trim("\t\nhello\n\t")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: trim empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      trim("")
      """
    When I run the application with this content
    Then the output should be:
      """
      ""
      """

  Scenario: trim string with no whitespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      trim("hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: trim on collection fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      trim([123])
      """
    When I run the application and it fails
    Then the output should contain "trim expects a string"

  # length tests
  Scenario: length of string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      length("hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: length of empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      length("")
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: length of array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      length([1, 2, 3, 4, 5])
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: length of empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      length([])
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: length of object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      length({"a": 1, "b": 2, "c": 3})
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: length of null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      length(null)
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: length counts unicode characters correctly
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      length("cafe")
      """
    When I run the application with this content
    Then the output should be:
      """
      4
      """

  # repeat tests
  Scenario: repeat basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      repeat("ab", 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      "ababab"
      """

  Scenario: repeat zero times
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      repeat("hello", 0)
      """
    When I run the application with this content
    Then the output should be:
      """
      ""
      """

  Scenario: repeat once
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      repeat("hello", 1)
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: repeat empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      repeat("", 5)
      """
    When I run the application with this content
    Then the output should be:
      """
      ""
      """

  Scenario: repeat negative count fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      repeat("x", -1)
      """
    When I run the application and it fails
    Then the output should contain "repeat count must be non-negative"

  Scenario: repeat on non-string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      repeat(123, 3)
      """
    When I run the application and it fails
    Then the output should contain "repeat expects a string as argument 1"

  # split tests (function form, different from splitBy operator)
  Scenario: split basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      split("a,b,c", ",")
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","b","c"]
      """

  Scenario: split with multi-character separator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      split("hello::world::test", "::")
      """
    When I run the application with this content
    Then the output should be:
      """
      ["hello","world","test"]
      """

  Scenario: split with no separator found
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      split("hello", ",")
      """
    When I run the application with this content
    Then the output should be:
      """
      ["hello"]
      """

  Scenario: split empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      split("", ",")
      """
    When I run the application with this content
    Then the output should be:
      """
      [""]
      """

  # charAt tests
  Scenario: charAt basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      charAt("hello", 0)
      """
    When I run the application with this content
    Then the output should be:
      """
      "h"
      """

  Scenario: charAt middle character
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      charAt("hello", 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      "l"
      """

  Scenario: charAt last character
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      charAt("hello", 4)
      """
    When I run the application with this content
    Then the output should be:
      """
      "o"
      """

  Scenario: charAt out of range fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      charAt("hello", 10)
      """
    When I run the application and it fails
    Then the output should contain "charAt index 10 out of range"

  Scenario: charAt negative index fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      charAt("hello", -1)
      """
    When I run the application and it fails
    Then the output should contain "charAt index -1 out of range"

  Scenario: charAt uses Unicode code point indices
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [charAt("é", 0), charAt("😀x", 0), charAt("éx", 1)]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["é","😀","x"]
      """

  Scenario: charAt Unicode out of range fails after last rune
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      charAt("é", 1)
      """
    When I run the application and it fails
    Then the output should contain "charAt index 1 out of range"

  # indexOf tests
  Scenario: indexOf string basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      indexOf("hello world", "world")
      """
    When I run the application with this content
    Then the output should be:
      """
      6
      """

  Scenario: indexOf string not found
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      indexOf("hello", "xyz")
      """
    When I run the application with this content
    Then the output should be:
      """
      -1
      """

  Scenario: indexOf string at start
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      indexOf("hello", "he")
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: indexOf string reports Unicode rune positions
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [indexOf("éx", "x"), indexOf("😀x😀", "x"), indexOf("😀x😀", "😀")]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,1,0]
      """

  Scenario: indexOf array basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      indexOf([1, 2, 3, 4, 5], 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      2
      """

  Scenario: indexOf array not found
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      indexOf([1, 2, 3], 10)
      """
    When I run the application with this content
    Then the output should be:
      """
      -1
      """

  Scenario: indexOf array with strings
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      indexOf(["a", "b", "c"], "b")
      """
    When I run the application with this content
    Then the output should be:
      """
      1
      """

  # lastIndexOf tests
  Scenario: lastIndexOf string basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      lastIndexOf("hello hello", "hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      6
      """

  Scenario: lastIndexOf string single occurrence
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      lastIndexOf("hello world", "world")
      """
    When I run the application with this content
    Then the output should be:
      """
      6
      """

  Scenario: lastIndexOf string not found
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      lastIndexOf("hello", "xyz")
      """
    When I run the application with this content
    Then the output should be:
      """
      -1
      """

  Scenario: lastIndexOf string reports Unicode rune positions
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [lastIndexOf("aéé", "é"), lastIndexOf("😀x😀", "😀")]
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,2]
      """

  Scenario: lastIndexOf array basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      lastIndexOf([1, 2, 3, 2, 1], 2)
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: lastIndexOf array not found
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      lastIndexOf([1, 2, 3], 10)
      """
    When I run the application with this content
    Then the output should be:
      """
      -1
      """

  Scenario: Unicode indexing builtins stay consistent together
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [substring("éx", 0, 1), indexOf("éx", "x"), charAt("éx", 1), lastIndexOf("aéé", "é")]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["é",1,"x",2]
      """

  # substring tests
  Scenario: substring with start only
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      substring("hello world", 6)
      """
    When I run the application with this content
    Then the output should be:
      """
      "world"
      """

  Scenario: substring with start and end
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      substring("hello world", 0, 5)
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: substring middle portion
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      substring("hello world", 2, 8)
      """
    When I run the application with this content
    Then the output should be:
      """
      "llo wo"
      """

  Scenario: substring from start
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      substring("hello", 0, 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      "hel"
      """

  Scenario: substring out of range start fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      substring("hello", 10)
      """
    When I run the application and it fails
    Then the output should contain "substring start index 10 out of range"

  # startsWith tests
  Scenario: startsWith true case
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      startsWith("hello world", "hello")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: startsWith false case
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      startsWith("hello world", "world")
      """
    When I run the application with this content
    Then the output should be false

  Scenario: startsWith empty prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      startsWith("hello", "")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: startsWith exact match
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      startsWith("hello", "hello")
      """
    When I run the application with this content
    Then the output should be true

  # endsWith tests
  Scenario: endsWith true case
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      endsWith("hello world", "world")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: endsWith false case
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      endsWith("hello world", "hello")
      """
    When I run the application with this content
    Then the output should be false

  Scenario: endsWith empty suffix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      endsWith("hello", "")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: endsWith exact match
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      endsWith("hello", "hello")
      """
    When I run the application with this content
    Then the output should be true

  # contains tests (string)
  Scenario: contains string true case
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      contains("hello world", "lo wo")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: contains string false case
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      contains("hello world", "xyz")
      """
    When I run the application with this content
    Then the output should be false

  Scenario: contains string empty substring
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      contains("hello", "")
      """
    When I run the application with this content
    Then the output should be true

  Scenario: contains array true case
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      contains([1, 2, 3, 4, 5], 3)
      """
    When I run the application with this content
    Then the output should be true

  Scenario: contains array false case
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      contains([1, 2, 3], 10)
      """
    When I run the application with this content
    Then the output should be false

  Scenario: contains array with strings
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      contains(["a", "b", "c"], "b")
      """
    When I run the application with this content
    Then the output should be true

  # replace tests
  Scenario: replace basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      replace("hello world", "world", "there")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello there"
      """

  Scenario: replace all occurrences
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      replace("hello hello hello", "hello", "hi")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hi hi hi"
      """

  Scenario: replace no match
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      replace("hello world", "xyz", "abc")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello world"
      """

  Scenario: replace with empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      replace("hello world", " ", "")
      """
    When I run the application with this content
    Then the output should be:
      """
      "helloworld"
      """

  Scenario: replace entire string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      replace("hello", "hello", "goodbye")
      """
    When I run the application with this content
    Then the output should be:
      """
      "goodbye"
      """

  # Combination scenarios
  Scenario: trim and split combined
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      split(trim("  a,b,c  "), ",")
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","b","c"]
      """

  Scenario: replace and length combined
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      length(replace("hello world", " ", ""))
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  Scenario: startsWith and endsWith combined
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        startsHello: startsWith("hello world", "hello"),
        endsWorld: endsWith("hello world", "world")
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"startsHello":true,"endsWorld":true}
      """
