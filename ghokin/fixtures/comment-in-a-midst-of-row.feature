Feature: Test Formatter
  Scenario: Comment between rows
    Given table my_table
      | col a | col b |
      | 1     | 2     |
      # A comment about the row below
      | 3     | 4     |
      # Another comment about the row below
      | 5     | 6     |
      | 7     | 8     |
      | 9     | 10    |
    Given table my_table
      | col a | col b |
      | 1     | 2     |
      # A comment about the row below
      | 3     | 4     |
      # Another comment about the row below
      | 5     | 6     |
    Given a second thing
    Given another second thing
    Then another thing happened
