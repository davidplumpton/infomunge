# Auto-generated from intensive testing failure 001
# Property: mutation_determinism
# Seed: -3677
# Shrunk expression: ((trim(now())))
Feature: Auto-generated intensive-testing failures

  Scenario: Auto-generated failure 001 (mutation_determinism)
    Given input payload is:
      """
      null
      """
    When infomunge processes:
      """
      %im 0.1
      output application/json
      ---
      ((trim(now())))
      """
    Then verify expected result or error for property "mutation_determinism"
