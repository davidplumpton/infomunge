Feature: Null stringification
  In order to keep language results independent of Go implementation details
  As a transformation author
  I want implicit null stringification to use the language spelling "null"

  Scenario: joinBy stringifies null alongside scalar values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [1, true, "test", null] joinBy ","
      """
    When I run the application with this content
    Then the output should be:
      """
      "1,true,test,null"
      """

  Scenario: join stringifies null alongside scalar values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      join([1, true, "test", null], ",")
      """
    When I run the application with this content
    Then the output should be:
      """
      "1,true,test,null"
      """

  Scenario: joinBy stringifies null recursively inside structured values
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [null, [null], {value: null}] joinBy "|"
      """
    When I run the application with this content
    Then the output should be:
      """
      "null|[null]|map[value:null]"
      """

  Scenario: Dynamic object keys stringify null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      {(null): 1}
      """
    When I run the application with this content
    Then the output should be:
      """
      {"null":1}
      """

  Scenario: groupBy keys stringify null
    Given the following input content:
      """
      %im 0.1
      output application/json
      ---
      [null] groupBy $
      """
    When I run the application with this content
    Then the output should be:
      """
      {"null":[null]}
      """

  Scenario: URI query values stringify null
    Given the following input content:
      """
      %im 0.1
      output application/json
      import * from dw::core::URL
      ---
      compose({scheme: "https", host: "example.com", query: {empty: null, items: [null]}})
      """
    When I run the application with this content
    Then the output should be:
      """
      "https://example.com?empty=null\u0026items=null"
      """
