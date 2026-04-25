package evaluator

import (
	"go/ast"
	"sync"
)

type SpecialBuiltinFunc = func(*ast.CallExpr, Context, int) (Value, error)
type RegularBuiltinFunc = func([]Value, *ast.CallExpr) (Value, error)

// builtinSpecialRegistry maps special builtin function names to their handlers.
// Special functions need unevaluated arguments and are called via GetBuiltinSpecial.
var builtinSpecialRegistry map[string]SpecialBuiltinFunc

// builtinFunctionRegistry maps regular builtin function names to their handlers.
// Regular functions receive evaluated arguments and are called via GetBuiltinFunction.
var builtinFunctionRegistry map[string]RegularBuiltinFunc

// builtinRegistryMu protects both builtin registries for concurrent reads/writes.
var builtinRegistryMu sync.RWMutex

func init() {
	builtinSpecialRegistry = map[string]SpecialBuiltinFunc{
		"__ifelse":          callBuiltinIfElse,
		"__default":         callBuiltinDefault,
		"__lambda":          callBuiltinLambdaAST,
		"lazy_eval":         callBuiltinLazyEval,
		"__while":           callBuiltinWhile,
		"__assign":          callBuiltinAssign,
		"__break":           callBuiltinBreak,
		"__continue":        callBuiltinContinue,
		"__seq":             callBuiltinSeq,
		"__do":              callBuiltinDo,
		"__updateExpr":      callBuiltinUpdateExpr,
		"__filter":          callBuiltinFilter,
		"__filter_selector": callBuiltinFilterSelector,
		"__map":             callBuiltinMap,
		"__reduce":          callBuiltinReduce,
		"__groupBy":         callBuiltinGroupBy,
		"__flatMap":         callBuiltinFlatMap,
		"__modcall":         callBuiltinModCall,
		"__coerce":          callBuiltinCoerce,
		"__case":            callBuiltinCase,
		"must":              callMust,
		"maxBy":             callBuiltinMaxBy,
		"minBy":             callBuiltinMinBy,
		"orderBy":           callBuiltinOrderBy,
		"distinctBy":        callBuiltinDistinctBy,
		"filterObject":      callBuiltinFilterObject,
		"groupBy":           callBuiltinGroupBy,
		"flatMap":           callBuiltinFlatMap,
		"mapObject":         callBuiltinMapObject,
		"pluck":             callBuiltinPluck,
		"__pluck":           callBuiltinPluck,
		"__toStream":        callBuiltinToStream,
		"__lazyMap":         callBuiltinLazyMap,
		"__lazyFilter":      callBuiltinLazyFilter,
		"__lazyReduce":      callBuiltinLazyReduce,
		"takeWhile":         callBuiltinTakeWhile,
		"dropWhile":         callBuiltinDropWhile,
		"some":              callBuiltinSome,
		"every":             callBuiltinEvery,
		"onNull":            callBuiltinOnNull,
		"then":              callBuiltinThen,
		"try":               callBuiltinTry,
		"orElse":            callBuiltinOrElse,
		"orElseTry":         callBuiltinOrElseTry,
		"read":              callBuiltinRead,
		"readUrl":           callBuiltinReadUrl,
		"write":             callBuiltinWrite,
		"eachItem":          callEachItemMatcher,
		"haveItem":          callHaveItemMatcher,
		"anyOf":             callAnyOfMatcher,
		"notBe":             callNotBeMatcher,
	}

	builtinFunctionRegistry = map[string]RegularBuiltinFunc{
		"__concat":                  callBuiltinConcat,
		"__remove":                  callBuiltinRemove,
		"sizeOf":                    callBuiltinSizeOf,
		"typeOf":                    callBuiltinTypeOf,
		"__isType":                  callBuiltinIsType,
		"isEmpty":                   callBuiltinIsEmpty,
		"flatten":                   callBuiltinFlatten,
		"unique":                    callBuiltinUnique,
		"reverse":                   callBuiltinReverse,
		"sort":                      callBuiltinSort,
		"join":                      callBuiltinJoin,
		"keys":                      callBuiltinKeys,
		"values":                    callBuiltinValues,
		"merge":                     callBuiltinMerge,
		"__with_attrs":              callBuiltinWithAttrs,
		"__update":                  callBuiltinUpdate,
		"__deep":                    callBuiltinDeep,
		"__objvalues":               callBuiltinObjectValues,
		"__multival":                callBuiltinMultival,
		"__metadata":                callBuiltinMetadata,
		"trim":                      callBuiltinTrim,
		"length":                    callBuiltinLength,
		"repeat":                    callBuiltinRepeat,
		"split":                     callBuiltinSplit,
		"substring":                 callBuiltinSubstring,
		"charAt":                    callBuiltinCharAt,
		"charCodeAt":                callBuiltinCharCodeAt,
		"fromCharCode":              callBuiltinFromCharCode,
		"indexOf":                   callBuiltinIndexOf,
		"lastIndexOf":               callBuiltinLastIndexOf,
		"appendIfMissing":           callBuiltinAppendIfMissing,
		"prependIfMissing":          callBuiltinPrependIfMissing,
		"toUpper":                   callBuiltinToUpper,
		"toLower":                   callBuiltinToLower,
		"toBase64":                  callBuiltinToBase64,
		"fromBase64":                callBuiltinFromBase64,
		"hash":                      callBuiltinHash,
		"toHex":                     callBuiltinToHex,
		"fromHex":                   callBuiltinFromHex,
		"upper":                     callBuiltinUpper,
		"lower":                     callBuiltinLower,
		"capitalize":                callBuiltinCapitalize,
		"camelize":                  callBuiltinCamelize,
		"dasherize":                 callBuiltinDasherize,
		"underscore":                callBuiltinUnderscore,
		"pluralize":                 callBuiltinPluralize,
		"singularize":               callBuiltinSingularize,
		"ordinalize":                callBuiltinOrdinalize,
		"leftPad":                   callBuiltinLeftPad,
		"rightPad":                  callBuiltinRightPad,
		"startsWith":                callBuiltinStartsWith,
		"endsWith":                  callBuiltinEndsWith,
		"contains":                  callBuiltinContains,
		"replace":                   callBuiltinReplace,
		"regex":                     callBuiltinRegex,
		"ceil":                      callBuiltinCeil,
		"floor":                     callBuiltinFloor,
		"round":                     callBuiltinRound,
		"sqrt":                      callBuiltinSqrt,
		"abs":                       callBuiltinAbs,
		"max":                       callBuiltinMax,
		"min":                       callBuiltinMin,
		"pow":                       callBuiltinPow,
		"sum":                       callBuiltinSum,
		"avg":                       callBuiltinAvg,
		"mod":                       callBuiltinMod,
		"toRadix":                   callBuiltinToRadix,
		"fromRadix":                 callBuiltinFromRadix,
		"toBinary":                  callBuiltinToBinary,
		"fromBinary":                callBuiltinFromBinary,
		"random":                    callBuiltinRandom,
		"randomInt":                 callBuiltinRandomInt,
		"uuid":                      callBuiltinUUID,
		"log":                       callBuiltinLog,
		"logDebug":                  callBuiltinLogDebug,
		"logInfo":                   callBuiltinLogInfo,
		"logWarn":                   callBuiltinLogWarn,
		"logError":                  callBuiltinLogError,
		"logWith":                   callBuiltinLogWith,
		"to":                        callBuiltinTo,
		"zip":                       callBuiltinZip,
		"unzip":                     callBuiltinUnzip,
		"range":                     callBuiltinRange,
		"distinct":                  callBuiltinDistinct,
		"__range":                   callBuiltinRange,
		"with":                      callBuiltinWith,
		"xsiType":                   callBuiltinXsiType,
		"evaluateCompatibilityFlag": callBuiltinEvaluateCompatibilityFlag,
		"now":                       callBuiltinNow,
		"daysBetween":               callBuiltinDaysBetween,
		"isLeapYear":                callBuiltinIsLeapYear,
		"today":                     callBuiltinToday,
		"tomorrow":                  callBuiltinTomorrow,
		"yesterday":                 callBuiltinYesterday,
		"date":                      callBuiltinDate,
		"time":                      callBuiltinTime,
		"dateTime":                  callBuiltinDateTime,
		"localDateTime":             callBuiltinLocalDateTime,
		"localTime":                 callBuiltinLocalTime,
		"atBeginningOfDay":          callBuiltinAtBeginningOfDay,
		"atBeginningOfHour":         callBuiltinAtBeginningOfHour,
		"atBeginningOfMonth":        callBuiltinAtBeginningOfMonth,
		"atBeginningOfWeek":         callBuiltinAtBeginningOfWeek,
		"atBeginningOfYear":         callBuiltinAtBeginningOfYear,
		"dayOfWeek":                 callBuiltinDayOfWeek,
		"dayOfYear":                 callBuiltinDayOfYear,
		"find":                      callBuiltinFind,
		"match":                     callBuiltinMatch,
		"matches":                   callBuiltinMatches,
		"scan":                      callBuiltinScan,
		"parseURI":                  callBuiltinParseURI,
		"compose":                   callBuiltinCompose,
		"encodeURI":                 callBuiltinEncodeURI,
		"decodeURI":                 callBuiltinDecodeURI,
		"encodeURIComponent":        callBuiltinEncodeURIComponent,
		"decodeURIComponent":        callBuiltinDecodeURIComponent,
		"isBlank":                   callBuiltinIsBlank,
		"isDecimal":                 callBuiltinIsDecimal,
		"isInteger":                 callBuiltinIsInteger,
		"isEven":                    callBuiltinIsEven,
		"isOdd":                     callBuiltinIsOdd,
		"objectToArray":             callBuiltinObjectToArray,
		"arrayToObject":             callBuiltinArrayToObject,
		"joinBy":                    callBuiltinJoinBy,
		"splitBy":                   callBuiltinSplitBy,
		"entriesOf":                 callBuiltinEntriesOf,
		"keysOf":                    callBuiltinKeysOf,
		"valuesOf":                  callBuiltinValuesOf,
		"namesOf":                   callBuiltinNamesOf,
		"beArray":                   callBeArray,
		"beObject":                  callBeObject,
		"beString":                  callBeString,
		"beNumber":                  callBeNumber,
		"beBoolean":                 callBeBoolean,
		"beNull":                    callBeNull,
		"beEmpty":                   callBeEmpty,
		"beBlank":                   callBeBlank,
		"equalTo":                   callEqualTo,
		"beGreaterThan":             callBeGreaterThan,
		"beLowerThan":               callBeLowerThan,
		"beOneOf":                   callBeOneOf,
		"containStr":                callContainStr,
		"containVal":                callContainVal,
		"startWith":                 callStartWith,
		"endWith":                   callEndWith,
		"haveSize":                  callHaveSize,
		"haveKey":                   callHaveKey,
		"haveValue":                 callHaveValue,
		"notBeNull":                 callNotBeNull,
		"fail":                      callBuiltinFail,
		"assert":                    callBuiltinAssert,
		"assertThat":                callBuiltinAssertThat,
		"force_eval":                callBuiltinForceEval,
		"envVar":                    callBuiltinEnvVar,
		"envVars":                   callBuiltinEnvVars,
		"slice":                     callBuiltinSlice,
		"prepend":                   callBuiltinPrepend,
		"append":                    callBuiltinAppend,
		"safe":                      callBuiltinSafeAccess,
		"first":                     callBuiltinFirst,
		"last":                      callBuiltinLast,
		"take":                      callBuiltinTake,
		"drop":                      callBuiltinDrop,
	}
}

func cloneSpecialRegistry(src map[string]SpecialBuiltinFunc) map[string]SpecialBuiltinFunc {
	dst := make(map[string]SpecialBuiltinFunc, len(src))
	for name, fn := range src {
		dst[name] = fn
	}
	return dst
}

func cloneFunctionRegistry(src map[string]RegularBuiltinFunc) map[string]RegularBuiltinFunc {
	dst := make(map[string]RegularBuiltinFunc, len(src))
	for name, fn := range src {
		dst[name] = fn
	}
	return dst
}

// RegisterBuiltinSpecial registers or overrides a special builtin at runtime.
func RegisterBuiltinSpecial(name string, fn SpecialBuiltinFunc) {
	builtinRegistryMu.Lock()
	defer builtinRegistryMu.Unlock()
	builtinSpecialRegistry[name] = fn
}

// RegisterBuiltinFunction registers or overrides a regular builtin at runtime.
func RegisterBuiltinFunction(name string, fn RegularBuiltinFunc) {
	builtinRegistryMu.Lock()
	defer builtinRegistryMu.Unlock()
	builtinFunctionRegistry[name] = fn
}

// SetBuiltinRegistriesForTesting swaps builtin registries and returns a restore func.
// The provided maps are cloned to avoid external mutation races.
func SetBuiltinRegistriesForTesting(
	special map[string]SpecialBuiltinFunc,
	regular map[string]RegularBuiltinFunc,
) func() {
	builtinRegistryMu.Lock()
	prevSpecial := cloneSpecialRegistry(builtinSpecialRegistry)
	prevRegular := cloneFunctionRegistry(builtinFunctionRegistry)
	builtinSpecialRegistry = cloneSpecialRegistry(special)
	builtinFunctionRegistry = cloneFunctionRegistry(regular)
	builtinRegistryMu.Unlock()

	return func() {
		builtinRegistryMu.Lock()
		builtinSpecialRegistry = prevSpecial
		builtinFunctionRegistry = prevRegular
		builtinRegistryMu.Unlock()
	}
}

// GetBuiltinSpecial returns a special function handler by name
// This maintains compatibility with existing visitor code
func GetBuiltinSpecial(name string) (SpecialBuiltinFunc, bool) {
	builtinRegistryMu.RLock()
	defer builtinRegistryMu.RUnlock()
	fn, ok := builtinSpecialRegistry[name]
	return fn, ok
}

// GetBuiltinFunction returns a regular function handler by name
func GetBuiltinFunction(name string) (RegularBuiltinFunc, bool) {
	builtinRegistryMu.RLock()
	defer builtinRegistryMu.RUnlock()
	fn, ok := builtinFunctionRegistry[name]
	return fn, ok
}
