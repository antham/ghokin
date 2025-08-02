Feature: A Feature
  Description

  Scenario Outline: A scenario to test
    Given a thing
      """
      [{ "type": <value> }]
      """

    Examples:
      | value       |
      | "\"hello\"" |
      | null        |
