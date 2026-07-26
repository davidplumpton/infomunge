Feature: String Functions
  In order to manipulate and transform strings
  As a developer
  I want to use the string functions: splitBy, lower, upper, and joinBy

  Scenario: splitBy basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "a-b-c" splitBy "-"
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","b","c"]
      """

  Scenario: splitBy with multi-character separator
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "hello WORLD hello" splitBy " WORLD "
      """
    When I run the application with this content
    Then the output should be:
      """
      ["hello","hello"]
      """

  Scenario: lower basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      lower("HeLLo")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello"
      """

  Scenario: upper basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      upper("HeLLo")
      """
    When I run the application with this content
    Then the output should be:
      """
      "HELLO"
      """

  Scenario: unary string functions coerce scalar values and propagate null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [lower(123), lower(true), lower(null), upper(-12.5), upper(false), upper(null), trim(12.5), trim(1e10), trim(true), trim(null)]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["123","true",null,"-12.5","FALSE",null,"12.5","1E+10","true",null]
      """

  Scenario: string predicates coerce scalar values and treat null sources as empty
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [startsWith(12345, 123), startsWith(true, true), startsWith(null, null), endsWith(12345, 45), endsWith(false, false), endsWith(null, null), contains(12345, 234), contains(1e10, "E"), contains(true, true), contains(12345, /23/), contains(null, null), contains("12345", 234), contains([1, null], null)]
      """
    When I run the application with this content
    Then the output should be:
      """
      [true,true,false,true,true,false,true,true,true,true,false,true,true]
      """

  Scenario: joinBy basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["a", "b", "c"] joinBy "-"
      """
    When I run the application with this content
    Then the output should be:
      """
      "a-b-c"
      """

  Scenario: joinBy with non-string elements
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, true, "test"] joinBy ","
      """
    When I run the application with this content
    Then the output should be:
      """
      "1,true,test"
      """

  Scenario: joinBy with empty array
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [] joinBy "-"
      """
    When I run the application with this content
    Then the output should be:
      """
      ""
      """

  Scenario: splitBy followed by joinBy associates from the left
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "a,b" splitBy "," joinBy "-"
      """
    When I run the application with this content
    Then the output should be:
      """
      "a-b"
      """

  Scenario: joinBy completes before following concatenation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      ["a", "b"] joinBy "-" ++ "c"
      """
    When I run the application with this content
    Then the output should be:
      """
      "a-bc"
      """

  Scenario: splitBy completes before following array concatenation
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "a,b" splitBy "," ++ ["c"]
      """
    When I run the application with this content
    Then the output should be:
      """
      ["a","b","c"]
      """

  Scenario: toLower and toUpper aliases
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {
        l: toLower("ABC"),
        u: toUpper("abc")
      }
      """
    When I run the application with this content
    Then the output should be:
      """
      {"l":"abc","u":"ABC"}
      """

  Scenario: toBase64 basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      toBase64("Hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      "SGVsbG8="
      """

  Scenario: fromBase64 basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      fromBase64("SGVsbG8=")
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello"
      """

  Scenario: hash with MD5
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      hash("Hello", "MD5")
      """
    When I run the application with this content
    Then the output should be:
      """
      "8b1a9953c4611296a827abf8c47804d7"
      """

  Scenario: hash with SHA-256
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      hash("Hello", "SHA-256")
      """
    When I run the application with this content
    Then the output should be:
      """
      "185f8db32271fe25f561a6fc938b2e264306ec304eda518007d1764826381969"
      """

  Scenario: toHex basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      toHex("Hello")
      """
    When I run the application with this content
    Then the output should be:
      """
      "48656c6c6f"
      """

  Scenario: fromHex basic usage
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      fromHex("48656c6c6f")
      """
    When I run the application with this content
    Then the output should be:
      """
      "Hello"
      """

  Scenario: splitBy requires string first argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      123 splitBy "-"
      """
    When I run the application and it fails
    Then the output should contain "splitBy expects first argument to be a string"

  Scenario: splitBy requires string second argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "abc" splitBy 1
      """
    When I run the application and it fails
    Then the output should contain "splitBy expects second argument to be a string"

  Scenario: joinBy requires array first argument
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      "abc" joinBy "-"
      """
    When I run the application and it fails
    Then the output should contain "joinBy: argument 1 expected array"

   Scenario: joinBy requires string second argument
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       ["a", "b"] joinBy 1
       """
     When I run the application and it fails
     Then the output should contain "joinBy: argument 2 expected string"

   Scenario: capitalize basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       capitalize("hello world")
       """
     When I run the application with this content
     Then the output should be:
       """
       "Hello World"
       """

   Scenario: capitalize single word
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       capitalize("hello")
       """
     When I run the application with this content
     Then the output should be:
       """
       "Hello"
       """

   Scenario: capitalize empty string
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       capitalize("")
       """
     When I run the application with this content
     Then the output should be:
       """
       ""
       """

   Scenario: capitalize splits camelCase into words
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       capitalize("StoreOfOrigin")
       """
     When I run the application with this content
     Then the output should be:
       """
       "Store Of Origin"
       """

   Scenario: capitalize handles accented and emoji-leading words
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       [capitalize("éclair"), capitalize("🙂 smile")]
       """
     When I run the application with this content
     Then the output should be:
       """
       ["Éclair","🙂 Smile"]
       """

   Scenario: capitalize detects non-ASCII camel-case boundaries
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       capitalize("déjàVu")
       """
     When I run the application with this content
     Then the output should be:
       """
       "Déjà Vu"
       """

   Scenario: camelize basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       camelize("hello-world")
       """
     When I run the application with this content
     Then the output should be:
       """
       "helloWorld"
       """

   Scenario: camelize with underscores
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       camelize("hello_world")
       """
     When I run the application with this content
     Then the output should be:
       """
       "helloWorld"
       """

   Scenario: camelize multiple parts
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       camelize("user-name-info")
       """
     When I run the application with this content
     Then the output should be:
       """
       "userNameInfo"
       """

   Scenario: camelize preserves camelCase boundaries
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       camelize("VersionNo")
       """
     When I run the application with this content
     Then the output should be:
       """
       "versionNo"
       """

   Scenario: camelize preserves camelCase in parts
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       camelize("some-VersionNo")
       """
     When I run the application with this content
     Then the output should be:
       """
       "someVersionNo"
       """

   Scenario: camelize handles accented and emoji-leading parts
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       [camelize("Éclair_test"), camelize("🙂_smile")]
       """
     When I run the application with this content
     Then the output should be:
       """
       ["éclairTest","🙂Smile"]
       """

   Scenario: dasherize basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dasherize("helloWorld")
       """
     When I run the application with this content
     Then the output should be:
       """
       "hello-world"
       """

   Scenario: dasherize multiple capitals
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dasherize("userNameInfo")
       """
     When I run the application with this content
     Then the output should be:
       """
       "user-name-info"
       """

   Scenario: dasherize detects non-ASCII camel-case boundaries
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dasherize("déjàVu")
       """
     When I run the application with this content
     Then the output should be:
       """
       "déjà-vu"
       """

   Scenario: underscore basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       underscore("helloWorld")
       """
     When I run the application with this content
     Then the output should be:
       """
       "hello_world"
       """

   Scenario: underscore multiple capitals
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       underscore("userNameInfo")
       """
     When I run the application with this content
     Then the output should be:
       """
       "user_name_info"
       """

   Scenario: underscore detects non-ASCII camel-case boundaries
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       underscore("déjàVu")
       """
     When I run the application with this content
     Then the output should be:
       """
       "déjà_vu"
       """

   Scenario: capitalize requires string
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       capitalize(123)
       """
     When I run the application and it fails
     Then the output should contain "capitalize expects a string"

   Scenario: camelize requires string
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       camelize(123)
       """
     When I run the application and it fails
     Then the output should contain "camelize expects a string"

   Scenario: dasherize requires string
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       dasherize(123)
       """
     When I run the application and it fails
     Then the output should contain "dasherize expects a string"

   Scenario: underscore requires string
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       underscore(123)
       """
     When I run the application and it fails
     Then the output should contain "underscore expects a string"

   Scenario: capitalize requires exactly 1 argument
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       capitalize()
       """
     When I run the application and it fails
     Then the output should contain "capitalize requires exactly 1 argument"

   Scenario: camelize requires exactly 1 argument
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       camelize()
       """
     When I run the application and it fails
     Then the output should contain "camelize requires exactly 1 argument"

   # leftPad tests
   Scenario: leftPad basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad("42", 5, "0")
       """
     When I run the application with this content
     Then the output should be:
       """
       "00042"
       """

   Scenario: leftPad with space padding
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad("test", 8, " ")
       """
     When I run the application with this content
     Then the output should be:
       """
       "    test"
       """

   Scenario: leftPad string already long enough
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad("hello", 3, "x")
       """
     When I run the application with this content
     Then the output should be:
       """
       "hello"
       """

   Scenario: leftPad with multi-character pad text
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad("x", 5, "ab")
       """
     When I run the application with this content
     Then the output should be:
       """
       "ababx"
       """

   Scenario: leftPad with empty string
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad("", 3, "*")
       """
     When I run the application with this content
     Then the output should be:
       """
       "***"
       """

   Scenario: leftPad requires exactly 3 arguments
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad("test", 5)
       """
     When I run the application and it fails
     Then the output should contain "leftPad requires exactly 3 arguments"

   Scenario: leftPad with non-string first argument fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad(123, 5, "0")
       """
     When I run the application and it fails
     Then the output should contain "leftPad expects a string as argument 1"

   Scenario: leftPad with non-integer size fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad("test", "5", "0")
       """
     When I run the application and it fails
     Then the output should contain "leftPad expects an integer as argument 2"

   Scenario: leftPad with empty padText fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       leftPad("test", 10, "")
       """
     When I run the application and it fails
     Then the output should contain "leftPad padText cannot be empty"

   # rightPad tests
   Scenario: rightPad basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad("42", 5, "0")
       """
     When I run the application with this content
     Then the output should be:
       """
       "42000"
       """

   Scenario: rightPad with space padding
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad("test", 8, " ")
       """
     When I run the application with this content
     Then the output should be:
       """
       "test    "
       """

   Scenario: rightPad string already long enough
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad("hello", 3, "x")
       """
     When I run the application with this content
     Then the output should be:
       """
       "hello"
       """

   Scenario: rightPad with multi-character pad text
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad("x", 5, "ab")
       """
     When I run the application with this content
     Then the output should be:
       """
       "xabab"
       """

   Scenario: rightPad with empty string
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad("", 3, "*")
       """
     When I run the application with this content
     Then the output should be:
       """
       "***"
       """

   Scenario: rightPad requires exactly 3 arguments
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad("test", 5)
       """
     When I run the application and it fails
     Then the output should contain "rightPad requires exactly 3 arguments"

   Scenario: rightPad with non-string first argument fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad(123, 5, "0")
       """
     When I run the application and it fails
     Then the output should contain "rightPad expects a string as argument 1"

   Scenario: rightPad with non-integer size fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad("test", "5", "0")
       """
     When I run the application and it fails
     Then the output should contain "rightPad expects an integer as argument 2"

   Scenario: rightPad with empty padText fails
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       rightPad("test", 10, "")
       """
     When I run the application and it fails
     Then the output should contain "rightPad padText cannot be empty"

   Scenario: pluralize basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       {
         a: pluralize("cat"),
         b: pluralize("bus"),
         c: pluralize("party")
       }
       """
     When I run the application with this content
     Then the output should be:
       """
       {"a":"cats","b":"buses","c":"parties"}
       """

   Scenario: singularize basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       {
         a: singularize("cats"),
         b: singularize("buses"),
         c: singularize("parties"),
         d: singularize("knives")
       }
       """
     When I run the application with this content
     Then the output should be:
       """
       {"a":"cat","b":"bus","c":"party","d":"knife"}
       """

   Scenario: ordinalize basic usage
     Given the following input content:
       """
       %im 0.1
       output application/json
       ---
       {
         a: ordinalize(1),
         b: ordinalize(2),
         c: ordinalize(3),
         d: ordinalize(4),
         e: ordinalize(11),
         f: ordinalize(13),
         g: ordinalize(22)
       }
       """
     When I run the application with this content
     Then the output should be:
       """
       {"a":"1st","b":"2nd","c":"3rd","d":"4th","e":"11th","f":"13th","g":"22nd"}
       """
