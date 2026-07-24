Feature: While Loops

  Scenario: Simple while loop that counts up
    Given the following script:
      """
      var i = 0
      ---
      while(i < 5) {
        i = i + 1
      }
      i
      """
    When I run the script
    Then the output should be valid JSON with number: 5

  Scenario: While loop with collection accumulation
    Given the following script:
      """
      var i = 0
      var result = []
      ---
      while(i < 3) {
        result = result + [i]
        i = i + 1
      }
      result
      """
    When I run the script
    Then the output should be valid JSON with array length of 3

  Scenario: While loop with early exit
    Given the following script:
      """
      var i = 0
      ---
      while(i < 100) {
        if(i == 5) break
        i = i + 1
      }
      i
      """
    When I run the script
    Then the output should be valid JSON with number: 5

  Scenario: While loop with continue statement
    Given the following script:
      """
      var i = 0
      var result = []
      ---
      while(i < 5) {
        i = i + 1
        if(i == 3) continue
        result = result + [i]
      }
      result
      """
    When I run the script
    Then the output should be valid JSON with array length of 4

  Scenario: While loop terminates on condition
    Given the following script:
      """
      var i = 0
      ---
      while(i < 10 && i < 5) {
        i = i + 1
      }
      i
      """
    When I run the script
    Then the output should be valid JSON with number: 5

  Scenario: Infinite loop is caught by timeout
    Given the following script:
      """
      var i = 0
      ---
      while(true) {
        i = i + 1
      }
      i
      """
    And I set the execution timeout to 1 seconds
    Then running the script should fail with error containing "timed out"
