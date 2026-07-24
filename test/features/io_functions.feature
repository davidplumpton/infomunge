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
    Then the output should be:
      """
      {"name":"Alice","age":30}
      """

  Scenario: read CSV content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("name,age\nAlice,30\nBob,25", "application/csv")
      """
    When I run the script
    Then the output should be:
      """
      [{"name":"Alice","age":"30"},{"name":"Bob","age":"25"}]
      """

  Scenario: read YAML content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("name: Alice\nage: 30", "application/yaml")
      """
    When I run the script
    Then the output should be:
      """
      {"name":"Alice","age":30}
      """

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

  Scenario: read flatfile content with structured schema
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("0001ALICE   000030NY\n0002BOB     000025CA", "application/flatfile", {schema: {fields: [{name: "id", length: 4, type: "int", align: "right", pad: "0"}, {name: "name", length: 8, align: "left", pad: " "}, {name: "age", length: 6, type: "int", align: "right", pad: "0"}, {name: "state", length: 2}]}})
      """
    When I run the script
    Then the output should be:
      """
      [{"age":30,"id":1,"name":"ALICE","state":"NY"},{"age":25,"id":2,"name":"BOB","state":"CA"}]
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

  Scenario: read structured java envelope
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("{\"@class\":\"java.util.LinkedHashMap\",\"value\":{\"name\":\"Alice\",\"age\":30}}", "application/java", {structured: true})
      """
    When I run the script
    Then the output should be:
      """
      {"name":"Alice","age":30}
      """

  Scenario: read structured java envelope with unsupported class fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("{\"@class\":\"com.example.CustomType\",\"value\":{\"name\":\"Alice\"}}", "application/java", {structured: true})
      """
    Then running the script should fail with error containing "unsupported java class"

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

  Scenario: read structured protobuf payload with schema
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({name: "Alice", age: 30, active: true}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "string"}, {number: 2, name: "age", type: "int32"}, {number: 3, name: "active", type: "bool"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "string"}, {number: 2, name: "age", type: "int32"}, {number: 3, name: "active", type: "bool"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"active":true,"age":30,"name":"Alice"}
      """

  Scenario: read structured protobuf payload with unknown field fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({name: "Alice", age: 30}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "string"}, {number: 2, name: "age", type: "int32"}]}}) ++ write({extra: 1}, "application/protobuf", {structured: true, schema: {fields: [{number: 9, name: "extra", type: "int32"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "string"}, {number: 2, name: "age", type: "int32"}]}})
      """
    Then running the script should fail with error containing "unknown field number"

  Scenario: read structured protobuf payload with descriptor set
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({name: "Bob", lucky_numbers: [1, 150]}, "application/protobuf", {structured: true, descriptor: {set: "\u000a]\u000a\fperson.proto\u0012\u0004test\"?\u000a\u0006Person\u0012\f\u000a\u0004name\u0018\u0001 \u0001(\t\u0012'\u000a\rlucky_numbers\u0018\u0002 \u0003(\u0005B\u0002\u0010\u0001R\fluckyNumbersb\u0006proto3", message: "test.Person"}}), "application/protobuf", {structured: true, descriptor: {set: "\u000a]\u000a\fperson.proto\u0012\u0004test\"?\u000a\u0006Person\u0012\f\u000a\u0004name\u0018\u0001 \u0001(\t\u0012'\u000a\rlucky_numbers\u0018\u0002 \u0003(\u0005B\u0002\u0010\u0001R\fluckyNumbersb\u0006proto3", message: "test.Person"}})
      """
    When I run the script
    Then the output should be:
      """
      {"lucky_numbers":[1,150],"name":"Bob"}
      """

  Scenario: read and write structured protobuf map payload with descriptor set
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({kv: {"1": 10, "2": 20}}, "application/protobuf", {structured: true, descriptor: {set: "\u000a^\u000a\u0007m.proto\u0012\u0001t\"H\u000a\u0001M\u0012\u0018\u000a\u0002kv\u0018\u0001 \u0003(\u000b2\f.t.M.KvEntry\u001a)\u000a\u0007KvEntry\u0012\u000b\u000a\u0003key\u0018\u0001 \u0001(\u0005\u0012\r\u000a\u0005value\u0018\u0002 \u0001(\u0005:\u00028\u0001b\u0006proto3", message: "t.M"}}), "application/protobuf", {structured: true, descriptor: {set: "\u000a^\u000a\u0007m.proto\u0012\u0001t\"H\u000a\u0001M\u0012\u0018\u000a\u0002kv\u0018\u0001 \u0003(\u000b2\f.t.M.KvEntry\u001a)\u000a\u0007KvEntry\u0012\u000b\u000a\u0003key\u0018\u0001 \u0001(\u0005\u0012\r\u000a\u0005value\u0018\u0002 \u0001(\u0005:\u00028\u0001b\u0006proto3", message: "t.M"}})
      """
    When I run the script
    Then the output should be:
      """
      {"kv":{"1":10,"2":20}}
      """

  Scenario: write structured protobuf map with descriptor set rejects invalid key type
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({kv: {"abc": 10}}, "application/protobuf", {structured: true, descriptor: {set: "\u000a^\u000a\u0007m.proto\u0012\u0001t\"H\u000a\u0001M\u0012\u0018\u000a\u0002kv\u0018\u0001 \u0003(\u000b2\f.t.M.KvEntry\u001a)\u000a\u0007KvEntry\u0012\u000b\u000a\u0003key\u0018\u0001 \u0001(\u0005\u0012\r\u000a\u0005value\u0018\u0002 \u0001(\u0005:\u00028\u0001b\u0006proto3", message: "t.M"}})
      """
    Then running the script should fail with error containing "map key \"abc\" is invalid"

  Scenario: write structured protobuf map with descriptor set rejects invalid value type
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({kv: {"1": "ten"}}, "application/protobuf", {structured: true, descriptor: {set: "\u000a^\u000a\u0007m.proto\u0012\u0001t\"H\u000a\u0001M\u0012\u0018\u000a\u0002kv\u0018\u0001 \u0003(\u000b2\f.t.M.KvEntry\u001a)\u000a\u0007KvEntry\u0012\u000b\u000a\u0003key\u0018\u0001 \u0001(\u0005\u0012\r\u000a\u0005value\u0018\u0002 \u0001(\u0005:\u00028\u0001b\u0006proto3", message: "t.M"}})
      """
    Then running the script should fail with error containing "expects int32-compatible number"

  Scenario: read packed protobuf payload without packed schema hint
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({ids: [1, 2, 300]}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "ids", type: "int32", repeated: true, packed: true}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "ids", type: "int32", repeated: true}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"ids":[1,2,300]}
      """

  Scenario: read xlsx content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("xlsx-bytes", "application/xlsx")
      """
    When I run the script
    Then the output should be:
      """
      "xlsx-bytes"
      """

  Scenario: read xlsx standard mime alias content
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("xlsx-alias-bytes", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
      """
    When I run the script
    Then the output should be:
      """
      "xlsx-alias-bytes"
      """

  Scenario: read and write structured xlsx workbook
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({Employees: [{name: "Alice", age: 30, active: true}, {name: "Bob", age: 41, active: false}], Sales: [{region: "west", amount: 12.5}]}, "application/xlsx", {structured: true}), "application/xlsx", {structured: true})
      """
    When I run the script
    Then the output should be:
      """
      {"Employees":[{"active":true,"age":30,"name":"Alice"},{"active":false,"age":41,"name":"Bob"}],"Sales":[{"amount":12.5,"region":"west"}]}
      """

  Scenario: read and write structured xlsx workbook through standard mime alias
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({Sheet1: [{name: "Alice", active: true}]}, "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", {structured: true}), "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet", {structured: true})
      """
    When I run the script
    Then the output should be:
      """
      {"Sheet1":[{"active":true,"name":"Alice"}]}
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
    Then the output should contain "name,age"
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

  Scenario: write object array to flatfile with structured schema
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write([{id: 1, name: "ALICE", age: 30, state: "NY"}, {id: 2, name: "BOB", age: 25, state: "CA"}], "application/flatfile", {schema: {fields: [{name: "id", length: 4, type: "int", align: "right", pad: "0"}, {name: "name", length: 8, align: "left", pad: " "}, {name: "age", length: 6, type: "int", align: "right", pad: "0"}, {name: "state", length: 2}]}})
      """
    When I run the script
    Then the output should be:
      """
      "0001ALICE   000030NY\n0002BOB     000025CA"
      """

  Scenario: read flatfile with malformed record length fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read("0001ALICE   000030NY\nBAD", "application/flatfile", {schema: {fields: [{name: "id", length: 4}, {name: "name", length: 8}, {name: "age", length: 6}, {name: "state", length: 2}]}})
      """
    Then running the script should fail with error containing "record 2 has length"

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

  Scenario: write structured java envelope
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice", age: 30}, "application/java", {structured: true})
      """
    When I run the script
    Then the output should be:
      """
      "{\"@class\":\"java.util.LinkedHashMap\",\"value\":{\"name\":\"Alice\",\"age\":30}}"
      """

  Scenario: write structured java envelope class mismatch fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/java", {structured: true, class: "java.util.List"})
      """
    Then running the script should fail with error containing "is incompatible with value type"

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

  Scenario: write structured protobuf with schema and round-trip decode
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({tags: ["a", "bb"]}, "application/x-protobuf", {structured: true, schema: {fields: [{number: 1, name: "tags", type: "string", repeated: true}]}}), "application/x-protobuf", {structured: true, schema: {fields: [{number: 1, name: "tags", type: "string", repeated: true}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"tags":["a","bb"]}
      """

  Scenario: write structured protobuf with unknown field fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice", extra: "x"}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "string"}]}})
      """
    Then running the script should fail with error containing "unknown field"

  Scenario: write string to xlsx
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("xlsx-bytes", "application/xlsx")
      """
    When I run the script
    Then the output should be:
      """
      "xlsx-bytes"
      """

  Scenario: write string to xlsx standard mime alias
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write("xlsx-alias-bytes", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
      """
    When I run the script
    Then the output should be:
      """
      "xlsx-alias-bytes"
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

  Scenario: readUrl reports disabled URL IO capability
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://example.com/data.json", "application/json")
      """
    When I run the script with URL IO disabled
    Then the output should contain "URL IO capability is disabled"

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

  Scenario: readUrl blocks hostnames that resolve to private addresses
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://localhost/secret", "text/plain")
      """
    Then running the script should fail with error containing "resolves to private address"

  Scenario: readUrl blocks link-local metadata addresses
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://169.254.169.254/latest/meta-data/", "text/plain")
      """
    Then running the script should fail with error containing "private/internal address"

  Scenario: readUrl blocks redirects to private targets
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      readUrl("http://redirect.example/start", "text/plain")
      """
    When I run the script with readUrl redirecting to "http://127.0.0.1/secret"
    Then the output should contain "private/internal address"

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
      {"name":"Alice","scores":[1,2,3],"active":true}
      """
