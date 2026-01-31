package evaluator

import "time"

// Evaluation depth and recursion limits
const (
	// MaxEvalDepth is the maximum recursion depth for expression evaluation
	// to prevent infinite recursion and stack overflow.
	MaxEvalDepth = 100

	// MaxDeepDepth is the maximum recursion depth for deep search operations
	// in nested data structures. Higher than MaxEvalDepth to allow deep traversal.
	MaxDeepDepth = 1000
)

// Loop control constants
const (
	// DefaultLoopTimeout is the default maximum duration for while loops
	// if no deadline is set in the execution context.
	DefaultLoopTimeout = 30 * time.Second

	// TimeoutCheckInterval controls how often (in iterations) we check
	// if a loop has exceeded its timeout. Checking every iteration would
	// be expensive, so we check periodically to balance responsiveness and overhead.
	TimeoutCheckInterval = 100
)

// Date and time constants
const (
	// HoursPerDay is the number of hours in a day for date calculations.
	HoursPerDay = 24

	// LeapYear divisors for determining if a year is a leap year.
	// A year is a leap year if:
	// - divisible by 4 AND not divisible by 100, OR
	// - divisible by 400
	LeapYearDivisor4   = 4
	LeapYearDivisor100 = 100
	LeapYearDivisor400 = 400

	// FirstMonthOfYear is January (1-indexed in Go's time package).
	FirstMonthOfYear = 1

	// FirstDayOfMonth is the first day of any month.
	FirstDayOfMonth = 1

	// SundayWeekday represents Sunday as weekday 0 in Go's time package.
	SundayWeekday = 0

	// DaysInWeek is the number of days in a week (used for dayOfWeek conversion).
	DaysInWeek = 7
)

// Numeric constants
const (
	// ParityDivisor is used to check if a number is even (num % ParityDivisor == 0)
	// or odd (num % ParityDivisor != 0).
	ParityDivisor = 2

	// DecimalBase is base 10 for decimal number formatting.
	DecimalBase = 10

	// FloatBitSize is 64-bit for float parsing operations.
	FloatBitSize = 64

	// ThousandsSeparatorGroupSize is the grouping size for thousands separators
	// in formatted numbers (e.g., 1,000,000).
	ThousandsSeparatorGroupSize = 3
)

// Boolean conversion constants
const (
	// BoolTrueAsInt is the integer representation of true.
	BoolTrueAsInt = 1

	// BoolFalseAsInt is the integer representation of false or nil.
	BoolFalseAsInt = 0
)

// Character constants
const (
	// SpaceCharacter is the space character used for blank string checks.
	SpaceCharacter = ' '
)
