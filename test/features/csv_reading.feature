Feature: CSV Reading
  In order to process tabular data
  As a developer
  I want to read and parse CSV content

  Scenario: Read a simple CSV string
    Given the following CSV input:
      """
      name,age
      Alice,30
      Bob,25
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      [{"name":"Alice","age":"30"},{"name":"Bob","age":"25"}]
      """

  Scenario: Error on empty header column name
    Given the following CSV input:
      """
      name,,age
      Alice,x,30
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "empty column name"

  Scenario: Error on mismatched column count
    Given the following CSV input:
      """
      name,age
      Alice,30,extra
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "wrong number of fields"

  Scenario: Read empty CSV content
    Given the following CSV input:
      """
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      null
      """

  # Basic Variations
  Scenario: Read CSV with single row
    Given the following CSV input:
      """
      name,age
      Alice,30
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      [{"name":"Alice","age":"30"}]
      """

  Scenario: Read CSV with header only
    Given the following CSV input:
      """
      name,age,city
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should be:
      """
      null
      """

  Scenario: Read CSV with many columns
    Given the following CSV input:
      """
      c1,c2,c3,c4,c5,c6,c7,c8,c9,c10,c11,c12,c13,c14,c15,c16,c17,c18,c19,c20
      v1,v2,v3,v4,v5,v6,v7,v8,v9,v10,v11,v12,v13,v14,v15,v16,v17,v18,v19,v20
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(payload[0])
      """
    When I run the script
    Then the output should be:
      """
      20
      """

  # Special Characters
  Scenario: Read CSV with quoted fields containing commas
    Given the following CSV input:
      """
      name,address
      Alice,"123 Main St, Apt 4"
      Bob,"456 Oak Ave, Suite 200"
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload[0].address
      """
    When I run the script
    Then the output should be:
      """
      "123 Main St, Apt 4"
      """

  Scenario: Read CSV with quoted fields containing newlines
    Given the following CSV input:
      """
      name,comment
      Alice,"This is
      a multi-line
      comment"
      Bob,"Single line"
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload[0].comment
      """
    When I run the script
    Then the output should contain "This is"

  Scenario: Read CSV with escaped quotes within quoted fields
    Given the following CSV input:
      """
      name,quote
      Alice,"She said ""Hello"""
      Bob,"He replied ""Hi there"""
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload[0].quote
      """
    When I run the script
    Then the output should be:
      """
      "She said \"Hello\""
      """

  # Edge Cases
  Scenario: Read CSV with heterogeneous data - missing values
    Given the following CSV input:
      """
      name,age,city
      Alice,30,NYC
      Bob,,
      Carol,25,LA
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    When I run the script
    Then the output should contain "Bob"

  Scenario: Read CSV with Unicode characters
    Given the following CSV input:
      """
      name,greeting
      Alice,Hello
      José,¡Hola!
      李明,你好
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload[2].name
      """
    When I run the script
    Then the output should be:
      """
      "李明"
      """

  Scenario: Read CSV with very long fields
    Given the following CSV input:
      """
      id,description
      1,This is a long CSV field with many characters and words to verify that the parser handles extended text content properly without any issues or truncation during parsing operations.
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(payload[0].description)
      """
    When I run the script
    Then the output should be:
      """
      180
      """

  # Error Handling
  Scenario: Error on CSV row with too many columns
    Given the following CSV input:
      """
      name,age
      Alice,30,extra1,extra2
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "wrong number of fields"

  Scenario: Error on CSV row with too few columns
    Given the following CSV input:
      """
      name,age,city
      Alice,30
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "wrong number of fields"
