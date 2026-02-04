package runner

import "strings"

var standardModuleSources = map[string]string{
	"dw::core::Arrays":  standardModuleArrays,
	"dw::core::Numbers": standardModuleNumbers,
	"dw::core::Strings": standardModuleStrings,
}

const standardModuleArrays = `%dw 2.0

fun countBy(array, criteria) =
  sizeOf(array filter (item) -> criteria(item))

fun divideBy(array, size) =
  if (size <= 0) []
  else if (sizeOf(array) == 0) []
  else [take(array, size)] ++ divideBy(drop(array, size), size)

fun drop(array, amount) = drop(array, amount)
fun dropWhile(array, criteria) = dropWhile(array, criteria)
fun every(array, criteria) = every(array, criteria)

fun firstWith(array, criteria) =
  first(array filter (item) -> criteria(item))

fun indexOf(array, value) =
  using (matches = find(array, value))
    if (sizeOf(matches) == 0) -1 else matches[0]

fun indexWhere(array, criteria) =
  array reduce (acc = -1, item, idx) ->
    if (acc != -1) acc else if (criteria(item)) idx else acc

fun join(left, right, leftKey, rightKey) =
  flatten(
    left map (l) ->
      ((right filter (r) -> leftKey(l) == rightKey(r))
        map (r) -> { l: l, r: r })
  )

fun leftJoin(left, right, leftKey, rightKey) =
  flatten(
    left map (l) ->
      if (sizeOf(right filter (r) -> leftKey(l) == rightKey(r)) == 0)
        [{ l: l, r: null }]
      else
        ((right filter (r) -> leftKey(l) == rightKey(r)) map (r) -> { l: l, r: r })
  )

fun outerJoin(left, right, leftKey, rightKey) =
  leftJoin(left, right, leftKey, rightKey)
    ++ ((right filter (r) -> sizeOf(left filter (l) -> leftKey(l) == rightKey(r)) == 0)
      map (r) -> { l: null, r: r })

fun partition(array, criteria) =
  {
    success: array filter (item) -> criteria(item),
    failure: array filter (item) -> not criteria(item)
  }

fun slice(array, start, end) = slice(array, start, end)
fun some(array, criteria) = some(array, criteria)

fun splitAt(array, index) =
  {
    l: take(array, index),
    r: drop(array, index)
  }

fun splitWhere(array, criteria) =
  {
    l: takeWhile(array, (item) -> not criteria(item)),
    r: dropWhile(array, (item) -> not criteria(item))
  }

fun sumBy(array, mapper) =
  sum(array map (item) -> mapper(item))

fun take(array, amount) = take(array, amount)
fun takeWhile(array, criteria) = takeWhile(array, criteria)
`

const standardModuleStrings = `%dw 2.0

fun appendIfMissing(text, suffix) = appendIfMissing(text, suffix)
fun prependIfMissing(text, prefix) = prependIfMissing(text, prefix)
fun charCodeAt(text, index) = charCodeAt(text, index)
fun fromCharCode(code) = fromCharCode(code)
`

const standardModuleNumbers = `%dw 2.0

fun toRadix(number, radix) = toRadix(number, radix)
fun fromRadix(text, radix) = fromRadix(text, radix)
fun toBinary(number) = toRadix(number, 2)
fun fromBinary(text) = fromRadix(text, 2)
`

func isStandardModule(moduleSpec string) bool {
	_, ok := standardModuleSources[moduleSpec]
	return ok
}

func moduleNameFromSpec(moduleSpec string) string {
	parts := strings.Split(moduleSpec, "::")
	return parts[len(parts)-1]
}
