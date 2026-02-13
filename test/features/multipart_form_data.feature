Feature: Multipart Form Data
  In order to process multipart payloads
  As a developer
  I want to read and write multipart/form-data content

  Scenario: Read multipart payload with repeated and file fields
    Given the following multipart input:
      """
      --boundary123
      Content-Disposition: form-data; name="name"

      Alice
      --boundary123
      Content-Disposition: form-data; name="tags"

      alpha
      --boundary123
      Content-Disposition: form-data; name="tags"

      beta
      --boundary123
      Content-Disposition: form-data; name="upload"; filename="hello.txt"
      Content-Type: text/plain

      Hello file
      --boundary123--
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
      {"name":"Alice","tags":["alpha","beta"],"upload":{"content":"Hello file","contentType":"text/plain","filename":"hello.txt"}}
      """

  Scenario: Write multipart payload from object data
    Given the following script:
      """
      %im 0.1
      output multipart/form-data
      ---
      {
        name: "Alice",
        tags: ["alpha", "beta"],
        upload: {
          filename: "hello.txt",
          contentType: "text/plain",
          content: "Hello file"
        }
      }
      """
    When I run the script
    Then the output should contain "--infomunge-boundary"
    And the output should contain "Content-Disposition: form-data; name=\"name\""
    And the output should contain "Content-Disposition: form-data; name=\"tags\""
    And the output should contain "filename=\"hello.txt\""
    And the output should contain "Content-Type: text/plain"
