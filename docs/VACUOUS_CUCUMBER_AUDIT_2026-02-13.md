# Cucumber Vacuous-Test Audit (2026-02-13)

Issue: `infomunge-l7sp`

## Summary

The suite has strong coverage in many areas, but several files rely on assertions that can pass while key behavior is broken. The highest-risk pattern is broad substring matching (`Then the output should contain ...`) for structured outputs and errors.

## Highest-Impact Findings

1. `test/features/lambda_expressions.feature` over-focuses on lambda serialization shape.
   - Evidence: many scenarios assert only `Then the output should contain a lambda function` (for example lines 15, 38, 49, 60, 71, 82, 93, 104, 115, 126, 137, 148, 159, 170, 181, 192, 203, 214, 225, 236, 247, 258, 269, 280, 291, 303).
   - Risk: expression bodies could be parsed incorrectly while preserving generic lambda formatting.
   - Recommendation: execute each lambda against sample inputs and assert exact outputs.

2. `test/features/logging_functions.feature` conflates return-value checks with log text checks.
   - Evidence: `Then the output should contain "[DEBUG]"` at lines 72 and 84; similar prefix checks for other levels.
   - Risk: tests can pass if logs are present even when returned expression value is wrong or formatting regresses.
   - Recommendation: separate stdout/stderr concerns and assert exact transformed result independently from log side effects.

3. `test/features/date_time_functions.feature` has duplicate and weak `now()` checks.
   - Evidence: `now returns current time in ISO 8601 format` (line 7) and `now output contains UTC timezone` (line 43) both rely on substring checks (`"T"` and `"Z"` at lines 16 and 52).
   - Risk: malformed timestamps can still satisfy substring presence.
   - Recommendation: parse output as RFC3339 and assert parse success plus UTC offset.

4. `test/features/io_functions.feature` validates parsed structures mostly via substring presence.
   - Evidence: `read basic JSON` line 7 and `read CSV content` line 19 rely on `"Alice"`/`"30"` substring checks (for example line 16, 28); nested JSON test at line 261 also uses contains checks.
   - Risk: structurally incorrect JSON (wrong nesting/types/order side effects) can still pass.
   - Recommendation: replace with exact JSON docstring assertions or typed JSON assertions (specific fields and types).

5. `test/features/dataweave_cookbook_examples.feature` uses cookbook-style contains assertions for complex transformations.
   - Evidence: scenarios at lines 26, 60, and 162 assert a few substrings instead of full transformed structure.
   - Risk: partial output can pass while dropped/incorrect fields go unnoticed.
   - Recommendation: assert complete normalized JSON objects/arrays (or precise key-set + value checks).

6. `test/features/assertion_matchers.feature` negative matcher tests share one coarse error assertion.
   - Evidence: repeated `Then the output should contain "assertion failed"` throughout file (e.g., lines 30, 56, 82, 122, 162, 188, 242, ...).
   - Risk: matcher-specific regressions collapse into generic failures without detection.
   - Recommendation: assert matcher-specific failure text (matcher name and mismatch details), not only generic prefix.

## Follow-Up Work

See linked beads issues created from this audit:
- `infomunge-b4tw` (lambda feature hardening)
- `infomunge-ggng` (logging result vs log-side-effect assertions)
- `infomunge-u6sg` (`now()` RFC3339/UTC assertion hardening)
- `infomunge-11f9` (`io_functions` and cookbook structured assertions)
- `infomunge-54b4` (`assertion_matchers` negative error specificity)
