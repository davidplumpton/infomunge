Feature: Aggregation Functions
  In order to aggregate and analyze arrays
  As a developer
  I want to use the aggregation functions: avg, sum, max, min, maxBy, minBy, orderBy

  Scenario: sum basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sum([1, 2, 3, 4, 5])
      """
    When I run the application with this content
    Then the output should be:
      """
      15
      """

  Scenario: sum with floats
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sum([1.5, 2.5, 3.0])
      """
    When I run the application with this content
    Then the output should be:
      """
      7
      """

  Scenario: sum empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sum([])
      """
    When I run the application with this content
    Then the output should be:
      """
      0
      """

  Scenario: sum single element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sum([42])
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: sum with multiple numeric values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sum([10, 20, 30])
      """
    When I run the application with this content
    Then the output should be:
      """
      60
      """

  Scenario: avg basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      avg([1, 2, 3, 4, 5])
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: avg with floats
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      avg([1.5, 2.5, 3.0])
      """
    When I run the application with this content
    Then the output should contain "2.33"

  Scenario: avg single element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      avg([42])
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: avg empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      avg([])
      """
    When I run the application and it fails
    Then the output should contain "empty"

  Scenario: max basic usage via maxBy
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": 3}, {"val": 1}, {"val": 9}, {"val": 2}] maxBy (x) -> x.val
      """
    When I run the application with this content
    Then the output should contain "\"val\":9"

  Scenario: maxBy with various numbers  
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": 5}, {"val": 10}, {"val": 3}] maxBy (x) -> x.val
      """
    When I run the application with this content
    Then the output should contain "\"val\":10"

  Scenario: min basic usage via minBy
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": 3}, {"val": 1}, {"val": 9}, {"val": 2}] minBy (x) -> x.val
      """
    When I run the application with this content
    Then the output should contain "\"val\":1"

  Scenario: minBy with various numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": 5}, {"val": 1}, {"val": 10}] minBy (x) -> x.val
      """
    When I run the application with this content
    Then the output should contain "\"val\":1"

  Scenario: maxBy basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}, {"name": "Charlie", "age": 35}] maxBy (person) -> person.age
      """
    When I run the application with this content
    Then the output should contain "Charlie"
    And the output should contain "35"

  Scenario: maxBy returns full object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"id": 1, "score": 50}, {"id": 2, "score": 100}, {"id": 3, "score": 75}] maxBy (item) -> item.score
      """
    When I run the application with this content
    Then the output should be valid JSON with 2 keys
    And the output should contain "\"id\":2"

  Scenario: maxBy with numeric values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"score": 45}, {"score": 89}, {"score": 67}] maxBy (person) -> person.score
      """
    When I run the application with this content
    Then the output should contain "89"

  Scenario: maxBy single element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"value": 42}] maxBy (item) -> item.value
      """
    When I run the application with this content
    Then the output should be:
      """
      {"value":42}
      """

  Scenario: maxBy empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] maxBy (x) -> x
      """
    When I run the application and it fails
    Then the output should contain "empty"

  Scenario: minBy basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}, {"name": "Charlie", "age": 35}] minBy (person) -> person.age
      """
    When I run the application with this content
    Then the output should contain "Bob"
    And the output should contain "25"

  Scenario: minBy returns full object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"id": 1, "score": 50}, {"id": 2, "score": 100}, {"id": 3, "score": 75}] minBy (item) -> item.score
      """
    When I run the application with this content
    Then the output should be valid JSON with 2 keys
    And the output should contain "\"id\":1"

  Scenario: minBy with numbers for string objects
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"priority": 3}, {"priority": 1}, {"priority": 2}] minBy (person) -> person.priority
      """
    When I run the application with this content
    Then the output should contain "\"priority\":1"

  Scenario: minBy single element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"value": 42}] minBy (item) -> item.value
      """
    When I run the application with this content
    Then the output should be:
      """
      {"value":42}
      """

  Scenario: minBy empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] minBy (x) -> x
      """
    When I run the application and it fails
    Then the output should contain "empty"

  Scenario: orderBy basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"order": 3}, {"order": 1}, {"order": 2}] orderBy (person) -> person.order
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"order":1},{"order":2},{"order":3}]
      """

  Scenario: orderBy numeric
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"id": 3}, {"id": 1}, {"id": 2}] orderBy (item) -> item.id
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"id":1},{"id":2},{"id":3}]
      """

  Scenario: orderBy reverse order
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"age": 25}, {"age": 35}, {"age": 30}] orderBy (person) -> person.age
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"age":25},{"age":30},{"age":35}]
      """

  Scenario: orderBy single element
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"value": 42}] orderBy (item) -> item.value
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"value":42}]
      """

  Scenario: orderBy empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] orderBy (x) -> x
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: orderBy with numeric priority
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"priority": 5}, {"priority": 2}, {"priority": 8}] orderBy (item) -> item.priority
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"priority":2},{"priority":5},{"priority":8}]
      """

  Scenario: sort distinguishes adjacent integers above the exact float range
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sort([9007199254740993, 9007199254740992])
      """
    When I run the application with this content
    Then the output should be:
      """
      [9007199254740992,9007199254740993]
      """

  Scenario: orderBy distinguishes adjacent integers above the exact float range
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [9007199254740993, 9007199254740992] orderBy $
      """
    When I run the application with this content
    Then the output should be:
      """
      [9007199254740992,9007199254740993]
      """

  Scenario: minBy and maxBy distinguish adjacent integers above the exact float range
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[9007199254740993, 9007199254740992] minBy $, [9007199254740992, 9007199254740993] maxBy $]
      """
    When I run the application with this content
    Then the output should be:
      """
      [9007199254740992,9007199254740993]
      """

  Scenario: orderBy reports incomparable key positions
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, "2", 0] orderBy $
      """
    When I run the application and it fails
    Then the error should contain "orderBy: cannot compare keys at indexes"
    And the error should contain "cannot compare values"
    And the error should contain "4:1:"

  Scenario: sum with strings converted to numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      sum([1, 2, 3, 4])
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  Scenario: maxBy with index parameter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": 10}, {"val": 50}, {"val": 30}] maxBy (item, index) -> item.val
      """
    When I run the application with this content
    Then the output should contain "50"

  Scenario: minBy with index parameter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": 10}, {"val": 50}, {"val": 30}] minBy (item, index) -> item.val
      """
    When I run the application with this content
    Then the output should contain "10"

  Scenario: orderBy with index parameter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"val": 30}, {"val": 10}, {"val": 20}] orderBy (item, index) -> item.val
      """
     When I run the application with this content
     Then the output should be:
       """
       [{"val":10},{"val":20},{"val":30}]
       """

   Scenario: max with array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       max([5, 2, 8, 1, 9])
       """
     When I run the application with this content
     Then the output should be:
       """
       9
       """

   Scenario: max with multiple arguments
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       max(5, 2, 8, 1, 9)
       """
     When I run the application with this content
     Then the output should be:
       """
       9
       """

   Scenario: max with single number
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       max(42)
       """
     When I run the application with this content
     Then the output should be:
       """
       42
       """

   Scenario: max with empty array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       max([])
       """
     When I run the application and it fails
     Then the output should contain "max function requires non-empty array"

   Scenario: min with array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       min([5, 2, 8, 1, 9])
       """
     When I run the application with this content
     Then the output should be:
       """
       1
       """

   Scenario: min with multiple arguments
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       min(5, 2, 8, 1, 9)
       """
     When I run the application with this content
     Then the output should be:
       """
       1
       """

   Scenario: min with single number
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       min(42)
       """
     When I run the application with this content
     Then the output should be:
       """
       42
       """

   Scenario: min with empty array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       min([])
       """
     When I run the application and it fails
     Then the output should contain "min function requires non-empty array"

  Scenario: filter chained with maxBy
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 5, 2, 8, 3] filter $ > 2 maxBy $
      """
    When I run the application with this content
    Then the output should be:
      """
      8
      """

  Scenario: filter chained with minBy
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 5, 2, 8, 3] filter $ > 2 minBy $
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: map chained with maxBy
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"n": 1}, {"n": 5}, {"n": 3}] map $.n maxBy $
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: filter chained with distinctBy
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"a": 1}, {"a": 2}, {"a": 1}, {"a": 3}] filter $.a > 0 distinctBy $.a
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"a":1},{"a":2},{"a":3}]
      """
