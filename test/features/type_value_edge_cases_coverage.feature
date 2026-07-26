Feature: Type and value edge case coverage
  In order to measure coverage for types.go and types_value.go
  As a maintainer
  I want type coercion edge cases exercised in-process

  # --- matchesTypeExactly for Array, Object, Function, Regex via union types ---

  Scenario: Array value matches Array in union without coercion
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] as Array | String
      """
    When I run the script
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: Object value matches Object in union without coercion
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {x: 1} as Object | String
      """
    When I run the script
    Then the output should be:
      """
      {"x":1}
      """

  Scenario: Lambda value matches Function in union without coercion
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(((x) -> x) as Function | String)
      """
    When I run the script
    Then the output should be:
      """
      "Function"
      """

  Scenario: Regex value matches Regex in union without coercion
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(/abc/ as Regex | String)
      """
    When I run the script
    Then the output should be:
      """
      "Regex"
      """

  # --- coerceToNamespace error paths ---

  Scenario: Namespace coercion from object without uri fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {prefix: "ns"} as Namespace
      """
    Then running the script should fail with error containing "uri"

  Scenario: Namespace coercion with empty uri string fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {uri: ""} as Namespace
      """
    Then running the script should fail with error containing "uri"

  Scenario: Namespace coercion with non-string prefix fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {uri: "http://example.com", prefix: 42} as Namespace
      """
    Then running the script should fail with error containing "prefix"

  Scenario: Namespace coercion from non-object non-namespace type fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      42 as Namespace
      """
    Then running the script should fail with error containing "cannot coerce"

  # --- Cyclic type detection ---

  Scenario: Cyclic custom type definition is detected and fails
    Given the following script:
      """
      %im 0.1
      type A = A
      output application/json
      ---
      1 as A
      """
    Then running the script should fail with error containing "cyclic"

  # --- mergeProperties with chained custom types ---

  Scenario: Chained custom types merge format properties
    Given the following script:
      """
      %im 0.1
      type Amount = String { format: "##.00" }
      type Price = Amount { locale: "en" }
      output application/json
      ---
      100 as Price
      """
    When I run the script
    Then the output should be:
      """
      "100.00"
      """

  # --- matchesTypeExactly for optional types ---

  Scenario: Optional type matches null exactly without coercion
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      null as String?
      """
    When I run the script
    Then the output should be null

  Scenario: Optional type matches non-null value exactly
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "hello" as String?
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  # --- coerceToUnion error with no matching type ---

  Scenario: Union coercion fails when no member matches
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      ((x) -> x) as Number | Boolean
      """
    Then running the script should fail with error containing "cannot coerce to any of union types"

  # --- IsNumeric and KindOf coverage ---

  Scenario: typeOf for integer value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(1)
      """
    When I run the script
    Then the output should be:
      """
      "Number"
      """

  Scenario: typeOf for float value
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(1.5)
      """
    When I run the script
    Then the output should be:
      """
      "Number"
      """

  Scenario: typeOf for regex literal
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(/hello/)
      """
    When I run the script
    Then the output should be:
      """
      "Regex"
      """

  Scenario: typeOf for lambda
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf((x) -> x + 1)
      """
    When I run the script
    Then the output should be:
      """
      "Function"
      """

  Scenario: typeOf for null
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(null)
      """
    When I run the script
    Then the output should be:
      """
      "Null"
      """

  Scenario: typeOf returns a Type value rather than a String
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [typeOf(typeOf(false)), typeOf(false) == "Boolean", typeOf(false) as String]
      """
    When I run the script
    Then the output should be:
      """
      ["Type",false,"Boolean"]
      """

  # --- splitUnion edge cases ---

  Scenario: Union with whitespace around pipe is handled correctly
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      42 as String | Number | Null
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  # --- ParseNumericLiteral coverage ---

  Scenario: Numeric coercion from integer string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "999" as Number
      """
    When I run the script
    Then the output should be:
      """
      999
      """

  Scenario: ParseBoolLiteral case insensitive
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "TRUE" as Boolean
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: ParseBoolLiteral false case
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      "FALSE" as Boolean
      """
    When I run the script
    Then the output should be:
      """
      false
      """
