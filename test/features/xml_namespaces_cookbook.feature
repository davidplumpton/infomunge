Feature: XML Namespaces Cookbook
  In order to support complex XML with namespaces
  As a developer
  I want to use DataWeave-style namespace declarations and prefixes

  Scenario: XML Namespaces from Cookbook Example
    Given the following script:
      """
      %im 0.1
      output application/xml
      ns orders http://www.acme.com/shemas/Orders
      ns stores http://www.acme.com/shemas/Stores
      ---
      root: {
        orders#orders: {
          stores#shipNodeId: "SF01",
          stores#shipNodeId @(shipsVia:"LA01"): "NY03"
        }
      }
      """
    When I run the script
    Then the output should be:
      """
      <root xmlns:orders="http://www.acme.com/shemas/Orders" xmlns:stores="http://www.acme.com/shemas/Stores"><orders:orders><stores:shipNodeId>SF01</stores:shipNodeId><stores:shipNodeId shipsVia="LA01">NY03</stores:shipNodeId></orders:orders></root>
      """
