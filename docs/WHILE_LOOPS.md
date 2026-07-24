# While Loops

InfoMunge supports stateful `while` loops in script bodies.

## Syntax

```im
var i = 0
var result = []
---
while (i < 3) {
  result = result + [i]
  i = i + 1
}
result
```

This script returns `[0,1,2]`. Statements in the loop body run in order, and
assignments update variables for later statements and iterations.

The condition must evaluate to a Boolean. It is checked before every iteration,
so a false initial condition skips the body.

## Control Flow

`break` exits the nearest loop. `continue` skips the rest of the current body
and begins the next condition check:

```im
var i = 0
var result = []
---
while (i < 5) {
  i = i + 1
  if (i == 3) continue
  result = result + [i]
}
result
```

This returns `[1,2,4,5]`. Nested loops keep their own `break` and `continue`
handling.

## Result

A loop expression returns the last completed body result. It returns `null`
when the body never completes an iteration or the loop exits with `break`
before producing a body result. Scripts commonly read an updated variable
after the loop instead of using the loop expression directly.

## Timeouts

Loop evaluation uses the caller's execution-context deadline when one is
present. Without a caller deadline, a loop has a 30-second default deadline.
This prevents an unbounded loop from running forever. The Godog harness uses a
shorter per-scenario deadline and lets timeout scenarios override it explicitly.

Behavioral coverage lives in `test/features/while_loops.feature`; preprocessing
and evaluator edge cases also have focused Go tests.
