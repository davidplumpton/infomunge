Feature: I/O Functions
  In order to read and write data in various formats
  As a developer
  I want to use the read, write, and readUrl functions

  # Read function tests
  Scenario: read basic JSON
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("{\"name\": \"Alice\", \"age\": 30}", "application/json")
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should contain "30"

  Scenario: read CSV content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("name,age\nAlice,30\nBob,25", "application/csv")
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should contain "30"

  Scenario: read YAML content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("name: Alice\nage: 30", "application/yaml")
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should contain "30"

  Scenario: read XML content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("<root><name>Alice</name><age>30</age></root>", "application/xml")
      """
    When I run the script
    Then the output should be:
      """
      {"root":{"age":"30","name":"Alice"}}
      """

  Scenario: read binary octet-stream content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("hello-binary", "application/octet-stream")
      """
    When I run the script
    Then the output should be:
      """
      "hello-binary"
      """

  Scenario: read avro content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("hello-avro", "application/avro")
      """
    When I run the script
    Then the output should be:
      """
      "hello-avro"
      """

  Scenario: read dw content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("%dw 2.0\noutput application/json\n---\n42", "application/dw")
      """
    When I run the script
    Then the output should be:
      """
      "%dw 2.0\noutput application/json\n---\n42"
      """

  Scenario: read flatfile content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("HDR0001ALICE   000030NY\nDTL0002BOB     000025CA", "application/flatfile")
      """
    When I run the script
    Then the output should be:
      """
      "HDR0001ALICE   000030NY\nDTL0002BOB     000025CA"
      """

  Scenario: read java content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("serialized-java-object", "application/java")
      """
    When I run the script
    Then the output should be:
      """
      "serialized-java-object"
      """

  Scenario: read protobuf content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("protobuf-bytes", "application/protobuf")
      """
    When I run the script
    Then the output should be:
      """
      "protobuf-bytes"
      """

  Scenario: read x-protobuf content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("x-protobuf-bytes", "application/x-protobuf")
      """
    When I run the script
    Then the output should be:
      """
      "x-protobuf-bytes"
      """

  Scenario: read with invalid JSON fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("{invalid json", "application/json")
      """
    Then running the script should fail with error containing "invalid"

  Scenario: read with unsupported mime type fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("some content", "application/unsupported")
      """
    Then running the script should fail with error containing "unsupported"

  Scenario: read requires two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("{\"a\": 1}")
      """
    Then running the script should fail with error containing "requires at least 2 arguments"

  Scenario: read with non-string content fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(123, "application/json")
      """
    Then running the script should fail with error containing "must be strings"

  Scenario: read with non-string mime type fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("{\"a\": 1}", 123)
      """
    Then running the script should fail with error containing "must be strings"

  Scenario: read empty JSON object
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("{}", "application/json")
      """
    When I run the script
    Then the output should be:
      """
      {}
      """

  Scenario: read empty JSON array
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("[]", "application/json")
      """
    When I run the script
    Then the output should be:
      """
      []
      """

  Scenario: read JSON with nested structures
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("{\"user\": {\"name\": \"Alice\", \"address\": {\"city\": \"NYC\"}}}", "application/json")
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should contain "NYC"

  # Write function tests
  Scenario: write object to JSON
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({"name": "Alice", "age": 30}, "application/json")
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should contain "age"

  Scenario: write array to JSON
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write([1, 2, 3], "application/json")
      """
    When I run the script
    Then the output should be:
      """
      "[1,2,3]"
      """

  Scenario: write object to CSV
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write([{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}], "application/csv")
      """
    When I run the script
    Then the output should contain "age,name"
    And the output should contain "Alice"
    And the output should contain "Bob"

  Scenario: write object to YAML
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({"name": "Alice", "age": 30}, "application/yaml")
      """
    When I run the script
    Then the output should contain "name"
    And the output should contain "Alice"

  Scenario: write object to XML
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({"root": {"name": "Alice"}}, "application/xml")
      """
    When I run the script
    Then the output should contain "Alice"

  Scenario: write string to binary octet-stream
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("abc-xyz", "application/octet-stream")
      """
    When I run the script
    Then the output should be:
      """
      "abc-xyz"
      """

  Scenario: write string to avro
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("avro-xyz", "application/avro")
      """
    When I run the script
    Then the output should be:
      """
      "avro-xyz"
      """

  Scenario: write string to dw
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("%dw 2.0\noutput application/json\n---\n{a: 1}", "application/dw")
      """
    When I run the script
    Then the output should be:
      """
      "%dw 2.0\noutput application/json\n---\n{a: 1}"
      """

  Scenario: write string to flatfile
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("HDR0001ALICE   000030NY\nDTL0002BOB     000025CA", "application/flatfile")
      """
    When I run the script
    Then the output should be:
      """
      "HDR0001ALICE   000030NY\nDTL0002BOB     000025CA"
      """

  Scenario: write string to java
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("serialized-java-object", "application/java")
      """
    When I run the script
    Then the output should be:
      """
      "serialized-java-object"
      """

  Scenario: write string to protobuf
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("protobuf-bytes", "application/protobuf")
      """
    When I run the script
    Then the output should be:
      """
      "protobuf-bytes"
      """

  Scenario: write string to x-protobuf
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("x-protobuf-bytes", "application/x-protobuf")
      """
    When I run the script
    Then the output should be:
      """
      "x-protobuf-bytes"
      """

  Scenario: write requires exactly two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({"a": 1})
      """
    Then running the script should fail with error containing "requires exactly 2 arguments"

  Scenario: write with non-string mime type fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({"a": 1}, 123)
      """
    Then running the script should fail with error containing "mimeType to be a string"

  Scenario: write null value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write(null, "application/json")
      """
    When I run the script
    Then the output should contain "null"

  Scenario: write string to JSON
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("hello", "application/json")
      """
    When I run the script
    Then the output should contain "hello"

  Scenario: write number to JSON
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write(42, "application/json")
      """
    When I run the script
    Then the output should be:
      """
      "42"
      """

  # Round-trip tests (read then write)
  Scenario: read JSON then write back to JSON
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write(read("{\"x\": 1}", "application/json"), "application/json")
      """
    When I run the script
    Then the output should contain "x"

  Scenario: read CSV then write to JSON
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write(read("name,age\nAlice,30", "application/csv"), "application/json")
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should contain "30"

  Scenario: read JSON then write to CSV
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write(read("[{\"name\": \"Alice\", \"age\": 30}]", "application/json"), "application/csv")
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should contain "name"
    And the output should contain "age"

  # ReadUrl function tests - Note: actual HTTP testing would require network access
  # These tests verify the error handling instead
  Scenario: readUrl with invalid URL fails gracefully
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://invalid-domain-that-does-not-exist-12345.local", "application/json")
      """
    Then running the script should fail with error containing "readUrl"

  Scenario: readUrl respects canceled evaluation context
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://1.1.1.1", "application/json")
      """
    When I run the script with a canceled evaluation context
    Then the output should contain "context canceled"

  Scenario: readUrl respects expired evaluation deadline
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://1.1.1.1", "application/json")
      """
    When I run the script with an expired evaluation deadline
    Then the output should contain "context deadline exceeded"

  Scenario: readUrl requires two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://example.com")
      """
    Then running the script should fail with error containing "requires exactly 2 arguments"

  Scenario: readUrl with non-string URL fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl(123, "application/json")
      """
    Then running the script should fail with error containing "expects url to be a string"

  Scenario: readUrl with non-string mime type fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://example.com", 123)
      """
    Then running the script should fail with error containing "expects mimeType to be a string"

  Scenario: readUrl blocks private IP addresses
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://127.0.0.1/secret", "application/json")
      """
    Then running the script should fail with error containing "private/internal address"

  Scenario: readUrl blocks link-local metadata addresses
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://169.254.169.254/latest/meta-data/", "text/plain")
      """
    Then running the script should fail with error containing "private/internal address"

  Scenario: readUrl blocks non-HTTP schemes
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("file:///etc/passwd", "text/plain")
      """
    Then running the script should fail with error containing "unsupported scheme"

  Scenario: readUrl blocks ftp scheme
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("ftp://example.com/file", "text/plain")
      """
    Then running the script should fail with error containing "unsupported scheme"

  # Complex scenarios combining read and write
  Scenario: transform CSV to JSON
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("name,score\nAlice,95\nBob,87", "application/csv") map (x) -> {name: x.name, score: x.score}
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should contain "95"

  Scenario: read and filter JSON array
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("[{\"name\": \"Alice\", \"age\": 30}, {\"name\": \"Bob\", \"age\": 25}]", "application/json") filter (x) -> x.age > 26
      """
    When I run the script
    Then the output should contain "Alice"
    And the output should not contain "Bob"

  Scenario: write formatted output with indentation
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({"a": {"b": 1}}, "application/json")
      """
    When I run the script
    Then the output should be:
      """
      "{\"a\":{\"b\":1}}"
      """

  Scenario: JSON read/write round trip preserves structure
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({"name": "Alice", "scores": [1, 2, 3], "active": true}, "application/json"), "application/json")
      """
    When I run the script
    Then the output should be:
      """
      {"active":true,"name":"Alice","scores":[1,2,3]}
      """
