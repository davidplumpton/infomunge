Feature: User function name resolution
  In order to extend transformations without collisions
  As an InfoMunge author
  I want user-defined functions to take precedence over builtins

  Scenario: Local functions shadow string and collection builtins
    Given the following script:
      """
      %im 0.1
      fun upper(value) = "custom:" ++ value
      fun sizeOf(value) = 99
      output application/json
      ---
      {text: upper("abc"), count: sizeOf([1, 2, 3])}
      """
    When I run the script
    Then the output should be:
      """
      {"text":"custom:abc","count":99}
      """

  Scenario: Imported functions shadow string and collection builtins
    Given a file named "modules/ShadowBuiltins.im" with content:
      """
      %im 0.1
      fun upper(value) = "module:" ++ value
      fun sizeOf(value) = 42
      """
    And the following input content:
      """
      %im 0.1
      import upper, sizeOf from modules::ShadowBuiltins
      output application/json
      ---
      {text: upper("abc"), count: sizeOf([1, 2, 3])}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"text":"module:abc","count":42}
      """

  Scenario: Local functions shadow special builtins
    Given the following script:
      """
      %im 0.1
      fun onNull(value, fallback) = "custom"
      output application/json
      ---
      onNull(null, "builtin")
      """
    When I run the script
    Then the output should be:
      """
      "custom"
      """

  Scenario: Core module wrappers explicitly call native implementations
    Given the following input content:
      """
      %im 0.1
      import drop, every from dw::core::Arrays
      output application/json
      ---
      {remaining: drop([1, 2, 3], 1), allPositive: every([1, 2, 3], (value) -> value > 0)}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"remaining":[2,3],"allPositive":true}
      """

  Scenario: Local functions cannot intercept generated syntax helpers
    Given the following script:
      """
      %im 0.1
      fun __default(value, fallback) = 99
      fun __lambda(parameters, body) = 99
      fun __map(values, mapper) = [99]
      fun __coerce(value, typeName) = "custom"
      fun fail() = 1 / 0
      output application/json
      ---
      {
        lazyDefault: 7 default fail(),
        fallback: null default 5,
        mapped: [1, 2] map (value) -> value + 1,
        coerced: "42" as Number
      }
      """
    When I run the script
    Then the output should be:
      """
      {"lazyDefault":7,"fallback":5,"mapped":[2,3],"coerced":42}
      """

  Scenario: Imported functions cannot intercept generated syntax helpers
    Given a file named "modules/ShadowSyntaxHelpers.im" with content:
      """
      %im 0.1
      fun __default(value, fallback) = 99
      fun __lambda(parameters, body) = 99
      fun __map(values, mapper) = [99]
      fun __coerce(value, typeName) = "custom"
      """
    And the following input content:
      """
      %im 0.1
      import __default, __lambda, __map, __coerce from modules::ShadowSyntaxHelpers
      output application/json
      ---
      {
        fallback: null default 5,
        mapped: [1, 2] map (value) -> value + 1,
        coerced: "42" as Number
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"fallback":5,"mapped":[2,3],"coerced":42}
      """
