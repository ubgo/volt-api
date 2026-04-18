# Behaviours

Operating modes that shift AI emphasis. Switch with `/codeskill behaviour <name>`.

| Name | When | Emphasis |
|------|------|----------|
| `default` | Normal work | Balanced — follow rules, ask when unclear |
| `careful` | Auth, payments, migrations | Extra verification, smaller diffs, no speculation |
| `review` | Reviewing PRs or diffs | Spot issues, check rules, no rewriting |
| `debug` | Investigating bugs | Root cause first, no fixes until cause known |
| `scaffold` | Boilerplate/new modules | Match existing templates, no novel design |

Active behaviour is reflected in CLAUDE.md. Run `/codeskill sync` after switching.
