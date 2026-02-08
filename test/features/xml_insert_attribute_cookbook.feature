Feature: XML Insert Attribute Cookbook
  Scenario: Insert an Attribute into an XML Tag
    Given the following XML input:
      """
      <bookstore>
        <book>
          <title>Everyday Italian</title>
          <year>2005</year>
          <price>30</price>
          <author>Giada De Laurentiis</author>
        </book>
        <book>
          <title>Harry Potter</title>
          <year>2005</year>
          <price>29.99</price>
          <author>J K. Rowling</author>
        </book>
        <book>
          <title>XQuery Kick Start</title>
          <year>2003</year>
          <price>49.99</price>
          <author>James McGovern</author>
          <author>Per Bothner</author>
          <author>Kurt Cagle</author>
          <author>James Linn</author>
          <author>Vaidyanathan Nagarajan</author>
        </book>
        <book>
          <title>Learning XML</title>
          <year>2003</year>
          <price>39.95</price>
          <author>Erik T. Ray</author>
        </book>
      </bookstore>
      """
    And the following script:
      """
      %im 0.1
      output application/xml
      ---
      bookstore: {
        (payload.bookstore.*book map (book) -> {
          book : {
            title @(lang: "en", year: book.year): book.title,
            price: book.price,
            (book.*author map (author) -> { author @(loc: "US"): author })
          }
        })
      }
      """
    When I run the script
    Then the output should contain:
      """
      <title lang="en" year="2005">Everyday Italian</title>
      """
    And the output should contain:
      """
      <author loc="US">Giada De Laurentiis</author>
      """
    And the output should contain:
      """
      <title lang="en" year="2003">XQuery Kick Start</title>
      """
    And the output should contain:
      """
      <author loc="US">James McGovern</author>
      """
    And the output should contain:
      """
      <author loc="US">Per Bothner</author>
      """

  Scenario: Attribute insertion keeps full object value before next key
    Given the following script:
      """
      %im 0.1
      output application/xml
      ---
      root: {
        child @(lang: "en", year: (1 + 1)): {
          first: "A",
          second: "B"
        },
        tail: "done"
      }
      """
    When I run the script
    Then the output should contain:
      """
      <child lang="en" year="2"><first>A</first><second>B</second></child>
      """
    And the output should contain:
      """
      <tail>done</tail>
      """
