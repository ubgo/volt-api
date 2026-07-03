# Configuration Rules

**Read before:** Adding or modifying application configuration.

## Checklist

- [ ] Edit PKL schema in `config/pkl/schema/`
- [ ] Update `Root.pkl` if adding new section
- [ ] Run `task config:gen` to regenerate Go code
- [ ] Update environment files in `config/pkl/env/*/`

## File Locations

| Type | Location |
|------|----------|
| Schema definitions | `config/pkl/schema/*.pkl` |
| Root config | `config/pkl/schema/Root.pkl` |
| Environment values | `config/pkl/env/<env>/app.pkl` |
| Generated Go code | `config/gen/` (don't edit) |

## Do

- Define types in PKL schema first
- Use appropriate PKL types (`String`, `Int`, `Boolean`, `Duration`)
- Set sensible defaults
- Add new sections to `Root.pkl`
- Update all environment files

## Don't

- Edit files in `config/gen/`
- Hardcode values in Go code
- Commit secrets to environment files

## Adding a Field to Existing Section

```pkl
// config/pkl/schema/App.pkl
module App

class Setting {
    existingField: String
    newField: String = "default"     // Add with default
    optionalField: String?           // Nullable field
}
```

Then run: `task config:gen`

## Adding a New Config Section

1. Create schema file:

```pkl
// config/pkl/schema/MyFeature.pkl
@go.Package { name = "config/gen/myfeatureconfig" }
module MyFeature

import "package://pkg.pkl-lang.org/pkl-go/pkl.golang@0.12.1#/go.pkl"

class Setting {
    enabled: Boolean = false
    apiKey: String
    timeout: Duration = 30.s
    maxRetries: Int = 3
}
```

2. Add to Root.pkl:

```pkl
// config/pkl/schema/Root.pkl
import "MyFeature.pkl"

// ... existing fields
myFeature: MyFeature.Setting
```

3. Update environment files:

```pkl
// config/pkl/env/default/app.pkl
myFeature {
    enabled = false
    apiKey = ""
    timeout = 30.s
    maxRetries = 3
}
```

4. Regenerate: `task config:gen`

## Accessing Config in Go

```go
// Via volt.Kernel
func doSomething(plugin *volt.Kernel) {
    cfg := plugin.RootConfig

    if cfg.MyFeature.Enabled {
        apiKey := cfg.MyFeature.ApiKey
        timeout := cfg.MyFeature.Timeout
        // ...
    }
}

// Direct access (less preferred)
import "config"

func init() {
    cfg := config.Get()
    fmt.Println(cfg.Core.Environment)
}
```

## PKL Types Reference

| PKL Type | Go Type | Example |
|----------|---------|---------|
| `String` | `string` | `"value"` |
| `Int` | `int` | `42` |
| `Boolean` | `bool` | `true` |
| `Float` | `float64` | `3.14` |
| `Duration` | `time.Duration` | `30.s`, `5.m`, `1.h` |
| `String?` | `*string` | nullable |
| `List<String>` | `[]string` | `List("a", "b")` |

## Environment Variables

PKL can read environment variables:

```pkl
// In app.pkl
myFeature {
    apiKey = read("env:MY_API_KEY")
}
```

## Running

```bash
# Regenerate config bindings
task config:gen

# Switch environment
CONFIG_ENV=prod go run ./webapp/cmd/api
```
