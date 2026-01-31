# Config Module

Type-safe configuration management using [Pkl](https://pkl-lang.org/) - Apple's configuration language with Go code generation.

## Overview

This module provides:

* **Type-safe configuration**: Pkl schemas generate Go structs with full type safety
* **Environment-based configs**: Separate configuration files for dev, staging, and production
* **Modular schemas**: Configuration split into logical modules (Core, Server, Database, etc.)
* **Validation**: Pkl validates configs at evaluation time before your app runs
* **Code generation**: Auto-generated Go code from Pkl schemas

## Directory Structure

```
config/
├── pkl/
│   ├── schema/
│   │   ├── Root.pkl          # Main root config (imports all modules)
│   │   ├── Core.pkl          # Core settings (appName, environment, paths)
│   │   ├── Server.pkl        # Server settings (hostname, port)
│   │   ├── Database.pkl      # Database connection settings
│   │   ├── Nats.pkl          # NATS messaging settings
│   │   ├── Graphql.pkl       # GraphQL settings
│   │   ├── Logger.pkl        # Logger and OpenTelemetry settings
│   │   └── App.pkl           # Project-specific (GitHub, WildAI)
│   └── env/
│       ├── default/          # Local development config
│       │   └── app.pkl
│       ├── prod/             # Production config
│       │   └── app.pkl
│       └── sample/           # Sample config (template)
│           └── app.pkl
├── gen/                      # Generated Go code (DO NOT EDIT)
│   ├── rootconfig/           # RootConfig struct
│   ├── coreconfig/           # Core settings
│   ├── serverconfig/         # Server settings
│   ├── databaseconfig/       # Database settings
│   ├── natsconfig/           # NATS settings
│   ├── graphqlconfig/        # GraphQL settings
│   ├── loggerconfig/         # Logger settings
│   └── appconfig/            # App-specific settings
├── main.go                   # Config loader functions
├── Taskfile.yml              # Task commands
└── go.mod
```

## Prerequisites

### Required Dependencies

| Dependency | Version | Description |
|------------|---------|-------------|
| [Pkl CLI](https://pkl-lang.org/main/current/pkl-cli/index.html) | >= 0.27.0 | Pkl language runtime |
| [pkl-gen-go](https://pkl-lang.org/go/current/quickstart.html) | >= 0.12.0 | Go code generator for Pkl |
| [Go](https://golang.org/) | >= 1.21 | Go programming language |
| [Task](https://taskfile.dev/) | >= 3.0 | Task runner (optional but recommended) |

### Installation

#### macOS (Homebrew)

```bash
# Install Pkl CLI
brew install pkl

# Install pkl-gen-go
brew install pkl-gen-go

# Install Task (optional)
brew install go-task
```

#### Linux

```bash
# Install Pkl CLI
curl -L -o pkl https://github.com/apple/pkl/releases/download/0.27.2/pkl-linux-amd64
chmod +x pkl
sudo mv pkl /usr/local/bin/

# Install pkl-gen-go
curl -L -o pkl-gen-go https://github.com/apple/pkl-go/releases/download/v0.12.1/pkl-gen-go-linux-amd64
chmod +x pkl-gen-go
sudo mv pkl-gen-go /usr/local/bin/

# Install Task (optional)
sh -c "$(curl --location https://taskfile.dev/install.sh)" -- -d -b /usr/local/bin
```

#### Windows

```powershell
# Using Scoop
scoop install pkl
scoop install pkl-gen-go

# Or download manually from:
# - https://github.com/apple/pkl/releases
# - https://github.com/apple/pkl-go/releases
```

### Verify Installation

Run this command to verify all dependencies are installed:

```bash
echo "Checking dependencies..." && \
pkl --version && \
pkl-gen-go --version && \
go version && \
echo "All dependencies installed!"
```

Expected output:

```
Checking dependencies...
Pkl 0.27.x (Linux amd64)
pkl-gen-go 0.12.x
go version go1.25.x darwin/arm64
All dependencies installed!
```

Or use this one-liner to check and report missing dependencies:

```bash
(command -v pkl >/dev/null && echo "pkl: $(pkl --version)" || echo "pkl: NOT INSTALLED") && \
(command -v pkl-gen-go >/dev/null && echo "pkl-gen-go: $(pkl-gen-go --version)" || echo "pkl-gen-go: NOT INSTALLED") && \
(command -v go >/dev/null && echo "go: $(go version)" || echo "go: NOT INSTALLED") && \
(command -v task >/dev/null && echo "task: $(task --version)" || echo "task: NOT INSTALLED (optional)")
```

## Usage

### Loading Configuration in Go

```go
package main

import (
    "config"
    "fmt"
)

func main() {
    // Loads config based on CONFIG_ENV environment variable
    // Defaults to "default" if not set
    cfg := config.Get()

    // Access nested config sections
    fmt.Printf("App: %s, Port: %s\n", cfg.Core.AppName, cfg.Server.Port)
    fmt.Printf("Database: %s@%s:%d\n", cfg.Database.Username, cfg.Database.Host, cfg.Database.Port)
    fmt.Printf("Environment: %s\n", cfg.Core.Environment)
}
```

### Environment Selection

Set `CONFIG_ENV` to load different configurations:

```bash
# Development (default)
CONFIG_ENV=default go run .

# Production
CONFIG_ENV=prod go run .

# Staging
CONFIG_ENV=stag go run .
```

## Schema Modules

The configuration is split into modular schemas in `pkl/schema/`:

### Root.pkl (Main Entry Point)

Imports and composes all other modules:

```pkl
server: Server.Setting
core: Core.Setting
database: Database.Setting
nats: Nats.Setting
graphql: Graphql.Setting
otel: Logger.Otel
logger: Logger.Setting

// Project-specific
app: App.Setting
github: App.Github
wildai: App.WildAI
```

### Core Settings (Core.pkl)

| Field | Type | Description |
|-------|------|-------------|
| `appName` | String | Application name identifier |
| `appKey` | String | Unique identifier for the microservice |
| `environment` | Environment | `"dev"` \| `"stag"` \| `"prod"` |
| `rootdir` | String | Root directory path |
| `tmpdir` | String | Temporary files directory |
| `uploadsdir` | String | Uploads directory |
| `internalKey` | String | Service-to-service communication key |
| `jwtsecret` | String | JWT token signing secret |
| `encryptionKey` | String | Data encryption key |

### Server Settings (Server.pkl)

```pkl
server {
  hostname = "localhost"
  port = "2315"
}
```

### Database Settings (Database.pkl)

```pkl
database {
  host = "localhost"
  port = 5432
  username = "postgres"
  password = "secret"
  dbName = "mydb"
  sslmode = "disable"  // or "require", "verify-full"
}
```

### NATS Settings (Nats.pkl)

```pkl
nats {
  status = true
  appKey = "my-app"
  url = "nats://localhost:4222"
  credsFile = "/path/to/creds.creds"  // Optional
}
```

### GraphQL Settings (Graphql.pkl)

```pkl
graphql {
  key = "playground-access-key"
}
```

### Logger & OpenTelemetry (Logger.pkl)

```pkl
otel {
  enable = true
  endpoint = "otel-collector:4317"
  apiKey = "your-api-key"
  serviceName = "my-service"
  serviceVersion = "1.0.0"
  enableMetrics = true
  enableLogs = true
  enableTraces = true
  enableHttp = true
  enableStdout = true
}

logger {
  enableStdout = true
  enableOtel = true
  enableFile = true
  localConfig {
    enabled = true
    logFile = "/tmp/logs/app.log"
  }
}
```

### App-Specific Settings (App.pkl)

#### GitHub App

```pkl
github {
  appId = "123456"
  appName = "my-github-app"
  clientId = "Iv23li..."
  clientSecret = "secret"
  privateKeyPath = core.rootdir + "/_keys/private-key.pem"
  webhookSecret = "webhook-secret"
  callbackUrl = "https://myapp.com/github/callback"
}
```

#### WildAI Integration

```pkl
wildai {
  partnerID = "prt_xxx"
  clientId = "app_xxx"
  clientSecret = "secret"
  partnerKey = "prt_key"
  mode = "local"  // "local", "stag", or "prod"
}
```

#### App Settings

```pkl
app {
  reactAppHost = "http://localhost:2312"
}
```

## Creating a New Environment

1. Create a new directory under `pkl/env/`:

   ```bash
   mkdir -p pkl/env/myenv
   ```

2. Create `app.pkl` that amends the root schema:

   ```pkl
   amends "../../schema/Root.pkl"

   core {
     appName = "myapp"
     appKey = "myapp-myenv"
     environment = "dev"
     rootdir = "/app"
     tmpdir = rootdir + "/tmp"
     uploadsdir = rootdir + "/uploads"
     internalKey = "secret"
     jwtsecret = "jwt-secret"
     encryptionKey = "encryption-key"
   }

   server {
     hostname = "0.0.0.0"
     port = "8080"
   }

   database {
     host = "db.example.com"
     port = 5432
     username = "user"
     password = "pass"
     dbName = "mydb"
     sslmode = "require"
   }

   // ... configure all other sections
   ```

3. Use it:

   ```bash
   CONFIG_ENV=myenv go run .
   ```

## Code Generation

After modifying any schema files (`pkl/schema/*.pkl`), regenerate Go code:

```bash
# Using Task
task gen

# Or manually
pkl-gen-go ./pkl/schema/Root.pkl --base-path config
```

This regenerates all files in `gen/*/`. **Never edit generated files manually.**

## Pkl Features Used

### Environment Variables

Read environment variables in Pkl configs:

```pkl
// Read with fallback to null if not set
callbackUrl = read?("env:GITHUB_CALLBACK_URL")

// Read with default value
port = read?("env:PORT") ?? "8080"
```

### Cross-Reference Between Sections

```pkl
core {
  rootdir = "/app"
  tmpdir = rootdir + "/tmp"
}

otel {
  serviceName = core.appKey  // Reference core section
}

logger {
  localConfig {
    logFile = core.tmpdir + "/logs/app.log"  // Reference core.tmpdir
  }
}
```

### Type Aliases

```pkl
typealias Environment = "dev" | "stag" | "prod"
```

## Troubleshooting

### "pkl: command not found"

Ensure Pkl is in your PATH:

```bash
export PATH="$PATH:/usr/local/bin"
```

### "module not found" when loading config

The config loader searches up to 3 directory levels for the `config` directory. Ensure you're running from within the project.

### Type mismatch errors

After changing the schema, regenerate Go code:

```bash
task gen
```

### Import errors in generated code

Ensure all schema files have correct `@go.Package` annotations and are properly imported in `Root.pkl`.

## Resources

* [Pkl Language Documentation](https://pkl-lang.org/main/current/index.html)
* [Pkl Go Bindings](https://pkl-lang.org/go/current/index.html)
* [pkl-gen-go Reference](https://pkl-lang.org/go/current/pkl-gen-go.html)
* [Pkl GitHub Repository](https://github.com/apple/pkl)
* [Pkl Go Repository](https://github.com/apple/pkl-go)
