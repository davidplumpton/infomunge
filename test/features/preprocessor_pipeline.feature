Feature: Preprocessor pipeline
  In order to keep syntax rewrites consistent
  As a developer
  I want the public runner to apply the full preprocessing pipeline

  Scenario: Full preprocessing path applies pre and post rewrite stages
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      {
        name: ({user: {name: null}}).user.name default "missing",
        matched: "Ada" matches /A.*/,
      } // trailing comment
      """
    When I run the script
    Then the output should be:
      """
      {"matched":true,"name":"missing"}
      """
