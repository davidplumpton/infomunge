Feature: Date Formatting
  As a user
  I want to format dates and times using the 'as' operator
  So that I can convert them to strings with specific formats

  Scenario: Format ISO DateTime to custom string
    Given the input "2023-01-01T12:00:00Z"
    When I evaluate "payload as String {format: 'yyyy/MM/dd HH:mm:ss'}"
    Then the result should be "2023/01/01 12:00:00"

  Scenario: Format simple Date to custom string
    Given the input "2023-05-20"
    When I evaluate "payload as String {format: 'dd-MM-yyyy'}"
    Then the result should be "20-05-2023"

  Scenario: Format DateTime with space to custom string
    Given the input "2023-12-25 10:30:00"
    When I evaluate "payload as String {format: 'yyyy.MM.dd'}"
    Then the result should be "2023.12.25"

  Scenario: Format Time to custom string
    Given the input "15:30:45"
    When I evaluate "payload as String {format: 'hh:mm a'}"
    Then the result should be "03:30 PM"

  Scenario: Format DateTime with fractional seconds and ISO timezone
    Given the input "2023-01-01T12:00:00Z"
    When I evaluate "payload as String {format: 'yyyy-MM-dd HH:mm:ss.SSSXXX'}"
    Then the result should be "2023-01-01 12:00:00.000Z"

  Scenario: Format number with format (existing functionality check)
    Given the input 123.456
    When I evaluate "payload as String {format: '#.0'}"
    Then the result should be "123.5"

  Scenario: Format now()
    Given the script:
      """
      %dw 2.0
      output application/json
      ---
      now() as String {format: "yyyy"}
      """
    When I execute the script
    Then the result should match "\d{4}"
