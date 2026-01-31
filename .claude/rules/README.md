# Context-Specific Rules

Rules in this directory are read **on-demand** by LLMs only when performing specific actions.

## Purpose

- **Token optimization:** LLM only loads rules when needed
- **Focused context:** Each rule file contains only relevant instructions
- **Maintainability:** Rules are modular and easy to update

## Usage

LLMs should read the appropriate rule file before performing the action:

| File | Read Before |
|------|-------------|
| `git-commit.md` | Running `git commit` |

## Adding New Rules

1. Create `[action-name].md` in this directory
2. Add entry to table in `AGENTS.md` under "Context-Specific Rules"
3. Keep rules focused and concise

## Format

Each rule file should have:
- Clear title
- "Read this before..." instruction
- Required/Forbidden sections
- Examples where helpful
