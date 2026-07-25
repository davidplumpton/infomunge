Feature: Lazy Evaluation
  As a user of InfoMunge
  I want lazy evaluation capabilities
  So that I can process large datasets efficiently

  Scenario: Create lazy value
    Given the script:
      """
      lazy_eval(42)
      """
    When I execute the script
    Then the result should be 42

  Scenario: Force evaluation of lazy value
    Given the script:
      """
      force_eval(lazy_eval(42))
      """
    When I execute the script
    Then the result should be 42

  Scenario: Lazy value caches evaluation
    Given the script:
      """
      %im 0.1
      output application/json
      var value = lazy_eval(uuid())
      ---
      force_eval(value) == force_eval(value)
      """
    When I execute the script
    Then the output should be true

  Scenario: Lazy map on stream
    Given the script:
      """
      __lazyMap(__toStream([1,2,3]), (x) -> x)
      """
    When I execute the script
    Then the output should be:
      """
      [1,2,3]
      """

  Scenario: Lazy reduce rejects invalid lambda arity
    Given the script:
      """
      __lazyReduce(__toStream([1,2,3]), (x) -> x, 0)
      """
    Then running the script should fail with error containing "lazyReduce lambda must have 2 or 3 parameters"

  Scenario: Lazy map with transformation
    Given the script:
      """
      __lazyMap(__toStream([1,2,3]), (x) -> x * 2)
      """
    When I execute the script
    Then the output should be:
      """
      [2,4,6]
      """

  Scenario: Lazy filter on stream
    Given the script:
      """
      __lazyFilter(__toStream([1,2,3,4,5]), (x) -> x > 2)
      """
    When I execute the script
    Then the output should be:
      """
      [3,4,5]
      """

  Scenario: Lazy reduce with sum
    Given the script:
      """
      __lazyReduce(__toStream([1,2,3,4]), (acc, x) -> acc + x, 0)
      """
    When I execute the script
    Then the result should be 10

  Scenario: Lazy reduce with product
    Given the script:
      """
      __lazyReduce(__toStream([1,2,3,4]), (acc, x) -> acc * x, 1)
      """
    When I execute the script
    Then the result should be 24

  Scenario: Chained lazy map operations
    Given the script:
      """
      __lazyMap(__lazyMap(__toStream([1,2,3]), (x) -> x + 1), (x) -> x * 2)
      """
    When I execute the script
    Then the output should be:
      """
      [4,6,8]
      """

  Scenario: Chained lazy filter and map
    Given the script:
      """
      __lazyMap(__lazyFilter(__toStream([1,2,3,4,5]), (x) -> x > 2), (x) -> x * 10)
      """
    When I execute the script
    Then the output should be:
      """
      [30,40,50]
      """

  Scenario: Documented lazy pipeline example is an executable standalone script
    Given the lazy evaluation documentation example
    When I execute the script
    Then the output should be:
      """
      [16,25]
      """

  Scenario: Lazy filter with empty result
    Given the script:
      """
      __lazyFilter(__toStream([1,2,3]), (x) -> x > 10)
      """
    When I execute the script
    Then the output should be:
      """
      null
      """

  Scenario: Lazy operations on empty stream
    Given the script:
      """
      __lazyMap(__toStream([]), (x) -> x * 2)
      """
    When I execute the script
    Then the output should be:
      """
      null
      """

  Scenario: Lazy reduce on empty stream
    Given the script:
      """
      __lazyReduce(__toStream([]), (acc, x) -> acc + x, 0)
      """
    When I execute the script
    Then the result should be 0

  Scenario: Lazy map requires lazy value
    Given the script:
      """
      __lazyMap([1,2,3], (x) -> x)
      """
    Then running the script should fail with error containing "__lazyMap first argument must be a lazy value"

  Scenario: Lazy filter requires lazy value
    Given the script:
      """
      __lazyFilter([1,2,3], (x) -> x > 1)
      """
    Then running the script should fail with error containing "__lazyFilter first argument must be a lazy value"

  Scenario: Lazy reduce requires lazy value
    Given the script:
      """
      __lazyReduce([1,2,3], (acc, x) -> acc + x, 0)
      """
    Then running the script should fail with error containing "__lazyReduce first argument must be a lazy value"

  Scenario: toStream requires array
    Given the script:
      """
      __toStream("not an array")
      """
    Then running the script should fail with error containing "__toStream expects an array"

  Scenario: Lazy map requires lambda
    Given the script:
      """
      __lazyMap(__toStream([1,2,3]), "not a lambda")
      """
    Then running the script should fail with error containing "__lazyMap second argument must be a lambda"

  Scenario: Lazy filter requires lambda
    Given the script:
      """
      __lazyFilter(__toStream([1,2,3]), "not a lambda")
      """
    Then running the script should fail with error containing "__lazyFilter second argument must be a lambda"

  Scenario: Lazy filter rejects non-boolean predicate result
    Given the script:
      """
      __lazyFilter(__toStream([1,2,3]), (x) -> x)
      """
    Then running the script should fail with error containing "filter lambda must return a boolean"

  Scenario: Lazy reduce requires lambda
    Given the script:
      """
      __lazyReduce(__toStream([1,2,3]), "not a lambda", 0)
      """
    Then running the script should fail with error containing "__lazyReduce second argument must be a lambda"

  Scenario: Lazy reduce with three parameter lambda
    Given the script:
      """
      __lazyReduce(__toStream([10,20,30]), (acc, x, idx) -> acc + x + idx, 0)
      """
    When I execute the script
    Then the result should be 63

  Scenario: Lazy evaluation with string transformation
    Given the script:
      """
      __lazyMap(__toStream(["a", "b", "c"]), (x) -> x ++ "!")
      """
    When I execute the script
    Then the output should be:
      """
      ["a!","b!","c!"]
      """

  Scenario: Lazy mode flag is rejected
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      1 + 1
      """
    When I run the application with this content using --lazy and it fails
    Then the error should contain "--lazy is currently unsupported"
