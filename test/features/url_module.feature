Feature: URL Module
  In order to work with URLs from a module
  As a developer
  I want the dw::core::URL module to expose URI parse/compose and encode/decode helpers

  Scenario: parseURI returns URL parts
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      parseURI("https://user:pass@example.com:8443/a/b?x=1&x=2&msg=hello%20world#top")
      """
    When I run the application with this content
    Then the output should be:
      """
      {"scheme":"https","host":"example.com","path":"/a/b","fragment":"top","query":{"msg":"hello world","x":["1","2"]},"user":"user","password":"pass","port":8443}
      """

  Scenario: compose builds a URI from parts
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      compose({scheme: "https", user: "user", password: "pass", host: "example.com", port: 8443, path: "/a/b", query: {x: ["1", "2"], msg: "hello world"}, fragment: "top"})
      """
    When I run the application with this content
    Then the output should be:
      """
      "https://user:pass@example.com:8443/a/b?msg=hello+world\u0026x=1\u0026x=2#top"
      """

  Scenario: encodeURI and decodeURI roundtrip a full URI
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      decodeURI(encodeURI("https://example.com/a path?q=hello world#frag ment"))
      """
    When I run the application with this content
    Then the output should be:
      """
      "https://example.com/a path?q=hello world#frag ment"
      """

  Scenario: encodeURIComponent and decodeURIComponent roundtrip components
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      decodeURIComponent(encodeURIComponent("a/b c?d=e"))
      """
    When I run the application with this content
    Then the output should be:
      """
      "a/b c?d=e"
      """

  Scenario: Import Url module and call with namespace
    Given the following input content:
      """
      %im 0.1
      output application/json
      import dw::core::URL
      ---
      URL::encodeURIComponent("hello world")
      """
    When I run the application with this content
    Then the output should be:
      """
      "hello%20world"
      """
