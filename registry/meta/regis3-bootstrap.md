---
regis3:
  type: skill
  name: regis3-bootstrap
  desc: Enables self-installation of skills from regis3 registry
  tags:
    - meta
    - bootstrap
    - core
---

# regis3 Registry Access

You have access to a skill registry via the `regis3` CLI. Use it to find and install relevant skills for the current task.

## Available Commands

```bash
# Search for relevant skills
regis3 search <query>

# List all available items
regis3 list
regis3 list --tag <tag>

# Get details about an item
regis3 info <type:name>

# Install items (with dependencies)
regis3 add <type:name> [<type:name> ...]

# See what's currently installed
regis3 status
```

## When to Use

Before starting a task, consider:

1. What domain is this task in?
2. What technologies are involved?
3. Are there relevant skills in the registry?

Search the registry and install relevant skills BEFORE writing code.
