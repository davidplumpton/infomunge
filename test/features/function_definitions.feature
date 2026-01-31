Feature: Function Definitions
  In order to define reusable logic
  As a developer
  I want to declare custom functions in the header and call them in the body

  Scenario: Simple function returning a field
    Given the following input content:
      """
      %im 0.1
      fun toUser(user) = {firstName: user.name}
      output application/json
      ---
      toUser({name: "Alice"})
      """
    When I run the application with this content
    Then the output should be:
      """
      {"firstName":"Alice"}
      """

  Scenario: Function with multiple parameters
    Given the following input content:
      """
      %im 0.1
      fun add(a, b) = a + b
      output application/json
      ---
      add(2, 3)
      """
    When I run the application with this content
    Then the output should be:
      """
      5
      """

  Scenario: Function with inline comment after body
    Given the following input content:
      """
      %im 0.1
      fun wrap(value) = {value: value} // ) ] } in comment
      output application/json
      ---
      wrap(1)
      """
    When I run the application with this content
    Then the output should be:
      """
      {"value":1}
      """

  Scenario: Function with arithmetic expression
    Given the following input content:
      """
      %im 0.1
      fun double(x) = x * 2
      output application/json
      ---
      double(21)
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Function returning an object literal
    Given the following input content:
      """
      %im 0.1
      fun makePerson(name, age) = {name: name, age: age}
      output application/json
      ---
      makePerson("Bob", 30)
      """
    When I run the application with this content
    Then the output should be:
      """
      {"age":30,"name":"Bob"}
      """

  Scenario: Function called multiple times
    Given the following input content:
      """
      %im 0.1
      fun square(n) = n * n
      output application/json
      ---
      [square(2), square(3), square(4)]
      """
    When I run the application with this content
    Then the output should be:
      """
      [4,9,16]
      """

  Scenario: Function using dot notation in body
    Given the following input content:
      """
      %im 0.1
      fun getName(obj) = obj.name
      output application/json
      ---
      getName({name: "Charlie", age: 25})
      """
    When I run the application with this content
    Then the output should be:
      """
      "Charlie"
      """

  Scenario: Function using string concatenation
    Given the following input content:
      """
      %im 0.1
      fun greet(name) = "Hello, " + name + "!"
      output application/json
      ---
      greet("World")
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello, World!"
      """

  Scenario: Multiple function definitions
    Given the following input content:
      """
      %im 0.1
      fun double(x) = x * 2
      fun triple(x) = x * 3
      output application/json
      ---
      {doubled: double(5), tripled: triple(5)}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"doubled":10,"tripled":15}
      """

  Scenario: Function using header variable
    Given the following input content:
      """
      %im 0.1
      var multiplier = 10
      fun scale(x) = x * multiplier
      output application/json
      ---
      scale(5)
      """
    When I run the application with this content
    Then the output should be:
      """
      50
      """

  Scenario: Function with no parameters
    Given the following input content:
      """
      %im 0.1
      fun getAnswer() = 42
      output application/json
      ---
      getAnswer()
      """
    When I run the application with this content
    Then the output should be:
      """
      42
      """

  Scenario: Function with array access
    Given the following input content:
      """
      %im 0.1
      fun first(arr) = arr[0]
      output application/json
      ---
      first([10, 20, 30])
      """
    When I run the application with this content
    Then the output should be:
      """
      10
      """

  Scenario: Function used in map
    Given the following input content:
      """
      %im 0.1
      fun double(x) = x * 2
      var nums = [1, 2, 3]
      output application/json
      ---
      nums map (n) -> double(n)
      """
    When I run the application with this content
    Then the output should be:
      """
      [2,4,6]
      """

  Scenario: Function with do block separator in header
    Given the following input content:
      """
      %im 0.1
      fun addOne(x) = do {
        var y = x + 1
        ---
        y
      }
      output application/json
      ---
      addOne(2)
      """
    When I run the application with this content
    Then the output should be:
      """
      3
      """

  Scenario: Multiline function with return type annotation
    Given the following input content:
      """
      %dw 2.0
      output application/json
      var myData = [{name:1},{name:2},{name:3}]
      fun myExternalFunction(data): Array =
          if(data.name == 1)
              []
          else if(data.name == 2)
              [{name: 3}, {name:5}]
          else
              [data]
      ---
      myData flatMap ((item, index) -> myExternalFunction(item))
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"name":3},{"name":5},{"name":3}]
      """

  Scenario: Multiline function with if/else body
    Given the following input content:
      """
      %im 0.1
      fun classify(n) =
          if (n > 10)
              "high"
          else if (n > 0)
              "low"
          else
              "zero"
      output application/json
      ---
      [classify(20), classify(5), classify(0)]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["high","low","zero"]
      """

  Scenario: Multiline function body starting on the equals line with match expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      fun deepTransform(data) = data match {
          case is Array  -> data map (item) -> deepTransform(item)
          case is Object -> data mapObject (v, k) -> {
              (k): deepTransform(v)
          }
          case n is Number -> round(n * 100) / 100
          else -> data
      }
      ---
      deepTransform([{price: 45.556}, [1, {nested: 2.345}]])
      """
    When I run the application with this content
    Then the output should be:
      """
      [{"price":45.56},[1,{"nested":2.35}]]
      """

  Scenario: Comments between function declarations in header
    Given the following input content:
      """
      %im 0.1
      output application/json
      // First function
      fun double(x) =
          x * 2
      // Second function
      fun triple(x) =
          x * 3
      ---
      [double(5), triple(5)]
      """
    When I run the application with this content
    Then the output should be:
      """
      [10,15]
      """

  Scenario: Reduce with object literal as accumulator default value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3, 4] reduce ((item, acc = {sum: 0, count: 0}) -> {sum: acc.sum + item, count: acc.count + 1})
      """
    When I run the application with this content
    Then the output should be:
      """
      {"count":4,"sum":10}
      """

  Scenario: Comments in body expression are stripped
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
          // Comment in body
          name: "test",
          // Another comment
          value: 42
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"name":"test","value":42}
      """

  Scenario: Multi-line if/else in body expression
    Given the following input content:
      """
      %im 0.1
      output application/json
      var x = 5
      ---
      {
          result: if (x > 10)
                      "high"
                  else
                      "low"
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"result":"low"}
      """
