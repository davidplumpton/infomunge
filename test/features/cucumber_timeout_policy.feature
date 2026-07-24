Feature: Test workflow policy

  Scenario: The unit target covers the repository unit suite
    Then the Makefile unit target should cover representative repository unit packages

  Scenario: Cucumber entry points use five minute Go test timeouts
    Then the Makefile cucumber targets should use a 5 minute Go test timeout
    And the repo-wide go test regression step should use a 5 minute Go test timeout
    And the authoritative testing docs should show portable cucumber commands with a 5 minute Go test timeout
    And the authoritative testing docs should show bounded repo-wide package test commands
    And the default mutation corpus test should be skipped outside soak mode
