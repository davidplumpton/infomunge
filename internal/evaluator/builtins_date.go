package evaluator

import (
	"fmt"
	"go/ast"
	"math"
	"time"

	unifiederrors "infomunge/internal/errors"
)

// callBuiltinNow implements the now() function.
func callBuiltinNow(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireNoArgs(args, "now", e); err != nil {
		return nil, err
	}
	return time.Now().UTC().Format(time.RFC3339Nano), nil
}

// callBuiltinDaysBetween implements the daysBetween(date1, date2) function.
func callBuiltinDaysBetween(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 2, "daysBetween requires exactly 2 arguments", e); err != nil {
		return nil, err
	}
	date1Str, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("daysBetween expects first argument to be a date string, got %T", args[0]), e.Pos())
	}
	date2Str, ok := args[1].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("daysBetween expects second argument to be a date string, got %T", args[1]), e.Pos())
	}

	// Parse dates (try ISO 8601 format)
	date1, err := time.Parse(time.RFC3339, date1Str)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("daysBetween: invalid date format for date1: %s", err), e.Pos())
	}

	date2, err := time.Parse(time.RFC3339, date2Str)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("daysBetween: invalid date format for date2: %s", err), e.Pos())
	}

	// Calculate days between
	diff := date1.Sub(date2)
	days := diff.Hours() / HoursPerDay

	return math.Round(days), nil
}

// callBuiltinIsLeapYear implements the isLeapYear(dateOrYear) function.
// Accepts either an integer year or a date string.
func callBuiltinIsLeapYear(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "isLeapYear requires exactly 1 argument", e); err != nil {
		return nil, err
	}

	var year int

	switch v := args[0].(type) {
	case int:
		year = v
	case float64:
		year = int(v)
	case string:
		// Parse date (try ISO 8601 format)
		date, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return nil, newPosError(fmt.Sprintf("isLeapYear: invalid date format: %s", err), e.Pos())
		}
		year = date.Year()
	default:
		return nil, newPosError(fmt.Sprintf("isLeapYear expects a date string or integer year, got %T", args[0]), e.Pos())
	}

	isLeap := (year%LeapYearDivisor4 == 0 && year%LeapYearDivisor100 != 0) || (year%LeapYearDivisor400 == 0)

	return isLeap, nil
}

// callBuiltinToday implements the today() function.
func callBuiltinToday(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireNoArgs(args, "today", e); err != nil {
		return nil, err
	}
	return time.Now().Format("2006-01-02"), nil
}

// callBuiltinTomorrow implements the tomorrow() function.
func callBuiltinTomorrow(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireNoArgs(args, "tomorrow", e); err != nil {
		return nil, err
	}
	return time.Now().AddDate(0, 0, 1).Format("2006-01-02"), nil
}

// callBuiltinYesterday implements the yesterday() function.
func callBuiltinYesterday(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireNoArgs(args, "yesterday", e); err != nil {
		return nil, err
	}
	return time.Now().AddDate(0, 0, -1).Format("2006-01-02"), nil
}

// callBuiltinDate implements the date(year, month, day) function.
func callBuiltinDate(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 3, "date requires exactly 3 arguments: year, month, day", e); err != nil {
		return nil, err
	}

	year, err := toInt(args[0], "date", e)
	if err != nil {
		return nil, err
	}
	month, err := toInt(args[1], "date", e)
	if err != nil {
		return nil, err
	}
	day, err := toInt(args[2], "date", e)
	if err != nil {
		return nil, err
	}

	d := time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
	return d.Format("2006-01-02"), nil
}

// callBuiltinTime implements the time(hour, minutes, seconds, timezone) function.
func callBuiltinTime(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 3 || len(args) > 4 {
		return nil, newPosError("time requires 3-4 arguments: hour, minutes, seconds[, timezone]", e.Pos())
	}

	hour, err := toInt(args[0], "time", e)
	if err != nil {
		return nil, err
	}
	minute, err := toInt(args[1], "time", e)
	if err != nil {
		return nil, err
	}
	second, err := toInt(args[2], "time", e)
	if err != nil {
		return nil, err
	}

	loc := time.UTC
	if len(args) == 4 {
		tzStr, ok := args[3].(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("time expects timezone to be a string, got %T", args[3]), e.Pos())
		}
		parsedLoc, err := time.LoadLocation(tzStr)
		if err != nil {
			// Try parsing as offset like "+05:00" or "-08:00"
			t, err2 := time.Parse("-07:00", tzStr)
			if err2 != nil {
				return nil, newPosError(fmt.Sprintf("time: invalid timezone: %s", tzStr), e.Pos())
			}
			_, offset := t.Zone()
			loc = time.FixedZone(tzStr, offset)
		} else {
			loc = parsedLoc
		}
	}

	t := time.Date(0, 1, 1, hour, minute, second, 0, loc)
	return t.Format("15:04:05Z07:00"), nil
}

// callBuiltinDateTime implements the dateTime(year, month, day, hour, minutes, seconds, timezone) function.
func callBuiltinDateTime(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if len(args) < 6 || len(args) > 7 {
		return nil, newPosError("dateTime requires 6-7 arguments: year, month, day, hour, minutes, seconds[, timezone]", e.Pos())
	}

	year, err := toInt(args[0], "dateTime", e)
	if err != nil {
		return nil, err
	}
	month, err := toInt(args[1], "dateTime", e)
	if err != nil {
		return nil, err
	}
	day, err := toInt(args[2], "dateTime", e)
	if err != nil {
		return nil, err
	}
	hour, err := toInt(args[3], "dateTime", e)
	if err != nil {
		return nil, err
	}
	minute, err := toInt(args[4], "dateTime", e)
	if err != nil {
		return nil, err
	}
	second, err := toInt(args[5], "dateTime", e)
	if err != nil {
		return nil, err
	}

	loc := time.UTC
	if len(args) == 7 {
		tzStr, ok := args[6].(string)
		if !ok {
			return nil, newPosError(fmt.Sprintf("dateTime expects timezone to be a string, got %T", args[6]), e.Pos())
		}
		parsedLoc, err := time.LoadLocation(tzStr)
		if err != nil {
			// Try parsing as offset
			t, err2 := time.Parse("-07:00", tzStr)
			if err2 != nil {
				return nil, newPosError(fmt.Sprintf("dateTime: invalid timezone: %s", tzStr), e.Pos())
			}
			_, offset := t.Zone()
			loc = time.FixedZone(tzStr, offset)
		} else {
			loc = parsedLoc
		}
	}

	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, loc)
	return t.Format(time.RFC3339), nil
}

// callBuiltinLocalDateTime implements the localDateTime(year, month, day, hour, minutes, seconds) function.
func callBuiltinLocalDateTime(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 6, "localDateTime requires exactly 6 arguments: year, month, day, hour, minutes, seconds", e); err != nil {
		return nil, err
	}

	year, err := toInt(args[0], "localDateTime", e)
	if err != nil {
		return nil, err
	}
	month, err := toInt(args[1], "localDateTime", e)
	if err != nil {
		return nil, err
	}
	day, err := toInt(args[2], "localDateTime", e)
	if err != nil {
		return nil, err
	}
	hour, err := toInt(args[3], "localDateTime", e)
	if err != nil {
		return nil, err
	}
	minute, err := toInt(args[4], "localDateTime", e)
	if err != nil {
		return nil, err
	}
	second, err := toInt(args[5], "localDateTime", e)
	if err != nil {
		return nil, err
	}

	t := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return t.Format("2006-01-02T15:04:05"), nil
}

// callBuiltinLocalTime implements the localTime(hour, minutes, seconds) function.
func callBuiltinLocalTime(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 3, "localTime requires exactly 3 arguments: hour, minutes, seconds", e); err != nil {
		return nil, err
	}

	hour, err := toInt(args[0], "localTime", e)
	if err != nil {
		return nil, err
	}
	minute, err := toInt(args[1], "localTime", e)
	if err != nil {
		return nil, err
	}
	second, err := toInt(args[2], "localTime", e)
	if err != nil {
		return nil, err
	}

	t := time.Date(0, 1, 1, hour, minute, second, 0, time.UTC)
	return t.Format("15:04:05"), nil
}

// callBuiltinAtBeginningOfDay implements the atBeginningOfDay(dateTime) function.
func callBuiltinAtBeginningOfDay(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	dateStr, err := requireOneStringArg(args, "atBeginningOfDay", e)
	if err != nil {
		return nil, err
	}
	t, err := parseDateTime(dateStr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("atBeginningOfDay: %s", err), e.Pos())
	}

	result := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return result.Format(time.RFC3339), nil
}

// callBuiltinAtBeginningOfHour implements the atBeginningOfHour(dateTime) function.
func callBuiltinAtBeginningOfHour(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	dateStr, err := requireOneStringArg(args, "atBeginningOfHour", e)
	if err != nil {
		return nil, err
	}
	t, err := parseDateTime(dateStr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("atBeginningOfHour: %s", err), e.Pos())
	}

	result := time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), 0, 0, 0, t.Location())
	return result.Format(time.RFC3339), nil
}

// callBuiltinAtBeginningOfMonth implements the atBeginningOfMonth(dateTime) function.
func callBuiltinAtBeginningOfMonth(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	dateStr, err := requireOneStringArg(args, "atBeginningOfMonth", e)
	if err != nil {
		return nil, err
	}
	t, err := parseDateTime(dateStr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("atBeginningOfMonth: %s", err), e.Pos())
	}

	result := time.Date(t.Year(), t.Month(), FirstDayOfMonth, 0, 0, 0, 0, t.Location())
	return result.Format(time.RFC3339), nil
}

// callBuiltinAtBeginningOfWeek implements the atBeginningOfWeek(dateTime) function.
func callBuiltinAtBeginningOfWeek(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	dateStr, err := requireOneStringArg(args, "atBeginningOfWeek", e)
	if err != nil {
		return nil, err
	}
	t, err := parseDateTime(dateStr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("atBeginningOfWeek: %s", err), e.Pos())
	}

	// Go back to Sunday (weekday 0)
	daysFromSunday := int(t.Weekday())
	result := time.Date(t.Year(), t.Month(), t.Day()-daysFromSunday, 0, 0, 0, 0, t.Location())
	return result.Format(time.RFC3339), nil
}

// callBuiltinAtBeginningOfYear implements the atBeginningOfYear(dateTime) function.
func callBuiltinAtBeginningOfYear(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	dateStr, err := requireOneStringArg(args, "atBeginningOfYear", e)
	if err != nil {
		return nil, err
	}
	t, err := parseDateTime(dateStr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("atBeginningOfYear: %s", err), e.Pos())
	}

	result := time.Date(t.Year(), time.Month(FirstMonthOfYear), FirstDayOfMonth, 0, 0, 0, 0, t.Location())
	return result.Format(time.RFC3339), nil
}

// callBuiltinDayOfWeek implements the dayOfWeek(date) function.
// Returns the day of the week as an integer (1 = Monday, 7 = Sunday) following DataWeave convention.
func callBuiltinDayOfWeek(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "dayOfWeek requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	dateStr, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("dayOfWeek expects a date string, got %T", args[0]), e.Pos())
	}
	t, err := parseDateTime(dateStr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("dayOfWeek: %s", err), e.Pos())
	}

	// Go's Weekday(): Sunday = 0, Monday = 1, ..., Saturday = 6
	// DataWeave convention: Monday = 1, Tuesday = 2, ..., Sunday = 7
	weekday := int(t.Weekday())
	if weekday == SundayWeekday {
		weekday = DaysInWeek
	}
	return float64(weekday), nil
}

// callBuiltinDayOfYear implements the dayOfYear(date) function.
// Returns the day of the year as an integer (1-366).
func callBuiltinDayOfYear(args []interface{}, e *ast.CallExpr) (interface{}, error) {
	if err := requireExactArgs(args, 1, "dayOfYear requires exactly 1 argument", e); err != nil {
		return nil, err
	}
	dateStr, ok := args[0].(string)
	if !ok {
		return nil, newPosError(fmt.Sprintf("dayOfYear expects a date string, got %T", args[0]), e.Pos())
	}
	t, err := parseDateTime(dateStr)
	if err != nil {
		return nil, newPosError(fmt.Sprintf("dayOfYear: %s", err), e.Pos())
	}

	return float64(t.YearDay()), nil
}

// parseDateTime tries to parse a date/time string in various formats.
func parseDateTime(s string) (time.Time, error) {
	layouts := []string{
		time.RFC3339,
		time.RFC3339Nano,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, s); err == nil {
			return t, nil
		}
	}
	return time.Time{}, unifiederrors.EvalErrorf("invalid date format: %s", s)
}

// toInt converts a value to an int for date/time functions.
func toInt(val interface{}, funcName string, e *ast.CallExpr) (int, error) {
	switch v := val.(type) {
	case int:
		return v, nil
	case int64:
		return int(v), nil
	case float64:
		return int(v), nil
	default:
		return 0, newPosError(fmt.Sprintf("%s expects integer, got %T", funcName, val), e.Pos())
	}
}
