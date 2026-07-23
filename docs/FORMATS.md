# Supported formats and fidelity

InfoMunge's format registry supports the MIME types and file extensions below.
“Structured” means the reader converts content into InfoMunge values and the
writer serializes those values in the named format. “Raw passthrough” means
input is returned as a byte-preserving Go string and output accepts only a
string or `[]byte`; it does not imply that InfoMunge understands the file's
schema or container format.

File extensions are used to detect CLI input formats. MIME aliases can be used
anywhere the corresponding canonical MIME type is accepted. MIME values must
match a registered value exactly.

<!-- format-matrix:start -->
| Format | Canonical MIME type | MIME aliases | Extensions | Input | Output | Fidelity, options, and limitations |
| --- | --- | --- | --- | --- | --- | --- |
| JSON | `application/json` | — | `.json` | Yes | Yes | Structured. Numbers are decoded using Go JSON semantics and output is compact JSON. |
| YAML | `application/yaml` | — | `.yaml`, `.yml` | Yes | Yes | Structured single-document YAML. |
| XML | `application/xml` | — | `.xml` | Yes | Yes | Structured object mapping. Attributes use `@`, text uses `#text`, and repeated elements become arrays. Whitespace-only text is discarded, so mixed-content documents are not lossless. |
| CSV | `application/csv` | `text/csv` | `.csv` | Yes | Yes | Structured. The first input row is the header and all input fields are strings. Output requires an array of objects, uses the sorted union of their keys, and renders nested values as compact JSON text. |
| NDJSON | `application/x-ndjson` | `application/ndjson` | `.ndjson`, `.jsonl` | Yes | Yes | Structured. Blank input lines are skipped; output requires an array and emits one compact JSON value per line. |
| URL-encoded form | `application/x-www-form-urlencoded` | — | `.urlencoded` | Yes | Yes | Structured object mapping. Repeated keys become arrays; decoded values remain strings. Output requires an object. |
| Multipart form | `multipart/form-data` | — | `.multipart`, `.formdata` | Yes | Yes | Structured object mapping with repeated names as arrays and file-like parts represented by `content`, `filename`, and `contentType`. Input discovers the boundary from the first delimiter; output uses a fixed InfoMunge boundary. It does not recursively decode part content. |
| Java properties | `text/x-java-properties` | — | `.properties` | Yes | No | Structured input only. Supports comments, continuations, common separators, and common escapes; it does not implement Java `\uXXXX` Unicode escapes. |
| Plain text | `text/plain` | — | `.txt` | Yes | Yes | Text input. String output is unchanged; other values are rendered as compact JSON when possible. |
| Binary | `application/octet-stream` | — | `.bin`, `.binary` | Yes | Yes | Raw passthrough only. No binary parsing, encoding, or base64 conversion is performed. |
| Avro | `application/avro` | — | `.avro` | Yes | Yes | Raw passthrough only. Avro containers and schemas are not decoded or encoded. |
| DataWeave source | `application/dw` | — | `.dw`, `.dwl` | Yes | Yes | Raw passthrough only. Registering this data format does not execute or validate the source as a script. |
| Flat file | `application/flatfile` | — | `.flatfile`, `.ffd` | Yes | Yes | Raw by default. Options: structured conversion requires `schema.fields`; fields specify `name`, positive byte `length`, and optional `type`, `align`, `pad`, and `trim`; `singleRecord` is optional. Records must exactly match the schema width. |
| Java object | `application/java` | — | `.java`, `.ser` | Yes | Yes | Raw by default. Options: `structured: true` reads or writes InfoMunge's JSON `{"@class": ..., "value": ...}` envelope; `strict` controls known-class validation and output may set `class`. This is not Java Object Serialization and does not parse Java source. |
| Protobuf | `application/protobuf` | `application/x-protobuf` | `.protobuf`, `.pb`, `.pbf` | Yes | Yes | Raw by default. Options: `structured: true` requires either an inline `schema` or serialized descriptor set plus `message`; `strict` defaults to true. Structured mode supports the implemented scalar, enum, repeated/packed, nested-message, and map fields, not arbitrary schema-free decoding. |
| XLSX | `application/xlsx` | `application/vnd.openxmlformats-officedocument.spreadsheetml.sheet` | `.xlsx`, `.excel` | Yes | Yes | Raw by default. Options: `structured: true` maps an object of sheet names to arrays of row objects; `strict` defaults to true. The first row supplies headers. Structured output creates a minimal workbook and does not preserve formulas, styles, charts, macros, or other workbook metadata. |
<!-- format-matrix:end -->

## Structured option examples

Format-specific options are available through the three-argument
`read(content, mimeType, options)` and `write(value, mimeType, options)`
functions, and through the equivalent Go APIs. CLI file inputs and an `output`
header select a format but do not supply these structured codec options.

Flat files accept `fields` directly or nested under `schema`:

```im
%im 0.1
output application/json
---
read("Alice     030", "application/flatfile", {schema: {singleRecord: true, fields: [{name: "name", length: 10}, {name: "age", length: 3, type: "integer", align: "right", pad: "0"}]}})
```

Protobuf structured mode needs an inline field schema or a serialized
`FileDescriptorSet` and message name:

```im
%im 0.1
output text/plain
---
write({name: "Alice", active: true}, "application/protobuf", {structured: true, schema: {fields: [{number: 1, name: "name", type: "string"}, {number: 2, name: "active", type: "bool"}]}})
```

Java and XLSX structured modes are opt-in. For example, Java structured output
produces InfoMunge's JSON envelope:

```im
%im 0.1
output text/plain
---
write({name: "Alice"}, "application/java", {structured: true})
```

XLSX structured output returns workbook bytes; this example reports their size:

```im
%im 0.1
output application/json
---
sizeOf(write({People: [{name: "Alice"}]}, "application/xlsx", {structured: true}))
```

Without those options, Flat file, Java object, Protobuf, and XLSX use raw
passthrough. Avro, DataWeave source, and Binary have no structured mode.
