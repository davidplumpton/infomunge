Feature: Assertion matchers runner coverage
  In order to measure cucumber coverage for assertion internals
  As a maintainer
  I want matcher and matcher-wrapper paths exercised in-process

  Scenario: Matcher success paths execute via runner
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [must([], beArray()), must({x: 1}, beObject()), must("", beBlank()), must("abc", equalTo("abc")), must(2, beGreaterThan(1)), must(2, beGreaterThan(2, true)), must(1, beLowerThan(2)), must(1, beLowerThan(1, true)), must("b", beOneOf(["a", "b"])), must("hello", containStr("ell")), must([1, 2, 3], containVal(2)), must("hello", startWith("he")), must("hello", endWith("lo")), must({a: 1, b: 2}, haveSize(2)), must({a: 1}, haveKey("a")), must({a: 1}, haveValue(1)), must([1, 2], eachItem(beNumber())), must([1, "two"], haveItem(beString())), must("x", anyOf([beString(), beNumber()])), must("x", notBe(beNumber())), must("x", notBeNull()), must(3, beNumber()), must(false, beBoolean()), must(null, beNull()), must([], beEmpty())]
      """
    When I run the script
    Then the output should be valid JSON with array length of 25

  Scenario: Compare values rejects incompatible types
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      must(1, beGreaterThan("a"))
      """
    Then running the script should fail with error containing "cannot compare values"

  Scenario: Comparison matchers preserve large integer precision
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      must(9007199254740993, beGreaterThan(9007199254740992))
      """
    When I run the script
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: must rejects non-matcher input
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      must("x", 1)
      """
    Then running the script should fail with error containing "expected matcher or array of matchers"

  Scenario: must rejects invalid matcher array entry
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      must("x", [beString(), 1])
      """
    Then running the script should fail with error containing "matcher at index 1 is not a valid matcher"

  Scenario: eachItem validates argument count
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      eachItem()
      """
    Then running the script should fail with error containing "eachItem() requires exactly 1 argument"

  Scenario: haveItem validates matcher argument type
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      haveItem(1)
      """
    Then running the script should fail with error containing "expected matcher, got int"

  Scenario: anyOf validates matcher array argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      anyOf(beString())
      """
    Then running the script should fail with error containing "expected array of matchers"

  Scenario: notBe validates matcher argument type
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      notBe(1)
      """
    Then running the script should fail with error containing "expected matcher, got int"

  Scenario: matcher constructor wrappers validate arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [beBlank("x"), equalTo(), beGreaterThan(1, "true"), beOneOf("x"), containStr(1), containVal(), startWith(1), endWith(1), haveSize("2"), haveKey(1), haveValue(), notBeNull(1)]
      """
    Then running the script should fail with error containing "beBlank() takes no arguments"
