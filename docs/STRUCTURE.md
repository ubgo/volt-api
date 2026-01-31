# Project Structure Guide

This document provides a detailed breakdown of the `ai_golang_boilerplate` directory structure. Use this as a map to understand where specific logic resides.

## Root Directory

| Path | Description |
| :--- | :--- |
| **`saas/`** | The core business logic, database schemas (Ent), and migrations. This is the most critical module. |
| **`apidash/`** | The main GraphQL API Gateway. Handles HTTP requests and implements the business logic. |
| **`lace/`** | Shared utility library. Contains common code used across all microservices (Logger, NATS, Cache). |
| **`config/`** | Configuration manager. Generates Go bindings from Apple Pkl configuration files. |
| **`.docker/`** | Docker Compose files for local development infrastructure (Postgres, NATS). |

***

## Module Details

### `saas/` (The Core)

* **`schema/`**: **SOURCE**. Edit this to change database fields/tables.
* **`gen/ent/`**: **GENERATED**. The Ent ORM code. **DO NOT EDIT**.
* **`pkg/`**: Business logic packages (e.g., `user`, `auth`).

### `apidash/` (The API Gateway)

* **`internal/graph/schema.graphqls`**: **SOURCE**. The GraphQL schema definitions.
* **`internal/graph/generated/`**: **GENERATED**. GraphQL boilerplate code. **DO NOT EDIT**.
* **`internal/graph/resolverfn/`**: **SOURCE**. Implement your API resolvers here.

### `config/` (Configuration)

* **`config/`**: **SOURCE**. Pkl files defining the app configuration structure.
* **`gen/`**: **GENERATED**. Go structs corresponding to the Pkl config.

### `lace/` (Shared Lib)

* **`nlog/`**: Structured logging wrapper (Zap).
* **`natso/`**: NATS JetStream wrapper for event-driven architecture.
* **`threadsafecache/`**: In-memory caching utilities.
