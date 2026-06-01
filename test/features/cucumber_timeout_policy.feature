Feature: Cucumber timeout policy

  Scenario: Cucumber entry points use five minute Go test timeouts
    Then the Makefile cucumber targets should use a 5 minute Go test timeout
    And the repo-wide go test regression step should use a 5 minute Go test timeout
    And the testing docs should show cucumber commands with a 5 minute Go test timeout
