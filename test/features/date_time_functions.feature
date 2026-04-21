Feature: Date/Time Functions
  In order to work with dates and times
  As a developer
  I want to use the now, daysBetween, and isLeapYear functions

  # Now function tests
  Scenario: now returns RFC3339 timestamp with UTC timezone
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      now()
      """
    When I run the script
    Then the output should be a valid RFC3339 timestamp
    And the output should have UTC offset minutes 0

  Scenario: now returns a string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(now())
      """
    When I run the script
    Then the output should be:
      """
      "String"
      """

  Scenario: now takes no arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      now(1)
      """
    Then running the script should fail with error containing "now takes no arguments"

  # DaysBetween function tests
  Scenario: daysBetween with same dates
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      0
      """

  Scenario: daysBetween with one day difference
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-01-02T00:00:00Z", "2024-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  Scenario: daysBetween with multiple days
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-01-10T00:00:00Z", "2024-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      9
      """

  Scenario: daysBetween with negative days
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-01-01T00:00:00Z", "2024-01-10T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      -9
      """

  Scenario: daysBetween across months
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-02-01T00:00:00Z", "2024-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      31
      """

  Scenario: daysBetween across years (leap year 2024)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2025-01-01T00:00:00Z", "2024-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      366
      """

  Scenario: daysBetween across years (non-leap year 2023)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-01-01T00:00:00Z", "2023-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      365
      """

  Scenario: daysBetween with time components (hours rounds to 1)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-01-01T12:00:00Z", "2024-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  Scenario: daysBetween requires two arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-01-01T00:00:00Z")
      """
    Then running the script should fail with error containing "requires exactly 2 arguments"

  Scenario: daysBetween with non-string date fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween(123, "2024-01-01T00:00:00Z")
      """
    Then running the script should fail with error containing "expects first argument to be a date string"

  Scenario: daysBetween with invalid date format fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-01-01", "2024-01-02T00:00:00Z")
      """
    Then running the script should fail with error containing "invalid date format"

  Scenario: daysBetween with both arguments non-string fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween(123, 456)
      """
    Then running the script should fail with error containing "expects first argument to be a date string"

  # IsLeapYear function tests
  Scenario: isLeapYear with leap year 2024
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("2024-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: isLeapYear with non-leap year 2023
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("2023-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: isLeapYear with leap year 2000
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("2000-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: isLeapYear with non-leap year 1900
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("1900-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: isLeapYear with leap year 2000 using Feb 29 date
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("2000-02-29T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: isLeapYear with leap year 2004
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("2004-12-31T23:59:59Z")
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: isLeapYear with non-leap year 2001
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("2001-06-15T12:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: isLeapYear with non-leap year 2100
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("2100-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: isLeapYear with leap year 1996
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("1996-02-29T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: isLeapYear with non-leap year 1997
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("1997-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: isLeapYear requires exactly one argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear()
      """
    Then running the script should fail with error containing "requires exactly 1 argument"

  Scenario: isLeapYear with integer leap year 2024
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear(2024)
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: isLeapYear with integer non-leap year 2023
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear(2023)
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: isLeapYear with integer century leap year 2000
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear(2000)
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: isLeapYear with integer century non-leap year 1900
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear(1900)
      """
    When I run the script
    Then the output should be:
      """
      false
      """

  Scenario: isLeapYear with invalid date format fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear("2024-13-01")
      """
    Then running the script should fail with error containing "invalid date format"

  # Combined date/time scenarios
  Scenario: check if current year is leap year
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      isLeapYear(now())
      """
    When I run the script
    Then the output should be valid JSON boolean

  Scenario: calculate days between two recent dates (leap year 2024)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween("2024-12-31T23:59:59Z", "2024-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      366
      """

  Scenario: use now in date comparison
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      daysBetween(now(), "2024-01-01T00:00:00Z") > 0
      """
    When I run the script
    Then the output should be:
      """
      true
      """

  Scenario: validate leap year for different centuries
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      [
        isLeapYear("1600-01-01T00:00:00Z"),
        isLeapYear("1700-01-01T00:00:00Z"),
        isLeapYear("1800-01-01T00:00:00Z"),
        isLeapYear("1900-01-01T00:00:00Z"),
        isLeapYear("2000-01-01T00:00:00Z")
      ]
      """
    When I run the script
    Then the output should contain "true"
    And the output should contain "false"

  # Today function tests
  Scenario: today returns a date string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(today())
      """
    When I run the script
    Then the output should be:
      """
      "String"
      """

  Scenario: today returns YYYY-MM-DD format
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      sizeOf(today())
      """
    When I run the script
    Then the output should be:
      """
      10
      """

  Scenario: today takes no arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      today(1)
      """
    Then running the script should fail with error containing "today takes no arguments"

  # Tomorrow function tests
  Scenario: tomorrow returns a date string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(tomorrow())
      """
    When I run the script
    Then the output should be:
      """
      "String"
      """

  Scenario: tomorrow takes no arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      tomorrow(1)
      """
    Then running the script should fail with error containing "tomorrow takes no arguments"

  # Yesterday function tests
  Scenario: yesterday returns a date string
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      typeOf(yesterday())
      """
    When I run the script
    Then the output should be:
      """
      "String"
      """

  Scenario: yesterday takes no arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      yesterday(1)
      """
    Then running the script should fail with error containing "yesterday takes no arguments"

  # Date constructor tests
  Scenario: date creates a date from year, month, day
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      date(2025, 12, 25)
      """
    When I run the script
    Then the output should be:
      """
      "2025-12-25"
      """

  Scenario: date with single digit month and day
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      date(2025, 1, 5)
      """
    When I run the script
    Then the output should be:
      """
      "2025-01-05"
      """

  Scenario: date requires exactly 3 arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      date(2025, 12)
      """
    Then running the script should fail with error containing "date requires exactly 3 arguments"

  # DateTime constructor tests
  Scenario: dateTime creates a datetime with timezone
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dateTime(2025, 12, 25, 10, 30, 0)
      """
    When I run the script
    Then the output should be:
      """
      "2025-12-25T10:30:00Z"
      """

  Scenario: dateTime requires 6-7 arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dateTime(2025, 12, 25)
      """
    Then running the script should fail with error containing "dateTime requires 6-7 arguments"

  # LocalDateTime constructor tests
  Scenario: localDateTime creates a datetime without timezone
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      localDateTime(2025, 12, 25, 10, 30, 45)
      """
    When I run the script
    Then the output should be:
      """
      "2025-12-25T10:30:45"
      """

  Scenario: localDateTime requires exactly 6 arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      localDateTime(2025, 12, 25, 10, 30)
      """
    Then running the script should fail with error containing "localDateTime requires exactly 6 arguments"

  # LocalTime constructor tests
  Scenario: localTime creates a time without timezone
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      localTime(14, 30, 45)
      """
    When I run the script
    Then the output should be:
      """
      "14:30:45"
      """

  Scenario: localTime requires exactly 3 arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      localTime(14, 30)
      """
    Then running the script should fail with error containing "localTime requires exactly 3 arguments"

  # Time constructor tests
  Scenario: time creates a time with UTC timezone
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      time(14, 30, 45)
      """
    When I run the script
    Then the output should be:
      """
      "14:30:45Z"
      """

  Scenario: time requires 3-4 arguments
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      time(14)
      """
    Then running the script should fail with error containing "time requires 3-4 arguments"

  # atBeginningOfDay tests
  Scenario: atBeginningOfDay resets time to midnight
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfDay("2025-12-25T14:30:45Z")
      """
    When I run the script
    Then the output should be:
      """
      "2025-12-25T00:00:00Z"
      """

  Scenario: atBeginningOfDay requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfDay()
      """
    Then running the script should fail with error containing "atBeginningOfDay requires exactly 1 argument"

  # atBeginningOfHour tests
  Scenario: atBeginningOfHour resets to start of hour
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfHour("2025-12-25T14:30:45Z")
      """
    When I run the script
    Then the output should be:
      """
      "2025-12-25T14:00:00Z"
      """

  Scenario: atBeginningOfHour requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfHour()
      """
    Then running the script should fail with error containing "atBeginningOfHour requires exactly 1 argument"

  # atBeginningOfMonth tests
  Scenario: atBeginningOfMonth resets to first day of month
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfMonth("2025-12-25T14:30:45Z")
      """
    When I run the script
    Then the output should be:
      """
      "2025-12-01T00:00:00Z"
      """

  Scenario: atBeginningOfMonth requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfMonth()
      """
    Then running the script should fail with error containing "atBeginningOfMonth requires exactly 1 argument"

  # atBeginningOfWeek tests
  Scenario: atBeginningOfWeek resets to Sunday
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfWeek("2025-12-25T14:30:45Z")
      """
    When I run the script
    Then the output should be:
      """
      "2025-12-21T00:00:00Z"
      """

  Scenario: atBeginningOfWeek requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfWeek()
      """
    Then running the script should fail with error containing "atBeginningOfWeek requires exactly 1 argument"

  # atBeginningOfYear tests
  Scenario: atBeginningOfYear resets to first day of year
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfYear("2025-12-25T14:30:45Z")
      """
    When I run the script
    Then the output should be:
      """
      "2025-01-01T00:00:00Z"
      """

  Scenario: atBeginningOfYear requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      atBeginningOfYear()
      """
    Then running the script should fail with error containing "atBeginningOfYear requires exactly 1 argument"

  # dayOfWeek function tests
  Scenario: dayOfWeek returns 1 for Monday
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("2025-01-20T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  Scenario: dayOfWeek returns 2 for Tuesday
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("2025-01-21T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      2
      """

  Scenario: dayOfWeek returns 3 for Wednesday
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("2025-01-22T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      3
      """

  Scenario: dayOfWeek returns 4 for Thursday
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("2025-01-23T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      4
      """

  Scenario: dayOfWeek returns 5 for Friday
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("2025-01-24T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      5
      """

  Scenario: dayOfWeek returns 6 for Saturday
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("2025-01-25T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      6
      """

  Scenario: dayOfWeek returns 7 for Sunday
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("2025-01-26T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      7
      """

  Scenario: dayOfWeek works with date-only format
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("2025-01-20")
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  Scenario: dayOfWeek requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek()
      """
    Then running the script should fail with error containing "dayOfWeek requires exactly 1 argument"

  Scenario: dayOfWeek with non-string date fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek(2025)
      """
    Then running the script should fail with error containing "dayOfWeek expects a date string"

  Scenario: dayOfWeek with invalid date format fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfWeek("not-a-date")
      """
    Then running the script should fail with error containing "invalid date format"

  # dayOfYear function tests
  Scenario: dayOfYear returns 1 for January 1st
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear("2025-01-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      1
      """

  Scenario: dayOfYear returns 32 for February 1st
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear("2025-02-01T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      32
      """

  Scenario: dayOfYear returns 365 for December 31st (non-leap year)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear("2025-12-31T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      365
      """

  Scenario: dayOfYear returns 366 for December 31st (leap year)
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear("2024-12-31T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      366
      """

  Scenario: dayOfYear returns 60 for Feb 29 in leap year
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear("2024-02-29T00:00:00Z")
      """
    When I run the script
    Then the output should be:
      """
      60
      """

  Scenario: dayOfYear works with date-only format
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear("2025-01-15")
      """
    When I run the script
    Then the output should be:
      """
      15
      """

  Scenario: dayOfYear requires exactly 1 argument
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear()
      """
    Then running the script should fail with error containing "dayOfYear requires exactly 1 argument"

  Scenario: dayOfYear with non-string date fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear(2025)
      """
    Then running the script should fail with error containing "dayOfYear expects a date string"

  Scenario: dayOfYear with invalid date format fails
    Given the following script:
      """
      %im 0.1
      output application/json
      ---
      dayOfYear("not-a-date")
      """
    Then running the script should fail with error containing "invalid date format"
