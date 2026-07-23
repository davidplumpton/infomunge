Feature: Safe server binding
  In order to avoid exposing script evaluation unintentionally
  As a server operator
  I want unauthenticated server mode to remain local

  Scenario: Server mode defaults to a loopback listen address
    When I inspect the application help
    Then the server listen default should be "127.0.0.1:8080"

  Scenario: Unauthenticated server mode rejects a wildcard listen address
    When I run the application with arguments "--server --listen :8080" and it fails
    Then the output should contain "server mode without --api-key must listen on a loopback address"
