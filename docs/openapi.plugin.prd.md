# OpenAPI Plugin — PRD

## Overview

Generate OpenAPI 3.1 specifications from Go source code via plego's Subscription/Filter/Publish pipeline.

## Pipeline

```
Subscription::Codebase → Filter::OpenAPI → Publish::Spec
```

## Plugin Specs

### Subscription::Codebase

Scans a Go project and emits one Entry per `.go` file.

#### Config

```yaml
- module: Subscription::Codebase
  config:
    path: ./cmd/api           # required: root directory
    patterns:                 # optional: default ["**/*.go"]
      - "**/*.go"
    exclude:                  # optional: globs to skip
      - "**/*_test.go"
      - "**/vendor/**"
```

#### Entry schema

| Field | Value |
|-------|-------|
| GUID  | relative file path |
| Title | filename |
| URL   | absolute file path |
| Body  | raw source text |

---

### Filter::OpenAPI

Parses Go source into an OpenAPI 3.1 document. Two annotation modes:

#### Mode A: Swagger-style comments (default)

```go
// ListUsers returns all users.
// @Summary  List users
// @Param   limit query int false "max results"
// @Success 200 array User
// @Router  /users [get]
func ListUsers(w http.ResponseWriter, r *http.Request) {
```

#### Mode B: Framework introspection (chi/gin/echo)

Auto-detects route registrations and extracts handler types without annotations.

```go
r.Get("/users", ListUsers)           // → GET /users
r.Post("/users", CreateUser)         // → POST /users
```

#### Config

```yaml
- module: Filter::OpenAPI
  config:
    title: "My API"
    version: "1.0.0"
    description: ""
    servers:
      - url: https://api.example.com
    format: 3.1                       # 3.0 or 3.1 (default 3.1)
    annotations: true                 # parse swagger-style comments
    introspection: true               # detect routes from framework calls
    frameworks:                       # framework auto-detection (default all)
      - chi
      - gin
      - echo
    type_resolution: shallow          # shallow | deep | none
```

#### Detection pipeline

```
1. Parse file → Go AST
2. Find route registration calls (r.Get, r.Post, r.Handle, etc.)
3. Resolve handler reference to function declaration
4. Extract request/response types from func signature
5. If annotations enabled: parse // @Summary, // @Param, etc.
6. If introspection enabled: follow framework patterns
7. Accumulate into OpenAPI document
```

#### Entry output (Body)

JSON-encoded OpenAPI 3.1 document.

```json
{
  "openapi": "3.1.0",
  "info": { "title": "My API", "version": "1.0.0" },
  "paths": {
    "/users": {
      "get": {
        "summary": "List users",
        "parameters": [...],
        "responses": { "200": { ... } }
      }
    }
  }
}
```

---

### Publish::Spec

Writes the OpenAPI document to disk in specified formats.

#### Config

```yaml
- module: Publish::Spec
  config:
    outdir: ./docs
    filename: openapi              # base name (default: openapi)
    formats:                       # default: [yaml]
      - yaml                       # → openapi.yaml
      - json                       # → openapi.json
      - markdown                   # → openapi.md (via redoc-cli or widdershins)
    pretty: true                   # indent output (default: true)
```

---

## Example pipeline

```yaml
pipeline:
  plugins:
    - module: Subscription::Codebase
      config:
        path: ./cmd/api
        exclude: ["**/*_test.go"]

    - module: Filter::OpenAPI
      config:
        title: "My API"
        version: "1.0.0"
        servers:
          - url: https://api.example.com
        annotations: true
        introspection: true

    - module: Publish::Spec
      config:
        outdir: ./docs
        formats: [yaml, json, markdown]
```

## Implementation plan

### Phase 1 — Core (MVP)

| Step | Task | Est. |
|------|------|------|
| 1.1 | Subscription::Codebase — file walker + Entry emit | 1d |
| 1.2 | Filter::OpenAPI — AST parser for `net/http` + Swagger comments | 3d |
| 1.3 | Publish::Spec — YAML/JSON writer | 0.5d |
| 1.4 | Integration test: generate spec from real handler | 0.5d |

### Phase 2 — Framework support

| Step | Task | Est. |
|------|------|------|
| 2.1 | chi router detection | 2d |
| 2.2 | gin router detection | 2d |
| 2.3 | echo router detection | 2d |
| 2.4 | Type resolution: struct field → JSON Schema | 3d |

### Phase 3 — Polish

| Step | Task | Est. |
|------|------|------|
| 3.1 | Markdown output via redoc-cli | 1d |
| 3.2 | Watch mode (yaqqle-style) | 0.5d |
| 3.3 | `plego.example.yaml` with OpenAPI pipeline | 0.5d |

## Open questions

1. Should type resolution handle generic types (`Some[T]`)?
2. How to handle external types from imported packages?
3. Should Publish::Spec bundle redoc-cli or shell out?
4. Do we need `Publish::Redoc` as a separate HTML renderer?
