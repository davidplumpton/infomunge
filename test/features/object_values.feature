Feature: Object Values
  In order to build complex data structures
  As a developer
  I want to use expressions and various data types as values within objects

  Scenario: Object with an expression as a value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      { sum: 1 + 2 }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"sum":3}
      """

  Scenario: Object with mixed value types
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        str: "hello",
        num: 42,
        bool: true,
        expression: 10 * 2
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"bool":true,"expression":20,"num":42,"str":"hello"}
      """

  Scenario: Nested objects with expressions
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       {
         outer: {
           inner: 5 + 5
         }
       }
       """
     When I run the application with this content
     Then the output should be:
       """
       {"outer":{"inner":10}}
       """

   Scenario: Object with concatenation operator without parentheses
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       { greeting: "hello" ++ " " ++ "world" }
       """
     When I run the application with this content
     Then the output should be:
       """
       {"greeting":"hello world"}
       """

   Scenario: Object with multiple complex expressions
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       {
         concat: "a" ++ "b",
         addition: 10 + 5,
         nested: { inner: "x" ++ "y" }
       }
       """
     When I run the application with this content
     Then the output should be:
       """
       {"addition":15,"concat":"ab","nested":{"inner":"xy"}}
       """

   Scenario: Multi-value selector on object (DataWeave .* syntax)
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       { name: "Alice", age: 30, city: "NYC" }.*
       """
     When I run the application with this content
     Then the output should be:
       """
       [30,"NYC","Alice"]
       """

   Scenario: Multi-value selector on array returns the array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       ["a", "b", "c"].*
       """
     When I run the application with this content
     Then the output should be:
       """
       ["a","b","c"]
       """

   Scenario: Multi-value selector on array of objects
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       [{a: 1}, {a: 2}].*a
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2]
       """

   Scenario: Multi-value selector on single object field
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       {a: 1}.*a
       """
     When I run the application with this content
     Then the output should be:
       """
       [1]
       """

   Scenario: Multi-value selector missing field returns empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       {b: 1}.*a
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: Multi-value selector on nested field access
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       [{a: {b: 1}}, {a: {b: 2}}].*a.b
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2]
       """

   Scenario: Multi-value selector chained with filter operator
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       [{a: 1}, {a: -1}, {a: 2}].*a filter $ > 0
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2]
       """

   Scenario: Multi-value selector on null returns empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       null.*a
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: Multi-value selector on mixed array of objects
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       [{a: 1}, {b: 2}].*a
       """
     When I run the application with this content
     Then the output should be:
       """
       [1]
       """

   Scenario: Multi-value selector on nested structures
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       {
         "object_vals": { x: 10, y: 20, z: 30 }.*,
         "array_vals": [1, 2, 3].*,
         "empty_object": {}.*
       }
       """
     When I run the application with this content
     Then the output should be:
       """
       {"array_vals":[1,2,3],"empty_object":[],"object_vals":[10,20,30]}
       """

  Scenario: Object merging with single object expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      { 
        "start": "ok",
        ({"merged": "value"}) 
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"merged":"value","start":"ok"}
      """

  Scenario: Object merging with array of objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      { 
        "base": 0,
        ([{"a": 1}, {"b": 2}])
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"a":1,"b":2,"base":0}
      """
