---
regis3:
  type: skill
  name: git-conventions
  desc: Git workflow and commit message conventions
  tags:
    - git
    - conventions
    - workflow
  status: stable
---

# Git Conventions

## Commit Messages

Use conventional commits format:

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Formatting
- `refactor`: Code restructuring
- `test`: Tests
- `chore`: Maintenance

## Branch Naming

- `feature/<name>` - New features
- `fix/<name>` - Bug fixes
- `release/<version>` - Release branches
