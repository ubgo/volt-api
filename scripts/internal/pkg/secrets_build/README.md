# Secrets Build Package

## Overview

The `secrets_build` package automates the management of production keys and secrets for GitHub Actions deployments. It eliminates the need to manually update multiple configuration files when adding or rotating keys, ensuring consistency across the deployment pipeline.

## Problem Statement

In a typical deployment workflow, production keys (such as private keys, credentials, certificates) need to be:

1. **Stored locally** in `_keys/` with potentially varying names (e.g., `github_2024_xyz.pem`, `github_2025_abc.pem`)
2. **Base64 encoded** and stored in GitHub Secrets for secure transmission
3. **Referenced in GitHub Actions workflows** as environment variables
4. **Deployed to Docker containers** with standardized, predictable filenames (e.g., `github_2024_prod.pem`)

Without automation, this process requires manual updates in multiple places:

* `.github/.secrets` file
* `.github/workflows/deploy.yml` (env vars and script sections)
* Docker deployment scripts
* PKL configuration files

This is error-prone and time-consuming, especially when rotating keys or adding new ones.

## Solution

This package provides a unified automation system that:

1. **Auto-discovers** all production keys matching the pattern `*.prod.*` in the `_keys/` directory
2. **Normalizes key names** to a standard format (e.g., `github_2024_xyz.prod.pem` → `github_2024_prod.pem`)
3. **Base64 encodes** key contents automatically
4. **Generates GitHub Secrets** entries in `.github/.secrets`
5. **Updates GitHub Actions workflow** (`.github/workflows/deploy.yml`) with:
   * Environment variable declarations
   * Deployment scripts to decode and save keys
   * Dynamic `envs` list for the SSH action

## How It Works

### Key Detection

The package scans the `_keys/` directory for files matching the pattern `*.prod.*` (e.g., `github_2024_xyz.prod.pem`, `nsg-default-cli.prod.creds`).

### Name Normalization

Original key names are normalized to a unified format:

* **Input**: `github_2024_xyz.prod.pem`
* **Secret/Env Var Name**: `GITHUB_2024_XYZ_PROD_KEY_B64` (same name used in both `.secrets` and `deploy.yml`)

The normalization process:

1. Extracts the base name before `.prod.` (e.g., `github_2024_xyz`)
2. Replaces the last segment with `prod` (e.g., `github_2024_xyz` → `github_2024_prod`)
3. Preserves the file extension (e.g., `.pem`, `.creds`)

### Secret Generation

The package generates secrets in two ways:

**Production Key Files** (from `_keys/` directory):

For each detected `*.prod.*` key file:

1. Reads the file content
2. Base64 encodes the content
3. Generates a secret key name: `{UPPERCASE_BASE_NAME}_KEY_B64` (e.g., `GITHUB_2024_XYZ_PROD_KEY_B64`)
4. Adds to `.github/.secrets` file with the same name that will be used in `deploy.yml`

**Environment Configuration**:

1. Reads `.env.prod` file
2. Base64 encodes the entire file content
3. Adds as `ENV_PROD_B64` to `.github/.secrets`
4. This allows the deployment script to decode and write the environment file on the server

### Workflow Generation

The package generates the deployment workflow by:

1. Adding environment variables to the `deploy` job's `env` section:
   ```yaml
   ENV_PROD_B64: ${{ secrets.ENV_PROD_B64 }}
   GITHUB_2024_XYZ_PROD_KEY_B64: ${{ secrets.GITHUB_2024_XYZ_PROD_KEY_B64 }}
   ```
   Note: The same variable name is used in both `.secrets` and `deploy.yml` for consistency.
2. Generating shell scripts to decode and save files:
   ```bash
   # Decode and save .env.prod
   echo "${ENV_PROD_B64}" | base64 -d > ${DEPLOY_PATH}/.env.prod
   chmod 600 ${DEPLOY_PATH}/.env.prod

   # Decode and save key files
   echo "${GITHUB_2024_XYZ_PROD_KEY_B64}" | base64 -d > ${DEPLOY_PATH}/_keys/github_2024_xyz.prod.pem
   chmod 600 ${DEPLOY_PATH}/_keys/*
   ```
3. Updating the `envs` list to include all generated environment variables (including `ENV_PROD_B64` and all key file variables)

## Usage

### Command

```bash
go run ./scripts gh:secrets:update
```

This command:

1. Reads production secrets from `.github/.secrets.prod`
2. Scans `_keys/` for `*.prod.*` files and encodes them
3. Reads and base64 encodes `.env.prod` as `ENV_PROD_B64`
4. Generates/updates `.github/.secrets` with all base64-encoded values
5. Regenerates `.github/workflows/deploy.yml` from the template

### Workflow

1. **Add a new key**: Place it in `_keys/` with the pattern `{name}.prod.{ext}`
   ```bash
   cp github_2024_new_key.pem _keys/github_2024_new_key.prod.pem
   ```

2. **Run the update command**:
   ```bash
   go run ./scripts gh:secrets:update
   ```

3. **Review generated files**:
   * `.github/.secrets` - Contains base64-encoded keys
   * `.github/workflows/deploy.yml` - Updated with new env vars and scripts

4. **Push secrets to GitHub** (if needed):
   ```bash
   # Use GitHub CLI to push secrets
   gh secret set GITHUB_2024_KEY < .github/.secrets
   # Or use your preferred method to sync .github/.secrets to GitHub Secrets
   ```

## Benefits

### 1. Single Source of Truth

* Keys are stored once in `_keys/` with descriptive names
* The system automatically handles normalization and encoding

### 2. Consistency

* All keys are deployed with predictable, unified names
* PKL configuration can reference standardized paths (e.g., `/app/_keys/github_2024_prod.pem`)

### 3. Reduced Manual Errors

* No need to manually update multiple files
* Automatic base64 encoding ensures correct format

### 4. Easy Key Rotation

* When rotating a key, simply replace the file in `_keys/`
* Run the update command - everything else is handled automatically

### 5. Scalability

* Adding new keys requires no code changes
* The system automatically discovers and processes all `*.prod.*` files

## File Structure

```
_keys/
  ├── github_2024_xyz.prod.pem      # Original key (any name)
  ├── nsg-default-cli.prod.creds    # Another key
  └── ...

.github/
  ├── .secrets                      # Auto-generated (base64 encoded)
  ├── .secrets.prod                # Manual secrets (non-keys)
  └── workflows/
      ├── deploy.yml.gotmpl        # Template
      └── deploy.yml               # Auto-generated
```

## Example

### Input

**File**: `_keys/github_2024_xyz.prod.pem`

```
-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQC...
-----END PRIVATE KEY-----
```

### Generated Output

**`.github/.secrets`**:

```
ENV_PROD_B64="CiMgU2VydmVyIENvbmZpZ3VyYXRpb24KSE9TVE5BTUU9bG9jYWxob3N0..."
GITHUB_2024_XYZ_PROD_KEY_B64="LS0tLS1CRUdJTiBQUklWQVRFIEtFWS0tLS0t..."
```

**`.github/workflows/deploy.yml`** (excerpt):

```yaml
env:
  ENV_PROD_B64: ${{ secrets.ENV_PROD_B64 }}
  GITHUB_2024_XYZ_PROD_KEY_B64: ${{ secrets.GITHUB_2024_XYZ_PROD_KEY_B64 }}
  # ... other env vars
envs: DEPLOY_PATH,IMAGE,...,ENV_PROD_B64,GITHUB_2024_XYZ_PROD_KEY_B64
script: |
  echo "${ENV_PROD_B64}" | base64 -d > ${DEPLOY_PATH}/.env.prod
  chmod 600 ${DEPLOY_PATH}/.env.prod
  echo "${GITHUB_2024_XYZ_PROD_KEY_B64}" | base64 -d > ${DEPLOY_PATH}/_keys/github_2024_xyz.prod.pem
  chmod 600 ${DEPLOY_PATH}/_keys/*
```

**Docker Container**:

* Key is available at: `/app/_keys/github_2024_prod.pem`
* PKL config can reference: `/app/_keys/github_2024_prod.pem`

## Technical Details

### Key Naming Convention

* **Original File**: `{name}.prod.{ext}` (e.g., `github_2024_xyz.prod.pem`)
* **Secret/Env Var Name**: `{UPPERCASE_NAME}_KEY_B64` (e.g., `GITHUB_2024_XYZ_PROD_KEY_B64`)
  * The same name is used in both `.github/.secrets` and `.github/workflows/deploy.yml`
  * This ensures consistency and makes it clear that the variable contains base64-encoded content

### Base64 Encoding

Keys are encoded using `base64.StdEncoding` to ensure compatibility with GitHub Actions and shell scripts.

### Template System

The package uses Go's `text/template` with custom delimiters (`[[` and `]]`) to avoid conflicts with GitHub Actions' `${{ }}` syntax.

## Maintenance

### Adding a New Key Type

1. Place the key file in `_keys/` with pattern `*.prod.*`
2. Run `go run ./scripts gh:secrets:update`
3. No code changes required - the system auto-discovers it

### Rotating a Key

1. Replace the key file in `_keys/` (keep the same name or update it)
2. Run `go run ./scripts gh:secrets:update`
3. Push updated secrets to GitHub using your preferred method (e.g., GitHub CLI)

### Removing a Key

1. Delete the key file from `_keys/`
2. Run `go run ./scripts gh:secrets:update`
3. The system automatically removes it from generated files

## See Also

* [Deployment Guide](../../../../docs/DEPLOY_GUIDE.md) - For deployment workflows
* [Taskfile.yml](../../../../Taskfile.yml) - For available CLI commands
