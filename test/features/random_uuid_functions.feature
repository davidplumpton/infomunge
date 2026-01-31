Feature: Random and UUID Functions
  In order to generate random values and unique identifiers
  As a developer
  I want to use the random, randomInt, and uuid functions

  Scenario: random takes no arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      random(5)
      """
    When I run the application and it fails
    Then the output should contain "random takes no arguments"

  Scenario: random returns a value
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      random() != null
      """
    When I run the application with this content
    Then the output should be true

  Scenario: randomInt requires exactly one argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      randomInt()
      """
    When I run the application and it fails
    Then the output should contain "randomInt requires exactly 1 argument"

  Scenario: randomInt with too many arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      randomInt(10, 20)
      """
    When I run the application and it fails
    Then the output should contain "randomInt requires exactly 1 argument"

  Scenario: randomInt requires max to be greater than 0
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      randomInt(0)
      """
    When I run the application and it fails
    Then the output should contain "randomInt max must be greater than 0"

  Scenario: randomInt with negative max
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      randomInt(-5)
      """
    When I run the application and it fails
    Then the output should contain "randomInt max must be greater than 0"

  Scenario: randomInt expects a number argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      randomInt("10")
      """
    When I run the application and it fails
    Then the output should contain "randomInt expects a number"

  Scenario: randomInt returns a number
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      randomInt(100) != null
      """
    When I run the application with this content
    Then the output should be true

  Scenario: randomInt with larger values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      randomInt(1000000) != null
      """
    When I run the application with this content
    Then the output should be true

  Scenario: randomInt with float argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      randomInt(10.7) != null
      """
    When I run the application with this content
    Then the output should be true

  Scenario: uuid takes no arguments
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      uuid("custom")
      """
    When I run the application and it fails
    Then the output should contain "uuid takes no arguments"

  Scenario: uuid returns a string
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      uuid() != null
      """
    When I run the application with this content
    Then the output should be true

  Scenario: uuid returns different values on successive calls
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      uuid() != uuid()
      """
    When I run the application with this content
    Then the output should be true

  Scenario: random multiple calls in array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        r1: random(),
        r2: random(),
        r3: random()
      }
      """
    When I run the application with this content
    Then the output should be valid JSON with 3 keys

  Scenario: randomInt multiple calls in array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        x1: randomInt(100),
        x2: randomInt(100),
        x3: randomInt(100)
      }
      """
    When I run the application with this content
    Then the output should be valid JSON with 3 keys

  Scenario: uuid in array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [uuid(), uuid(), uuid()]
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 3
