Feature: Protobuf structured field type coverage
  In order to measure coverage for protobuf_structured.go
  As a maintainer
  I want protobuf field types and option error paths exercised in-process

  # --- sint32 (zigzag encoding) ---

  Scenario: sint32 round-trip with negative value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: -1}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sint32"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sint32"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":-1}
      """

  Scenario: sint32 round-trip with positive value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: 42}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sint32"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sint32"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":42}
      """

  # --- sint64 (zigzag encoding) ---

  Scenario: sint64 round-trip with negative value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: -100}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sint64"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sint64"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":-100}
      """

  # --- double ---

  Scenario: double round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: 1.5}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "double"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "double"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":1.5}
      """

  # --- float ---

  Scenario: float round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: 2.0}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "float"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "float"}]}})
      """
    When I run the script
    Then the output should contain "val"

  # --- uint32 ---

  Scenario: uint32 round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: 300}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "uint32"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "uint32"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":300}
      """

  # --- uint64 ---

  Scenario: uint64 round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: 1000}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "uint64"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "uint64"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":1000}
      """

  # --- sfixed32 ---

  Scenario: sfixed32 round-trip with negative value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: -5}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sfixed32"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sfixed32"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":-5}
      """

  # --- sfixed64 ---

  Scenario: sfixed64 round-trip with negative value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: -9}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sfixed64"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "sfixed64"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":-9}
      """

  # --- fixed32 ---

  Scenario: fixed32 round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: 7}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "fixed32"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "fixed32"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":7}
      """

  # --- fixed64 ---

  Scenario: fixed64 round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({val: 99}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "fixed64"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "val", type: "fixed64"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"val":99}
      """

  # --- enum ---

  Scenario: enum round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({status: 2}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "status", type: "enum"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "status", type: "enum"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"status":2}
      """

  # --- bytes ---

  Scenario: bytes round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({data: "hello"}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "data", type: "bytes"}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "data", type: "bytes"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"data":"hello"}
      """

  # --- nested message ---

  Scenario: nested message round-trip
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({person: {name: "Bob", age: 25}}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "person", type: "message", schema: {fields: [{number: 1, name: "name", type: "string"}, {number: 2, name: "age", type: "int32"}]}}]}}), "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "person", type: "message", schema: {fields: [{number: 1, name: "name", type: "string"}, {number: 2, name: "age", type: "int32"}]}}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"person":{"age":25,"name":"Bob"}}
      """

  # --- non-strict mode ---

  Scenario: non-strict read silently skips unknown fields
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      read(write({name: "Alice", extra: 99}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "string"}, {number: 5, name: "extra", type: "int32"}]}}), "application/protobuf", {structured: true, strict: false, schema: {fields: [{number: 1, name: "name", type: "string"}]}})
      """
    When I run the script
    Then the output should be:
      """
      {"name":"Alice"}
      """

  Scenario: non-strict write allows unknown object fields
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(write({name: "Alice", ignored: "extra"}, "application/protobuf", {structured: true, strict: false, schema: {fields: [{number: 1, name: "name", type: "string"}]}})) > 0
      """
    When I run the script
    Then the output should be true

  # --- option validation errors ---

  Scenario: schema option without structured=true fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {schema: {fields: [{number: 1, name: "name", type: "string"}]}})
      """
    Then running the script should fail with error containing "schema requires structured=true"

  Scenario: descriptor option without structured=true fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {descriptor: "bytes"})
      """
    Then running the script should fail with error containing "descriptor requires structured=true"

  Scenario: structured=true without schema or descriptor fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {structured: true})
      """
    Then running the script should fail with error containing "requires schema or descriptor"

  Scenario: schema and descriptor mutually exclusive
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "string"}]}, descriptorSet: "bytes"})
      """
    Then running the script should fail with error containing "mutually exclusive"

  Scenario: structured option must be a boolean
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {structured: "yes"})
      """
    Then running the script should fail with error containing "structured must be a boolean"

  Scenario: unsupported protobuf option fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {unknownOpt: true})
      """
    Then running the script should fail with error containing "unsupported protobuf option"

  Scenario: schema fields must be non-empty array
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {structured: true, schema: {fields: []}})
      """
    Then running the script should fail with error containing "fields must be a non-empty array"

  Scenario: schema field type must be supported
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "badtype"}]}})
      """
    Then running the script should fail with error containing "unsupported type"

  Scenario: schema field number must be positive
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({name: "Alice"}, "application/protobuf", {structured: true, schema: {fields: [{number: 0, name: "name", type: "string"}]}})
      """
    Then running the script should fail with error containing "positive integer"

  Scenario: schema duplicate field number fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({a: "x", b: "y"}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "a", type: "string"}, {number: 1, name: "b", type: "string"}]}})
      """
    Then running the script should fail with error containing "duplicate field number"

  Scenario: message field without nested schema fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      write({inner: {x: 1}}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "inner", type: "message"}]}})
      """
    Then running the script should fail with error containing "requires nested schema"
