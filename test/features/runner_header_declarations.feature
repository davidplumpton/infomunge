Feature: Runner header declaration paths (in-process coverage)
  In order to verify runner header parsing and declaration handling logic
  As a developer
  I want to exercise parseHeader, directive dispatch, variable declarations, function declarations,
  type declarations, namespace declarations, imports, and shared declaration IR application in-process

  # --- Variable declarations ---

  Scenario: Simple string variable via in-process runner
    Given the following script:
      """
      %im 0.1
      var name = "InfoMunge"
      output application/json
      ---
      { tool: name }
      """
    When I run the script
    Then the output should be:
      """
      {"tool":"InfoMunge"}
      """

  Scenario: Number variable in expression via in-process runner
    Given the following script:
      """
      %im 0.1
      var base = 10
      output application/json
      ---
      { result: base * 2 }
      """
    When I run the script
    Then the output should be:
      """
      {"result":20}
      """

  Scenario: Multiple variables via in-process runner
    Given the following script:
      """
      %im 0.1
      var firstName = "John"
      var lastName = "Doe"
      output application/json
      ---
      { fullName: firstName ++ " " ++ lastName }
      """
    When I run the script
    Then the output should be:
      """
      {"fullName":"John Doe"}
      """

  Scenario: Multiline object variable via in-process runner
    Given the following script:
      """
      %im 0.1
      var config =
        {
          name: "InfoMunge",
          version: 1
        }
      output application/json
      ---
      { tool: config.name, version: config.version }
      """
    When I run the script
    Then the output should be:
      """
      {"tool":"InfoMunge","version":1}
      """

  Scenario: Variable referencing another variable
    Given the following script:
      """
      %im 0.1
      var x = 10
      var y = x + 5
      output application/json
      ---
      y
      """
    When I run the script
    Then the output should be:
      """
      15
      """

  Scenario: Variable with array value
    Given the following script:
      """
      %im 0.1
      var items = [1, 2, 3]
      output application/json
      ---
      sizeOf(items)
      """
    When I run the script
    Then the output should be:
      """
      3
      """

  Scenario: Reserved runtime metadata names cannot be declared as variables
    Given the following script:
      """
      %im 0.1
      var __output_metadata__ = "shadow"
      output application/json
      ---
      __output_metadata__
      """
    Then running the script should fail with error containing "\"__output_metadata__\" is reserved for runtime metadata"

  # --- Function declarations ---

  Scenario: Simple function via in-process runner
    Given the following script:
      """
      %im 0.1
      fun toUser(user) = {firstName: user.name}
      output application/json
      ---
      toUser({name: "Alice"})
      """
    When I run the script
    Then the output should be:
      """
      {"firstName":"Alice"}
      """

  Scenario: Function with multiple parameters via in-process runner
    Given the following script:
      """
      %im 0.1
      fun add(a, b) = a + b
      output application/json
      ---
      add(2, 3)
      """
    When I run the script
    Then the output should be:
      """
      5
      """

  Scenario: Function with no parameters via in-process runner
    Given the following script:
      """
      %im 0.1
      fun getAnswer() = 42
      output application/json
      ---
      getAnswer()
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: Function with inline comment via in-process runner
    Given the following script:
      """
      %im 0.1
      fun wrap(value) = {value: value} // ) ] } in comment
      output application/json
      ---
      wrap(1)
      """
    When I run the script
    Then the output should be:
      """
      {"value":1}
      """

  Scenario: Multiple function definitions via in-process runner
    Given the following script:
      """
      %im 0.1
      fun double(x) = x * 2
      fun triple(x) = x * 3
      output application/json
      ---
      {doubled: double(5), tripled: triple(5)}
      """
    When I run the script
    Then the output should be:
      """
      {"doubled":10,"tripled":15}
      """

  Scenario: Function using header variable via in-process runner
    Given the following script:
      """
      %im 0.1
      var multiplier = 10
      fun scale(x) = x * multiplier
      output application/json
      ---
      scale(5)
      """
    When I run the script
    Then the output should be:
      """
      50
      """

  Scenario: Multiline function with do block via in-process runner
    Given the following script:
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
    When I run the script
    Then the output should be:
      """
      3
      """

  Scenario: Multiline function with if/else body via in-process runner
    Given the following script:
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
    When I run the script
    Then the output should be:
      """
      ["high","low","zero"]
      """

  Scenario: Comments between function declarations via in-process runner
    Given the following script:
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
    When I run the script
    Then the output should be:
      """
      [10,15]
      """

  # --- Type declarations ---

  Scenario: Simple type declaration via in-process runner
    Given the following script:
      """
      %im 0.1
      type Currency = String
      output application/json
      ---
      "hello"
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  Scenario: Type declaration with format property via in-process runner
    Given the following script:
      """
      %im 0.1
      type Currency = String { format: "##.00" }
      output application/json
      ---
      42.5 as Currency
      """
    When I run the script
    Then the output should be:
      """
      "42.50"
      """

  Scenario: Type with multiple properties via in-process runner
    Given the following script:
      """
      %im 0.1
      type Price = String { format: "##.00", prefix: "$" }
      output application/json
      ---
      99.9 as Price
      """
    When I run the script
    Then the output should be:
      """
      "99.90"
      """

  Scenario: Type with empty properties block via in-process runner
    Given the following script:
      """
      %im 0.1
      type Plain = String { }
      output application/json
      ---
      "hello" as Plain
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  Scenario: Multiple type declarations via in-process runner
    Given the following script:
      """
      %im 0.1
      type Name = String
      type Age = Number
      output application/json
      ---
      {name: "Alice", age: 30}
      """
    When I run the script
    Then the output should be:
      """
      {"age":30,"name":"Alice"}
      """

  # --- Namespace declarations ---

  Scenario: Prefixed namespace declaration via in-process runner
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces=All
      ns a http://example.com/a
      ---
      root: { child: "value" }
      """
    When I run the script through the runner output path
    Then the output should contain "xmlns:a"

  Scenario: Default namespace declaration via in-process runner
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces=All
      ns http://example.com/default
      ---
      root: "value"
      """
    When I run the script through the runner output path
    Then the output should contain "xmlns="

  Scenario: Multiple namespace declarations via in-process runner
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaredNamespaces=All
      ns a http://example.com/a
      ns b http://example.com/b
      ---
      root: { child: "value" }
      """
    When I run the script through the runner output path
    Then the output should contain "xmlns:a"
    And the output should contain "xmlns:b"

  # --- Import directive ---

  Scenario: Import module and call function via in-process runner
    Given the following script:
      """
      %im 0.1
      import modules::MathUtils
      output application/json
      ---
      MathUtils::double(21)
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: Import all from module via in-process runner
    Given the following script:
      """
      %im 0.1
      import * from modules::MathUtils
      output application/json
      ---
      double(21)
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: Import specific functions via in-process runner
    Given the following script:
      """
      %im 0.1
      import double, triple from modules::MathUtils
      output application/json
      ---
      {doubled: double(7), tripled: triple(7)}
      """
    When I run the script
    Then the output should be:
      """
      {"doubled":14,"tripled":21}
      """

  Scenario: Import with module variable via in-process runner
    Given the following script:
      """
      %im 0.1
      import modules::MathUtils
      output application/json
      ---
      MathUtils::addOffset(10)
      """
    When I run the script
    Then the output should be:
      """
      15
      """

  Scenario: Mix imported and local functions via in-process runner
    Given the following script:
      """
      %im 0.1
      import double from modules::MathUtils
      fun quadruple(x) = double(double(x))
      output application/json
      ---
      quadruple(5)
      """
    When I run the script
    Then the output should be:
      """
      20
      """

  # --- Output option parsing ---

  Scenario: Output options with brace syntax via in-process runner
    Given the following script:
      """
      %im 0.1
      output application/xml {writeDeclaration=false}
      ---
      item: "test"
      """
    When I run the script through the runner output path
    Then the output should not contain "<?xml"
    And the output should contain "<item>test</item>"

  Scenario: Output options with comma-separated values via in-process runner
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaration=false,writeNilOnNull=true
      ---
      root: { item: null }
      """
    When I run the script through the runner output path
    Then the output should not contain "<?xml"
    And the output should contain "xsi:nil"

  Scenario: Output options with quoted value via in-process runner
    Given the following script:
      """
      %im 0.1
      output application/xml skipNullOn="elements"
      ---
      root: { item: null, keep: "yes" }
      """
    When I run the script through the runner output path
    Then the output should not contain "item"
    And the output should contain "<keep>yes</keep>"

  Scenario: Output options with colon separator via in-process runner
    Given the following script:
      """
      %im 0.1
      output application/xml writeDeclaration:false
      ---
      item: "test"
      """
    When I run the script through the runner output path
    Then the output should not contain "<?xml"
    And the output should contain "<item>test</item>"

  # --- Input directive ---

  Scenario: Input directive sets mime type via in-process runner
    Given the following script:
      """
      %im 0.1
      input application/json
      output application/json
      ---
      "ok"
      """
    When I run the script
    Then the output should be:
      """
      "ok"
      """

  # --- Single-line header normalization ---

  Scenario: Single-line header with vars and output via in-process runner
    Given the following script:
      """
      %im 0.1 var message = "output application/xml is just text" var amount = 2 + 3 output application/json --- {message: message, amount: amount}
      """
    When I run the script
    Then the output should be:
      """
      {"amount":5,"message":"output application/xml is just text"}
      """

  Scenario: Single-line header with function definition via in-process runner
    Given the following script:
      """
      %im 0.1 fun double(x) = x * 2 output application/json --- double(21)
      """
    When I run the script
    Then the output should be:
      """
      42
      """

  Scenario: Single-line header with type definition via in-process runner
    Given the following script:
      """
      %im 0.1 type Name = String output application/json --- "hello"
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  Scenario: Single-line header with namespace via in-process runner
    Given the following script:
      """
      %im 0.1 ns a urn:example:a output application/xml writeDeclaredNamespaces=All --- root: "value"
      """
    When I run the script through the runner output path
    Then the output should contain "xmlns:a"

  # --- Error paths ---

  Scenario: Missing equals in variable declaration
    Given the following script:
      """
      %im 0.1
      var badvar
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "missing '='"

  Scenario: Empty variable expression
    Given the following script:
      """
      %im 0.1
      var x =
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "invalid variable declaration"

  Scenario: Missing function parameter list
    Given the following script:
      """
      %im 0.1
      fun noparen = 42
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "missing parameter list"

  Scenario: Missing equals after function parameters
    Given the following script:
      """
      %im 0.1
      fun noeq(x) x * 2
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "missing '='"

  Scenario: Missing equals in type declaration
    Given the following script:
      """
      %im 0.1
      type Bad String
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "missing '='"

  Scenario: Missing base type in type declaration
    Given the following script:
      """
      %im 0.1
      type Bad =
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "missing base type"

  Scenario: Invalid property syntax in type declaration
    Given the following script:
      """
      %im 0.1
      type Bad = String { nocolon }
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "invalid type declaration"

  Scenario: Invalid namespace declaration with no arguments
    Given the following script:
      """
      %im 0.1
      ns
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "unrecognized header directive"

  Scenario: Unrecognized header directive
    Given the following script:
      """
      %im 0.1
      banana split
      output application/json
      ---
      1
      """
    Then running the script should fail with error containing "unrecognized header directive"

  Scenario: Missing header separator via output path
    Given the following script:
      """
      %im 0.1
      output application/json
      42
      """
    Then running the script through the runner output path should fail with error containing "script must have a header"

  Scenario: Invalid output option format
    Given the following script:
      """
      %im 0.1
      output application/json badoption
      ---
      42
      """
    Then running the script should fail with error containing "invalid output option"

  # --- RunFromStringWithContext output path ---

  Scenario: JSON output through full output path
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      { name: "Alice", age: 30 }
      """
    When I run the script through the runner output path
    Then the output should contain "Alice"
    And the output should contain "30"

  Scenario: CSV output through full output path
    Given the following script:
      """
      %im 0.1
      output application/csv
      ---
      [{ name: "Bob", age: 25 }]
      """
    When I run the script through the runner output path
    Then the output should contain "Bob"
    And the output should contain "25"

  # --- Payload input with in-process runner ---

  Scenario: JSON payload with in-process runner
    Given the following JSON input:
      """
      {"key": "value"}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.key
      """
    When I run the script
    Then the output should be:
      """
      "value"
      """

  Scenario: XML payload with in-process runner
    Given the following XML input:
      """
      <root><item>hello</item></root>
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.root.item
      """
    When I run the script
    Then the output should be:
      """
      "hello"
      """

  # --- Combined header declarations ---

  Scenario: All declaration types together via in-process runner
    Given the following script:
      """
      %im 0.1
      var prefix = "Result"
      fun format(label, val) = label ++ ": " ++ val
      type Tag = String
      output application/json
      ---
      format(prefix, "42" as Tag)
      """
    When I run the script
    Then the output should be:
      """
      "Result: 42"
      """

  Scenario: Variable and function with import via in-process runner
    Given the following script:
      """
      %im 0.1
      import double from modules::MathUtils
      var base = 5
      fun quad(x) = double(double(x))
      output application/json
      ---
      quad(base)
      """
    When I run the script
    Then the output should be:
      """
      20
      """

  Scenario: Import, blank lines, comments, var, and function stay ordered through header execution
    Given the following script:
      """
      %im 0.1
      import double from modules::MathUtils

      var base = 10
      // local helper layered on imported function
      fun scale(x) = double(x) + base
      output application/json
      ---
      scale(11)
      """
    When I run the script
    Then the output should be:
      """
      32
      """

  Scenario: Multiline var and multiline fun in one header via unified dispatch
    Given the following script:
      """
      %im 0.1
      var config =
        {
          greeting: "Hello",
          name: "World"
        }
      fun greet(cfg) = cfg.greeting ++ ", " ++ cfg.name ++ "!"
      output application/json
      ---
      greet(config)
      """
    When I run the script
    Then the output should be:
      """
      "Hello, World!"
      """

  Scenario: Multiline fun with balanced braces in header
    Given the following script:
      """
      %im 0.1
      fun buildPair(a, b) = {
        first: a,
        second: b
      }
      var x = 10
      output application/json
      ---
      buildPair(x, x + 5)
      """
    When I run the script
    Then the output should be:
      """
      {"first":10,"second":15}
      """
