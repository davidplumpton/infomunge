package evaluator

func dateBuiltinSpecs() []BuiltinSpec {
	return []BuiltinSpec{
		regularBuiltinSpec("now", builtinCategoryDates, exactArity(0), callBuiltinNow, "current date time"),
		regularBuiltinSpec("daysBetween", builtinCategoryDates, exactArity(2), callBuiltinDaysBetween, "days between dates"),
		regularBuiltinSpec("isLeapYear", builtinCategoryDates, exactArity(1), callBuiltinIsLeapYear, "leap year check"),
		regularBuiltinSpec("today", builtinCategoryDates, exactArity(0), callBuiltinToday, "current date"),
		regularBuiltinSpec("tomorrow", builtinCategoryDates, exactArity(0), callBuiltinTomorrow, "tomorrow date"),
		regularBuiltinSpec("yesterday", builtinCategoryDates, exactArity(0), callBuiltinYesterday, "yesterday date"),
		regularBuiltinSpec("date", builtinCategoryDates, exactArity(3), callBuiltinDate, "date constructor"),
		withArityMessages(
			regularBuiltinSpec("time", builtinCategoryDates, rangeArity(3, 4), callBuiltinTime, "time constructor"),
			"time requires 3-4 arguments: hour, minutes, seconds[, timezone]",
			"time requires 3-4 arguments: hour, minutes, seconds[, timezone]",
		),
		withArityMessages(
			regularBuiltinSpec("dateTime", builtinCategoryDates, rangeArity(6, 7), callBuiltinDateTime, "date-time constructor"),
			"dateTime requires 6-7 arguments: year, month, day, hour, minutes, seconds[, timezone]",
			"dateTime requires 6-7 arguments: year, month, day, hour, minutes, seconds[, timezone]",
		),
		regularBuiltinSpec("localDateTime", builtinCategoryDates, exactArity(6), callBuiltinLocalDateTime, "local date-time constructor"),
		regularBuiltinSpec("localTime", builtinCategoryDates, exactArity(3), callBuiltinLocalTime, "local time constructor"),
		regularBuiltinSpec("atBeginningOfDay", builtinCategoryDates, exactArity(1), callBuiltinAtBeginningOfDay, "beginning of day"),
		regularBuiltinSpec("atBeginningOfHour", builtinCategoryDates, exactArity(1), callBuiltinAtBeginningOfHour, "beginning of hour"),
		regularBuiltinSpec("atBeginningOfMonth", builtinCategoryDates, exactArity(1), callBuiltinAtBeginningOfMonth, "beginning of month"),
		regularBuiltinSpec("atBeginningOfWeek", builtinCategoryDates, exactArity(1), callBuiltinAtBeginningOfWeek, "beginning of week"),
		regularBuiltinSpec("atBeginningOfYear", builtinCategoryDates, exactArity(1), callBuiltinAtBeginningOfYear, "beginning of year"),
		regularBuiltinSpec("dayOfWeek", builtinCategoryDates, exactArity(1), callBuiltinDayOfWeek, "day of week"),
		regularBuiltinSpec("dayOfYear", builtinCategoryDates, exactArity(1), callBuiltinDayOfYear, "day of year"),
	}
}
