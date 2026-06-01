Feature: Repo-wide package discovery command

  Scenario: Scratch helpers under tmp do not break package discovery
    When I run repo-wide package discovery from the repo root
    Then the output should contain "ok"

  Scenario: Repo-wide package discovery stays stable on consecutive runs
    When I run repo-wide package discovery from the repo root
    Then the output should contain "ok"
    When I run repo-wide package discovery from the repo root
    Then the output should contain "ok"
