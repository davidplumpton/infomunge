# Date Formatting

InfoMunge supports formatting strings that parse as dates/times using Java-style
`SimpleDateFormat` patterns. The implementation intentionally supports a small,
explicit subset that maps to Go's `time` layouts.

## Supported Java Tokens

| Java token | Meaning | Go layout |
| --- | --- | --- |
| `yyyy` | 4-digit year | `2006` |
| `yy` | 2-digit year | `06` |
| `MMMM` | Full month name | `January` |
| `MMM` | Short month name | `Jan` |
| `MM` | Month number (zero-padded) | `01` |
| `M` | Month number | `1` |
| `dd` | Day of month (zero-padded) | `02` |
| `d` | Day of month | `2` |
| `EEEE` | Full weekday name | `Monday` |
| `EEE` / `EE` / `E` | Short weekday name | `Mon` |
| `HH` | Hour (00-23) | `15` |
| `H` | Hour (0-23) | `15` |
| `hh` | Hour (01-12) | `03` |
| `h` | Hour (1-12) | `3` |
| `mm` | Minute (zero-padded) | `04` |
| `m` | Minute | `4` |
| `ss` | Second (zero-padded) | `05` |
| `s` | Second | `5` |
| `SSS` | Milliseconds (zero-padded) | `000` |
| `SS` | Centiseconds (zero-padded) | `00` |
| `S` | Deciseconds | `0` |
| `a` | AM/PM marker | `PM` |
| `z` | General time zone | `MST` |
| `Z` | RFC822 time zone | `-0700` |
| `XXX` | ISO 8601 time zone | `Z07:00` |
| `XX` | ISO 8601 time zone | `Z0700` |
| `X` | ISO 8601 time zone | `Z07` |

## Literal Text

Single quotes escape literal text: `yyyy-MM-dd'T'HH:mm:ss` renders a literal `T`.
Use doubled quotes `''` to emit a single quote.

## Limitations

Unsupported Java tokens are left unchanged in the output pattern. This keeps
the behavior predictable but is not a full `SimpleDateFormat` implementation.
If broader DataWeave compatibility is needed, consider implementing a dedicated
Java pattern parser (or a vetted library) in a follow-up task.
