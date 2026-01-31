Feature: Array Functions
  In order to manipulate and transform arrays
  As a developer
  I want to use the array functions: groupBy, pluck, and flatMap

  Scenario: groupBy basic usage with simple property
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"type": "A", "value": 1}, {"type": "B", "value": 2}, {"type": "A", "value": 3}] groupBy (x) -> x.type
      """
    When I run the application with this content
    Then the output should contain "\"A\":"
    And the output should contain "\"B\":"

  Scenario: groupBy with numeric grouping
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"id": 1, "category": 10}, {"id": 2, "category": 20}, {"id": 3, "category": 10}] groupBy (x) -> x.category
      """
    When I run the application with this content
    Then the output should be valid JSON with 2 keys

  Scenario: groupBy with index parameter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice"}, {"name": "Bob"}, {"name": "Charlie"}] groupBy (item, index) -> index
      """
    When I run the application with this content
    Then the output should be valid JSON with 3 keys

  Scenario: groupBy with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] groupBy (x) -> x
      """
    When I run the application with this content
    Then the output should be:
      """
      {}
      """

  Scenario: prepend adds item to array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      prepend([2, 3], 1)
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: append adds item to array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      append([1, 2], 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: groupBy requires lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] groupBy "not a lambda"
      """
    When I run the application and it fails
    Then the output should contain "groupBy expects a lambda function"

  Scenario: groupBy on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "not an array" groupBy (x) -> x
      """
    When I run the application and it fails
    Then the output should contain "groupBy expects an array"

  Scenario: pluck basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice", "age": 30}, {"name": "Bob", "age": 25}, {"name": "Charlie", "age": 35}] pluck "name"
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice","Bob","Charlie"]
      """

  Scenario: pluck with numeric property
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"id": 1, "value": 100}, {"id": 2, "value": 200}] pluck "value"
      """
    When I run the application with this content
    Then the output should be:
      """
      [100,200]
      """

  Scenario: pluck nested property
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"user": {"name": "Alice"}}, {"user": {"name": "Bob"}}] pluck "user.name"
      """
    When I run the application with this content
    Then the output should be:
      """
      ["Alice","Bob"]
      """

  Scenario: pluck with missing properties returns nulls
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice"}, {"name": "Bob"}, {}] pluck "name"
      """
    When I run the application with this content
    Then the output should contain "Alice"
    And the output should contain "Bob"
    And the output should contain "null"

  Scenario: pluck on empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] pluck "name"
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: pluck on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "not an array" pluck "name"
      """
    When I run the application and it fails
    Then the output should contain "pluck expects an array"

  Scenario: pluck requires string key
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice"}] pluck 123
      """
    When I run the application and it fails
    Then the output should contain "pluck on array expects a string or lambda function"

  Scenario: flatMap basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [[1, 2], [3, 4], [5]] flatMap (x) -> x
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3,4,5]
      """

  Scenario: flatMap with transformation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] flatMap (x) -> [x, x * 2]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,2,4,3,6]
      """

  Scenario: flatMap with conditional logic
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4] flatMap (x) -> if (x > 2) [x, x + 10] else [x]
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2,3,13,4,14]
      """

  Scenario: flatMap with index parameter
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["a", "b", "c"] flatMap (item, index) -> [index, item]
      """
    When I run the application with this content
    Then the output should be:
      """
      [0,"a",1,"b",2,"c"]
      """

  Scenario: flatMap flattens only one level
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2] flatMap (x) -> [[x, x]]
      """
    When I run the application with this content
    Then the output should be:
      """
      [[1,1],[2,2]]
      """

  Scenario: flatMap with non-array results
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] flatMap (x) -> x * 2
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,4,6]
      """

  Scenario: flatMap on empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] flatMap (x) -> [x, x]
      """
    When I run the application with this content
    Then the output should be:
      """
      []
      """

  Scenario: flatMap on non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "not an array" flatMap (x) -> [x]
      """
    When I run the application and it fails
    Then the output should contain "flatMap expects an array"

  Scenario: flatMap requires lambda
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3] flatMap "not a lambda"
      """
    When I run the application and it fails
    Then the output should contain "flatMap expects a lambda function"

  Scenario: pluck with deeply nested property
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"a": {"b": {"c": 1}}}, {"a": {"b": {"c": 2}}}] pluck "a.b.c"
      """
    When I run the application with this content
    Then the output should be:
      """
      [1,2]
      """

  Scenario: groupBy with string values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [{"name": "Alice", "status": "active"}, {"name": "Bob", "status": "inactive"}, {"name": "Charlie", "status": "active"}] groupBy (x) -> x.status
      """
    When I run the application with this content
    Then the output should contain "\"active\":"
    And the output should contain "\"inactive\":"
    And the output should be valid JSON with 2 keys

   Scenario: flatMap combined with filter
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       [[1, 2, 3], [4, 5]] flatMap (x) -> x
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2,3,4,5]
       """

   Scenario: first basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       first([1, 2, 3])
       """
     When I run the application with this content
     Then the output should be:
       """
       1
       """

   Scenario: first on empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       first([])
       """
     When I run the application with this content
     Then the output should be:
       """
       null
       """

   Scenario: first on non-array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       first("not an array")
       """
     When I run the application and it fails
     Then the output should contain "first: argument 1 expected array, got string"

   Scenario: last basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       last([1, 2, 3])
       """
     When I run the application with this content
     Then the output should be:
       """
       3
       """

   Scenario: last on empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       last([])
       """
     When I run the application with this content
     Then the output should be:
       """
       null
       """

   Scenario: last on non-array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       last("not an array")
       """
     When I run the application and it fails
     Then the output should contain "last: argument 1 expected array, got string"

   Scenario: take basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       take([1, 2, 3, 4, 5], 3)
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2,3]
       """

   Scenario: take more than array length
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       take([1, 2, 3], 10)
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2,3]
       """

   Scenario: take zero elements
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       take([1, 2, 3], 0)
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: take from empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       take([], 3)
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: take on non-array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       take("not an array", 2)
       """
     When I run the application and it fails
     Then the output should contain "take: argument 1 expected array, got string"

   Scenario: drop basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       drop([1, 2, 3, 4, 5], 2)
       """
     When I run the application with this content
     Then the output should be:
       """
       [3,4,5]
       """

   Scenario: drop more than array length
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       drop([1, 2, 3], 10)
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: drop zero elements
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       drop([1, 2, 3], 0)
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2,3]
       """

   Scenario: drop from empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       drop([], 3)
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: drop on non-array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       drop("not an array", 2)
       """
     When I run the application and it fails
     Then the output should contain "drop: argument 1 expected array, got string"

   Scenario: takeWhile basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       takeWhile([1, 2, 3, 4, 5], (x) -> x < 4)
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2,3]
       """

   Scenario: takeWhile with all matching
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       takeWhile([1, 2, 3], (x) -> x < 10)
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2,3]
       """

   Scenario: takeWhile with none matching
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       takeWhile([1, 2, 3], (x) -> x > 10)
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: takeWhile on empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       takeWhile([], (x) -> x < 4)
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: takeWhile requires lambda
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       takeWhile([1, 2, 3], "not a lambda")
       """
     When I run the application and it fails
     Then the output should contain "takeWhile expects a lambda function"

   Scenario: takeWhile on non-array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       takeWhile("not an array", (x) -> x < 4)
       """
     When I run the application and it fails
     Then the output should contain "takeWhile expects an array"

   Scenario: dropWhile basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dropWhile([1, 2, 3, 4, 5], (x) -> x < 4)
       """
     When I run the application with this content
     Then the output should be:
       """
       [4,5]
       """

   Scenario: dropWhile with all matching
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dropWhile([1, 2, 3], (x) -> x < 10)
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: dropWhile with none matching
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dropWhile([1, 2, 3], (x) -> x > 10)
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2,3]
       """

   Scenario: dropWhile on empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dropWhile([], (x) -> x < 4)
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: dropWhile requires lambda
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dropWhile([1, 2, 3], "not a lambda")
       """
     When I run the application and it fails
     Then the output should contain "dropWhile expects a lambda function"

   Scenario: dropWhile on non-array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dropWhile("not an array", (x) -> x < 4)
       """
      When I run the application and it fails
      Then the output should contain "dropWhile expects an array"

   Scenario: distinct basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       distinct([1, 2, 2, 3, 3, 3])
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,2,3]
       """

   Scenario: distinct with strings
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       distinct(["a", "b", "a", "c", "b"])
       """
     When I run the application with this content
     Then the output should be:
       """
       ["a","b","c"]
       """

   Scenario: distinct with mixed types
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       distinct([1, "1", 1, 2])
       """
     When I run the application with this content
     Then the output should be:
       """
       [1,"1",2]
       """

   Scenario: distinct with objects
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       distinct([{"x": 1}, {"x": 2}, {"x": 1}])
       """
     When I run the application with this content
     Then the output should be:
       """
       [{"x":1},{"x":2}]
       """

   Scenario: distinct with empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       distinct([])
       """
     When I run the application with this content
     Then the output should be:
       """
       []
       """

   Scenario: distinct preserves order
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       distinct([3, 1, 2, 1, 3, 2])
       """
     When I run the application with this content
     Then the output should be:
       """
       [3,1,2]
       """

   Scenario: distinct on non-array fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       distinct("not an array")
       """
     When I run the application and it fails
     Then the output should contain "distinct: argument 1 expected array, got string"

   Scenario: some basic usage true
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       some([1, 2, 3], (x) -> x > 2)
       """
     When I run the application with this content
     Then the output should be true

   Scenario: some basic usage false
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       some([1, 2, 3], (x) -> x > 10)
       """
     When I run the application with this content
     Then the output should be false

   Scenario: some with empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       some([], (x) -> x > 0)
       """
     When I run the application with this content
     Then the output should be false

   Scenario: some all match
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       some([2, 4, 6], (x) -> x mod 2 == 0)
       """
     When I run the application with this content
     Then the output should be true

   Scenario: every basic usage true
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       every([1, 2, 3], (x) -> x > 0)
       """
     When I run the application with this content
     Then the output should be true

   Scenario: every basic usage false
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       every([1, 2, 3], (x) -> x > 1)
       """
     When I run the application with this content
     Then the output should be false

   Scenario: every with empty array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       every([], (x) -> x > 0)
       """
     When I run the application with this content
     Then the output should be true

   Scenario: every all match
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       every([2, 4, 6], (x) -> x mod 2 == 0)
       """
     When I run the application with this content
     Then the output should be true

   Scenario: some requires lambda
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       some([1, 2, 3], "not a lambda")
       """
     When I run the application and it fails
     Then the output should contain "some expects a lambda function"

   Scenario: every requires lambda
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       every([1, 2, 3], "not a lambda")
       """
     When I run the application and it fails
     Then the output should contain "every expects a lambda function"

   Scenario: some requires array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       some("not an array", (x) -> x > 0)
       """
     When I run the application and it fails
     Then the output should contain "some expects an array"

   Scenario: every requires array
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       every("not an array", (x) -> x > 0)
       """
     When I run the application and it fails
     Then the output should contain "every expects an array"

  Scenario: slice basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      slice([1, 2, 3, 4], 1, 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,3]
      """
