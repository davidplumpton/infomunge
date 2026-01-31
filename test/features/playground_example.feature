Feature: Playground example script

  Scenario: Example script produces XML report
    Given the following JSON input:
      """
      [
        {"type": "A", "amount": 10},
        {"type": "B", "amount": 4},
        {"type": "A", "amount": 6},
        {"type": "B", "amount": 1}
      ]
      """
    And the following script:
      """
      %im 0.1
      fun sum(nums) = nums reduce (acc, n) -> acc + n
      output application/xml
      ---
      {
        report @(generated: "true"): {
          types: payload
            groupBy (t) -> t.type
            mapObject (v, k) -> [k, {
              count: sizeOf(v),
              total: sum(v map (e) -> e.amount)
            }],
          notes: "Types: " ++ (
            (payload
              groupBy (t) -> t.type
              pluck (v, k) -> k
            ) joinBy ", "
          )
        }
      }
      """
    When I run the script
    Then the output should contain "report"
    And the output should contain "A"
    And the output should contain "B"
