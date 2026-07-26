Feature: Complex DataWeave Expressions
  In order to validate composed transformations
  As a developer
  I want to combine multiple language features in a single script

  Scenario: Filter active users and enrich with defaults
    Given the following input content:
      """
      %im 0.1
      output application/json
      var users = [
        {id: 1, name: "Ada", roles: ["admin", "ops"], active: true},
        {id: 2, name: "Bob", roles: ["guest"], active: false},
        {id: 3, name: "Cara", active: true}
      ]
      ---
      users
        filter (u) -> u.active
        map (u, idx) -> {
          userId: u.id,
          label: "$(idx): " ++ (if (u.name == nil) "Unknown" else u.name),
          roleCount: if (u.roles == nil) 0 else sizeOf(u.roles),
          isPrivileged: if (u.roles == nil) false else contains(u.roles, "admin")
        }
      """
    When I run the application with this content
    Then the output should be valid JSON with array length of 2
    And the output should contain "\"userId\":1"
    And the output should contain "\"label\":\"0: Ada\""
    And the output should contain "\"roleCount\":2"
    And the output should contain "\"isPrivileged\":true"
    And the output should contain "\"userId\":3"
    And the output should contain "\"label\":\"1: Cara\""
    And the output should contain "\"roleCount\":0"
    And the output should contain "\"isPrivileged\":false"

  Scenario: Group by type and summarize totals
    Given the following input content:
      """
      %im 0.1
      output application/json
      var tx = [
        {type: "A", amount: 10},
        {type: "B", amount: 4},
        {type: "A", amount: 6},
        {type: "B", amount: 1}
      ]
      ---
      tx
        groupBy (t) -> t.type
        mapObject (v, k) -> [k, {
          count: sizeOf(v),
          total: (v map (item) -> item.amount) reduce (x, acc) -> acc + x
        }]
      """
    When I run the application with this content
    Then the output should contain "\"A\":{\"count\":2,\"total\":16}"
    And the output should contain "\"B\":{\"count\":2,\"total\":5}"

  Scenario: Large range with filter/map/reduce
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      range(10000)
        filter (x) -> mod(x, 2) == 0
        map (x) -> x + 1
        reduce (x, acc) -> acc + x
      """
    When I run the script
    Then the output should be valid JSON with number close to 25000000

  Scenario: Signed scientific notation survives nested default and contains operators
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [-4.4420514987125e+10] contains (-4.4420514987125e+10 default 3.4133172375e+07)
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: Using expression inside map
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [1, 2, 3]
        map (u, idx) -> using(mult = 2, label = "item") {
          value: u * mult,
          label: label ++ "-" ++ idx
        }
      """
    When I run the script
    Then the output should contain "\"value\":2"
    And the output should contain "\"label\":\"item-0\""
    And the output should contain "\"value\":6"
    And the output should contain "\"label\":\"item-2\""
