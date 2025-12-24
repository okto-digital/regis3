---
regis3:
  type: skill
  name: testing
  desc: Testing best practices and patterns
  deps:
    - skill:git-conventions
  tags:
    - testing
    - quality
    - tdd
  status: stable
---

# Testing Best Practices

## Test Structure

Use the AAA pattern:
- **Arrange**: Set up test data
- **Act**: Execute the code under test
- **Assert**: Verify the results

## Naming Conventions

Test names should describe:
1. What is being tested
2. Under what conditions
3. Expected outcome

Example: `TestUserLogin_WithValidCredentials_ReturnsToken`
