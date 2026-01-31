Feature: Assertion Matchers
  In order to validate data
  As a developer
  I want to use assertion matchers with the must function

  # Type matchers - beArray
  Scenario: beArray with array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], beArray())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beArray with non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", beArray())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Type matchers - beObject
  Scenario: beObject with object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must({a: 1}, beObject())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beObject with non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2], beObject())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Type matchers - beString
  Scenario: beString with string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", beString())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beString with non-string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(42, beString())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Type matchers - beNumber
  Scenario: beNumber with integer
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(42, beNumber())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beNumber with float
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(3.14, beNumber())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beNumber with non-number fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("42", beNumber())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Type matchers - beBoolean
  Scenario: beBoolean with true
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(true, beBoolean())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beBoolean with false
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(false, beBoolean())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beBoolean with non-boolean fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(1, beBoolean())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Type matchers - beNull
  Scenario: beNull with null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(null, beNull())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beNull with non-null fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", beNull())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Type matchers - beEmpty
  Scenario: beEmpty with empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("", beEmpty())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beEmpty with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([], beEmpty())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beEmpty with empty object
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must({}, beEmpty())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beEmpty with non-empty string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", beEmpty())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  Scenario: beEmpty with non-empty array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1], beEmpty())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Type matchers - beBlank
  Scenario: beBlank with empty string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("", beBlank())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beBlank with whitespace string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("   ", beBlank())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beBlank with non-blank string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", beBlank())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  Scenario: beBlank with non-string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(42, beBlank())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Comparison matchers - equalTo
  Scenario: equalTo with equal numbers
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(42, equalTo(42))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: equalTo with equal strings
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", equalTo("hello"))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: equalTo with equal arrays
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], equalTo([1, 2, 3]))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: equalTo with unequal values fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(42, equalTo(43))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Comparison matchers - beGreaterThan
  Scenario: beGreaterThan with greater number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(10, beGreaterThan(5))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beGreaterThan with equal number fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(5, beGreaterThan(5))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  Scenario: beGreaterThan inclusive with equal number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(5, beGreaterThan(5, true))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beGreaterThan with lesser number fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(3, beGreaterThan(5))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Comparison matchers - beLowerThan
  Scenario: beLowerThan with lower number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(3, beLowerThan(5))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beLowerThan with equal number fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(5, beLowerThan(5))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  Scenario: beLowerThan inclusive with equal number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(5, beLowerThan(5, true))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beLowerThan with greater number fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(10, beLowerThan(5))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Comparison matchers - beOneOf
  Scenario: beOneOf with matching value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("b", beOneOf(["a", "b", "c"]))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: beOneOf with non-matching value fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("d", beOneOf(["a", "b", "c"]))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Content matchers - containStr
  Scenario: containStr with matching substring
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello world", containStr("world"))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: containStr with non-matching substring fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello world", containStr("foo"))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  Scenario: containStr with non-string fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(42, containStr("4"))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Content matchers - containVal
  Scenario: containVal with matching value in array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], containVal(2))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: containVal with non-matching value fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], containVal(4))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  Scenario: containVal with non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", containVal("h"))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Content matchers - startWith
  Scenario: startWith with matching prefix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello world", startWith("hello"))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: startWith with non-matching prefix fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello world", startWith("world"))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Content matchers - endWith
  Scenario: endWith with matching suffix
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello world", endWith("world"))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: endWith with non-matching suffix fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello world", endWith("hello"))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Content matchers - haveSize
  Scenario: haveSize with array of correct size
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], haveSize(3))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: haveSize with string of correct length
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", haveSize(5))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: haveSize with object of correct size
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must({a: 1, b: 2}, haveSize(2))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: haveSize with incorrect size fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], haveSize(5))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Content matchers - haveKey
  Scenario: haveKey with existing key
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must({name: "Alice", age: 30}, haveKey("name"))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: haveKey with missing key fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must({name: "Alice"}, haveKey("age"))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  Scenario: haveKey with non-object fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2], haveKey("0"))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Content matchers - haveValue
  Scenario: haveValue with existing value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must({name: "Alice", age: 30}, haveValue("Alice"))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: haveValue with missing value fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must({name: "Alice"}, haveValue("Bob"))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Null handling - notBeNull
  Scenario: notBeNull with non-null value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", notBeNull())
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: notBeNull with null fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(null, notBeNull())
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Composite matchers - eachItem
  Scenario: eachItem with all items matching
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], eachItem(beNumber()))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: eachItem with non-matching item fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, "two", 3], eachItem(beNumber()))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  Scenario: eachItem with non-array fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", eachItem(beString()))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Composite matchers - haveItem
  Scenario: haveItem with at least one matching item
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, "two", 3], haveItem(beString()))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: haveItem with no matching items fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], haveItem(beString()))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Composite matchers - anyOf
  Scenario: anyOf with first matcher matching
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", anyOf([beString(), beNumber()]))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: anyOf with second matcher matching
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(42, anyOf([beString(), beNumber()]))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: anyOf with no matchers matching fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must(true, anyOf([beString(), beNumber()]))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Composite matchers - notBe
  Scenario: notBe negates matcher that fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", notBe(beNumber()))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: notBe negates matcher that passes fails
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", notBe(beString()))
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Multiple matchers with must
  Scenario: must with array of matchers all passing
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", [beString(), haveSize(5)])
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: must with array of matchers one failing
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must("hello", [beString(), haveSize(10)])
      """
    When I run the application and it fails
    Then the output should contain "assertion failed"

  # Complex nested matchers
  Scenario: eachItem with equalTo matcher
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([5, 5, 5], eachItem(equalTo(5)))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: haveItem with beGreaterThan matcher
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 5, 10], haveItem(beGreaterThan(7)))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """

  Scenario: notBe with beEmpty matcher
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      must([1, 2, 3], notBe(beEmpty()))
      """
    When I run the application with this content
    Then the output should be:
      """
      {"Success":true,"Message":""}
      """
