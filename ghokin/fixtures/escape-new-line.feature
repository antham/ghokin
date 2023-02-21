Feature: Test

  Scenario: Newline in table value
    Given I have a table:
      | a | b \n c | \n d | \n e \n |
      | f | g      | h    | i       |
