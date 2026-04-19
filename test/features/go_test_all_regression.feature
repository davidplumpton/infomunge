Feature: Repo-wide go test command

  Scenario: Scratch helpers under tmp do not break package discovery
    When I run "go test ./..." from the repo root
    Then the output should contain "ok"

  Scenario: Repo-wide go test command stays stable on consecutive runs
    When I run "go test ./..." from the repo root
    Then the output should contain "ok"
    When I run "go test ./..." from the repo root
    Then the output should contain "ok"
