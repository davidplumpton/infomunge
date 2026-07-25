Feature: String Functions - Unicode and Special Character Edge Cases
  In order to handle real-world string data
  As a developer
  I want string functions to handle Unicode, special characters, and edge cases correctly

  # Unicode and Emoji Tests
  Scenario: Unicode emoji handling
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "Hello 👋 World 🌍"
      """
    When I run the script
    Then the output should be:
      """
      "Hello 👋 World 🌍"
      """

  Scenario: Unicode emoji in upper/lower functions
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      upper("hello 👋 world")
      """
    When I run the script
    Then the output should be:
      """
      "HELLO 👋 WORLD"
      """

  Scenario: Unicode emoji length
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "👋" | length
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  Scenario: Non-Latin script (Greek)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      upper("hello αβγδ world")
      """
    When I run the script
    Then the output should be:
      """
      "HELLO ΑΒΓΔ WORLD"
      """

  Scenario: Non-Latin script (Cyrillic)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      lower("ПРИВЕТ ЖЕ")
      """
    When I run the script
    Then the output should be:
      """
      "привет же"
      """

  Scenario: Non-Latin script (Arabic)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "مرحبا" | length
      """
    When I run the script
    Then the output should be:
      """
      5
      """

  Scenario: Non-Latin script (CJK - Chinese)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "你好世界" | length
      """
    When I run the script
    Then the output should be:
      """
      4
      """

  Scenario: Non-Latin script (Devanagari - Hindi)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "नमस्ते" | length
      """
    When I run the script
    Then the output should be:
      """
      6
      """

  # Combining Characters and Diacritics
  Scenario: Combining diacritical marks (accents)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "café" | length
      """
    When I run the script
    Then the output should be:
      """
      4
      """

  Scenario: Multiple combining marks
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "ñoño" | length
      """
    When I run the script
    Then the output should be:
      """
      4
      """

  # Escape Sequences
  Scenario: Newline escape sequence
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "line1\nline2"
      """
    When I run the script
    Then the output should be:
      """
      "line1\nline2"
      """

  Scenario: Tab escape sequence
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "col1\tcol2"
      """
    When I run the script
    Then the output should be:
      """
      "col1\tcol2"
      """

  Scenario: Carriage return escape sequence
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "line1\rline2" | length
      """
    When I run the script
    Then the output should be:
      """
      11
      """

  Scenario: Backslash escape sequence
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "path\\to\\file"
      """
    When I run the script
    Then the output should be:
      """
      "path\\to\\file"
      """

  Scenario: String literal containing if(
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "if("
      """
    When I run the script
    Then the output should be:
      """
      "if("
      """

  Scenario: Quote escape sequence
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "He said \"Hello\""
      """
    When I run the script
    Then the output should be:
      """
      "He said \"Hello\""
      """

  # Regex with Special Characters
  Scenario: Regex with special regex characters (dot)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "test.txt" matches "test\\.txt"
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Regex with special regex characters (asterisk)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "aaaa" matches "a*"
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Regex character class rejects a trailing non-match
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "123abc" matches "[0-9]+"
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: Negated regex character class rejects a trailing non-match
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "abc123" matches "[^0-9]+"
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  # String with Whitespace Variations
  Scenario: String with multiple spaces
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello    world" | trim
      """
    When I run the script
    Then the output should be:
      """
      "hello    world"
      """

  Scenario: String with leading whitespace
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "   hello" | trim
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  Scenario: String with trailing whitespace
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello   " | trim
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  Scenario: String with mixed whitespace types
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      " \t hello \n " | trim
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  # Very Long Strings
  Scenario: Very long string (1000+ characters)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ("a" repeat 1000) | length
      """
    When I run the script
    Then the output should be:
      """
      1000
      """

  Scenario: Very long Unicode string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ("你好" repeat 500) | length
      """
    When I run the script
    Then the output should be:
      """
      1000
      """

  # Null bytes and Binary-like data
  Scenario: String with null byte handling
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello\u0000world" | length
      """
    When I run the script
    Then the output should be:
      """
      11
      """

  # Empty and Minimal Strings
  Scenario: Empty string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "" | length
      """
    When I run the script
    Then the output should be:
      """
      0
      """

  Scenario: Single character string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "a" | length
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  Scenario: Single emoji string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "👋" | length
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  # String Splitting with Special Characters
  Scenario: Split with regex special character
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "a.b.c" splitBy "."
      """
    When I run the script
    Then the output should be:
      """
      ["a","b","c"]
      """

  Scenario: Split with newline separator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "line1\nline2\nline3" splitBy "\n"
      """
    When I run the script
    Then the output should be:
      """
      ["line1","line2","line3"]
      """

  Scenario: Split with tab separator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "col1\tcol2\tcol3" splitBy "\t"
      """
    When I run the script
    Then the output should be:
      """
      ["col1","col2","col3"]
      """

  Scenario: Split with multi-character separator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "oneANDtwoANDthree" splitBy "AND"
      """
    When I run the script
    Then the output should be:
      """
      ["one","two","three"]
      """

  Scenario: Split with Unicode separator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "one→two→three" splitBy "→"
      """
    When I run the script
    Then the output should be:
      """
      ["one","two","three"]
      """

  # String Joining with Special Characters
  Scenario: Join with newline separator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ["line1", "line2", "line3"] joinBy "\n"
      """
    When I run the script
    Then the output should be:
      """
      "line1\nline2\nline3"
      """

  Scenario: Join with tab separator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ["col1", "col2", "col3"] joinBy "\t"
      """
    When I run the script
    Then the output should be:
      """
      "col1\tcol2\tcol3"
      """

  Scenario: Join with Unicode separator
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ["one", "two", "three"] joinBy "→"
      """
    When I run the script
    Then the output should be:
      """
      "one→two→three"
      """

  # Case-sensitive operations with Unicode
  Scenario: Unicode case sensitivity in matches
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "HELLO" matches ("hello" as String)
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: Case-insensitive matches with Unicode
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "HELLO" matches "hello" as String
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  # Edge cases with Substring operations
  Scenario: Substring with Unicode
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "你好世界" substring(0, 2)
      """
    When I run the script
    Then the output should be:
      """
      "你好"
      """

  Scenario: Substring with emoji
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "Hello👋World🌍" substring(5, 6)
      """
    When I run the script
    Then the output should be:
      """
      "👋"
      """

  Scenario: Substring operator basic extraction
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello world" substring(6)
      """
    When I run the script
    Then the output should be:
      """
      "world"
      """

  Scenario: Substring operator with start and end
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello world" substring(0, 5)
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  Scenario: Substring operator with negative start index fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" substring(-1, 2)
      """
    Then running the script should fail with error containing "substring start index -1 out of range"

  Scenario: Substring operator with end before start fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" substring(4, 2)
      """
    Then running the script should fail with error containing "substring end index 2 out of range"

  Scenario: Substring operator with end out of range fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" substring(1, 10)
      """
    Then running the script should fail with error containing "substring end index 10 out of range"

  Scenario: Substring operator on empty string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "" substring(0, 0)
      """
    When I run the script
    Then the output should be:
      """
      ""
      """

  Scenario: Substring operator combined with other string operations
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ("hello" ++ " world") substring(0, 5) ++ "!"
      """
    When I run the script
    Then the output should be:
      """
      "hello!"
      """
