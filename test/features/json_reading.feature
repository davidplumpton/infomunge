Feature: JSON Reading
  In order to process JSON data
  As a developer
  I want to read and parse JSON content

  Scenario: Read a simple JSON string
    Given the following JSON input:
      """
      {"name": "Alice", "age": 30}
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
      {"age":30,"name":"Alice"}
      """

  # Basic Types - JSON primitives and simple structures
  Scenario: Read JSON array of primitives
    Given the following JSON input:
      """
      [1, 2, 3, 4, 5]
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
      [1,2,3,4,5]
      """

  Scenario: Read JSON string value
    Given the following JSON input:
      """
      "Hello, World!"
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
      "Hello, World!"
      """

  Scenario: Read JSON number value
    Given the following JSON input:
      """
      42.5
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
      42.5
      """

  Scenario: Read JSON boolean value
    Given the following JSON input:
      """
      true
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
      true
      """

  Scenario: Read JSON null value
    Given the following JSON input:
      """
      null
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

  # Complex Structures
  Scenario: Read deeply nested JSON object
    Given the following JSON input:
      """
      {"level1": {"level2": {"level3": {"level4": "deep value"}}}}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.level1.level2.level3.level4
      """
    When I run the script
    Then the output should be:
      """
      "deep value"
      """

  Scenario: Read array of objects
    Given the following JSON input:
      """
      [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload[0].name
      """
    When I run the script
    Then the output should be:
      """
      "Alice"
      """

  Scenario: Read object with mixed array and object values
    Given the following JSON input:
      """
      {"users": [{"name": "Alice"}, {"name": "Bob"}], "meta": {"count": 2}}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      { count: payload.meta.count, first_user: payload.users[0].name }
      """
    When I run the script
    Then the output should be:
      """
      {"count":2,"first_user":"Alice"}
      """

  Scenario: Read empty JSON object and array
    Given the following JSON input:
      """
      {"empty_obj": {}, "empty_arr": []}
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
      {"empty_arr":[],"empty_obj":{}}
      """

  # Edge Cases
  Scenario: Read JSON with Unicode characters
    Given the following JSON input:
      """
      {"greeting": "Hello, 世界! 🌍"}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.greeting
      """
    When I run the script
    Then the output should be:
      """
      "Hello, 世界! 🌍"
      """

  Scenario: Read JSON with escaped special characters
    Given the following JSON input:
      """
      {"path": "C:\\Users\\Alice\\file.txt", "quote": "He said \"Hello\""}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload.path
      """
    When I run the script
    Then the output should be:
      """
      "C:\\Users\\Alice\\file.txt"
      """

  Scenario: Read JSON with numeric precision
    Given the following JSON input:
      """
      {"large_int": 9007199254740991, "decimal": 0.1, "negative": -42.5}
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
      {"decimal":0.1,"large_int":9007199254740991,"negative":-42.5}
      """

  # Error Handling
  Scenario: Error on malformed JSON - missing closing brace
    Given the following JSON input:
      """
      {"name": "Alice", "age": 30
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "JSON parse error"
    And running the script should fail with error containing "ValidationError"

  Scenario: Error on malformed JSON - trailing comma
    Given the following JSON input:
      """
      {"name": "Alice", "age": 30,}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "JSON parse error"

  Scenario: Error on malformed JSON - invalid key format
    Given the following JSON input:
      """
      {name: "Alice", age: 30}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      payload
      """
    Then running the script should fail with error containing "JSON parse error"

  # Large Data
  Scenario: Read JSON array with 100+ elements
    Given the following JSON input:
      """
      [0,1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16,17,18,19,20,21,22,23,24,25,26,27,28,29,30,31,32,33,34,35,36,37,38,39,40,41,42,43,44,45,46,47,48,49,50,51,52,53,54,55,56,57,58,59,60,61,62,63,64,65,66,67,68,69,70,71,72,73,74,75,76,77,78,79,80,81,82,83,84,85,86,87,88,89,90,91,92,93,94,95,96,97,98,99,100]
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(payload)
      """
    When I run the script
    Then the output should be:
      """
      101
      """

  Scenario: Read JSON object with 50+ keys
    Given the following JSON input:
      """
      {"k1":"v1","k2":"v2","k3":"v3","k4":"v4","k5":"v5","k6":"v6","k7":"v7","k8":"v8","k9":"v9","k10":"v10","k11":"v11","k12":"v12","k13":"v13","k14":"v14","k15":"v15","k16":"v16","k17":"v17","k18":"v18","k19":"v19","k20":"v20","k21":"v21","k22":"v22","k23":"v23","k24":"v24","k25":"v25","k26":"v26","k27":"v27","k28":"v28","k29":"v29","k30":"v30","k31":"v31","k32":"v32","k33":"v33","k34":"v34","k35":"v35","k36":"v36","k37":"v37","k38":"v38","k39":"v39","k40":"v40","k41":"v41","k42":"v42","k43":"v43","k44":"v44","k45":"v45","k46":"v46","k47":"v47","k48":"v48","k49":"v49","k50":"v50"}
      """
    And the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(payload)
      """
    When I run the script
    Then the output should be:
      """
      50
      """
