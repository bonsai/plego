# ADR-001: Codebase → OpenAPI → Spec pipeline

## Status

Proposed

## Context

We want to extend plego's Subscription/Filter/Publish model to generate OpenAPI specifications directly from Go source code. This enables:

- API spec as a pipeline output (not a manual document)
- Spec regeneration on code change via `watch`
- Reuse of existing Go types and handler signatures
- Integration with documentation generators

## Decision

Add three new plugins following the same pattern as yaqqle:

### Subscription::Codebase

Scans a Go project directory and produces one Entry per Go file (or one Entry per package).

```
Subscription::Codebase:
  path: ./cmd/api
  pattern: "**/*.go"
```

Entry schema:
- GUID: file path (relative to root)
- Title: file name
- Body: raw source text
- URL: absolute file path

### Filter::OpenAPI

Parses Go source to extract route definitions, request/response types, and handler signatures. Produces an OpenAPI 3.1 document.

Detection strategy:
1. Parse Go AST for `http.Handler`, `http.HandlerFunc`, or framework-specific patterns (chi, gin, echo)
2. Extract method + path from route registrations
3. Extract request/response struct types from handler signatures
4. Read `// @Summary`, `// @Param` etc. from comments (Swagger-style annotations)

Configuration options:
```
Filter::OpenAPI:
  title: "My API"
  version: "1.0.0"
  description: ""
  servers:
    - url: https://api.example.com
  annotations: true   # parse Swagger-style comments
```

Entry output (Body): JSON-encoded OpenAPI 3.1 document.

### Publish::Spec

Writes the OpenAPI document in one or more formats.

```
Publish::Spec:
  outdir: ./docs
  formats:           # default: [yaml]
    - yaml           # openapi.yaml
    - json           # openapi.json
    - markdown       # openapi.md (via external converter)
```

### Pipeline example

```yaml
pipeline:
  plugins:
    - module: Subscription::Codebase
      config:
        path: ./cmd/api
    - module: Filter::OpenAPI
      config:
        title: "My API"
        version: "1.0.0"
    - module: Publish::Spec
      config:
        outdir: ./docs
        formats: [yaml, markdown]
    - module: Publish::Filesystem  # or Gmail, GoogleDrive, etc.
      config:
        path: ./docs/openapi.yaml
```

## Consequences

Positive:
- Single source of truth (code) drives API docs
- No manual spec drift
- Reuses plego's state, watch, and publish infrastructure
- New plugins are independently testable

Negative:
- Go AST parsing is non-trivial; initial version may only support standard `net/http`
- Framework-specific patterns (chi, gin, echo) require per-framework support
- Swagger annotation parsing adds complexity

## Alternatives considered

1. **External tool (swaggo/swag)**: Already exists but is not pipeline-native. Could be wrapped in a Filter plugin.
2. **Manual OpenAPI YAML**: The status quo. No automation, prone to drift.
3. **protoc + grpc-gateway**: Only applicable for protobuf-based services.

## Open questions

1. Should Filter::OpenAPI be a single plugin with framework adapters, or one plugin per framework?
2. Should Publish::Spec include a Markdown renderer, or delegate to an external tool?
3. How deep should type resolution go (embedded structs, generics, external types)?
