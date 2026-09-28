---
name: golang-swagger
description: "OpenAPI/Swagger docs for Golang with swaggo/swag — annotation comments, swag init, framework integrations, security definitions, and struct tags. Apply when adding or maintaining Swagger docs in a Go API, or when the codebase imports `github.com/swaggo/swag` or a swaggo framework adapter."
user-invocable: true
license: MIT
compatibility: Designed for Claude Code, Codex or similar harness. Requires go and swag CLI.
metadata:
  author: samber
  version: "1.1.5"
  openclaw:
    emoji: "📋"
    homepage: https://github.com/samber/cc-skills-golang
    requires:
      bins:
        - go
        - swag
    install:
      - kind: go
        package: github.com/swaggo/swag/cmd/swag@latest
        bins: [swag]
    skill-library-version: "2.0.0-rc5"
allowed-tools: Read Edit Write Glob Grep Bash(go:*) Bash(golangci-lint:*) Bash(git:*) Agent WebFetch mcp__context7__resolve-library-id mcp__context7__query-docs Bash(swag:*) AskUserQuestion Bash(godig:*) Bash(gopls:*) LSP mcp__gopls__*
paths:
  - "**/*.go"
---

**Persona:** You are a Go API documentation engineer. You treat docs as a contract — accurate, complete annotations prevent integration bugs and make the Swagger UI the source of truth for API consumers.

**Modes:**

- **Build** — adding Swagger to a new or existing Go project. Done when `swag init` succeeds and `/swagger/index.html` lists every routed handler.
- **Audit** — done when every routed handler has `@Router`, `@Success` and `@Failure`, protected routes have `@Security`, and `swag init` regenerates `docs/` with no diff. Deliver ranked findings with file:line; if the user asked for fixes, apply them and re-run `swag init`.

**Dependencies:**

- swag: `go install github.com/swaggo/swag/cmd/swag@latest`

## Setup

```bash
swag init                        # generates docs/ with docs.go, swagger.json, swagger.yaml
swag init -g cmd/api/main.go     # if general info is not in main.go
swag fmt                         # format annotation comments (like go fmt)
```

Import the generated `docs` package to register the spec — blank (`_ "yourmodule/docs"`) when only serving the UI, named (`docs "yourmodule/docs"`) when overriding `docs.SwaggerInfo` at runtime. Then mount the UI, e.g. for Gin (`gin-swagger` + `swaggo/files`), and open `/swagger/index.html`:

```go
r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
```

Read [swag-cli.md](references/swag-cli.md) for other framework adapters (Echo, Fiber, Chi, net/http — each wires differently), a host or base path set per environment, generic or envelope response types (`api.Response[model.User]`, `api.Envelope{data=model.User}`), response headers, MIME aliases, and `swag init` flags such as `--tags '!Internal'`.

## General API Info

Place in `main.go` (or the file passed via `-g`). These annotations define the top-level spec:

```go
// @title           My API
// @version         1.0
// @description     Short description of the API.
// @host            localhost:8080
// @BasePath        /api/v1
// @schemes         http https

// @contact.name    API Support
// @contact.email   support@example.com
// @license.name    Apache 2.0

// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and the JWT token.
```

## Operation Annotations

Annotate each handler function. The standard doc comment (`// FuncName godoc`) must precede swag annotations — it anchors indentation for `swag fmt`.

```go
// ShowAccount godoc
// @Summary      Get account by ID
// @Description  Returns account details for the given ID.
// @Tags         accounts
// @Accept       json
// @Produce      json
// @Param        id      path  int  true  "Account ID"
// @Param        filter  query string false "Optional search filter"
// @Success      200  {object}  model.Account
// @Success      204  "No content"
// @Failure      400  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Router       /accounts/{id} [get]
// @Security     Bearer
func ShowAccount(c *gin.Context) {}
```

**@Param** format: `@Param <name> <in> <type> <required> "<description>" [attributes]`

| `<in>`     | Usage                                |
| ---------- | ------------------------------------ |
| `path`     | URL path segment (`/users/{id}`)     |
| `query`    | URL query string (`?filter=x`)       |
| `body`     | Request body — type must be a struct |
| `header`   | HTTP header                          |
| `formData` | Multipart/form field                 |

Optional attributes on `@Param`: `default(v)`, `minimum(n)`, `maximum(n)`, `minLength(n)`, `maxLength(n)`, `Enums(a,b,c)`, `example(v)`, `collectionFormat(multi)`.

**@Success/@Failure** format: `@Success <code> {<kind>} <type> "<description>"`

| `<kind>`             | When             |
| -------------------- | ---------------- |
| `{object}`           | Single struct    |
| `{array}`            | Slice of structs |
| `string` / `integer` | Primitive        |

## Security Definitions

Define once at the API level (in main.go), apply per endpoint with `@Security`.

```go
// Bearer / JWT
// @securityDefinitions.apikey Bearer
// @in header
// @name Authorization

// API key in header
// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name X-API-Key

// Basic auth
// @securityDefinitions.basic BasicAuth

// OAuth2 authorization code
// @securityDefinitions.oauth2.authorizationCode OAuth2
// @authorizationUrl https://example.com/oauth/authorize
// @tokenUrl https://example.com/oauth/token
// @scope.read Read access
// @scope.write Write access
```

Apply to an endpoint:

```go
// @Security Bearer
// @Security OAuth2[read, write]
// @Security BasicAuth && ApiKeyAuth   // AND — both required
```

## Struct Tags

Enrich models without changing their Go type:

```go
type CreateUserRequest struct {
    Name   string `json:"name" example:"Jane Doe" minLength:"2" maxLength:"100"`
    Role   string `json:"role" enums:"admin,user,guest" example:"user"`
    Age    int    `json:"age" minimum:"18" maximum:"120"`
    Avatar []byte `json:"avatar" swaggertype:"string" format:"base64"`
    Secret string `json:"-" swaggerignore:"true"`  // excluded from docs
}
```

| Tag | Purpose |
| --- | --- |
| `example` | Example value shown in Swagger UI |
| `enums` | Comma-separated allowed values |
| `swaggertype` | Override detected type (e.g., `"primitive,integer"` for `time.Time`) |
| `swaggerignore:"true"` | Exclude field from the generated schema |
| `extensions` | Add OpenAPI extensions: `extensions:"x-nullable,x-deprecated=true"` |

## Common Mistakes

| Mistake | Why it breaks | Fix |
| --- | --- | --- |
| Missing `_ "yourmodule/docs"` import | Schema not registered; UI loads empty | Add blank import in main.go or server init |
| Stale `docs/` after code changes | Docs diverge from implementation; consumers get wrong schema | Re-run `swag init` after every annotation change |
| Model type from a dependency (`uuid.UUID`, `sql.NullString`, `gorm.Model`) | swag parses only the `-d` directories, so `swag init` fails with `cannot find type definition` | Add `--parseDependency`, or override the field with `swaggertype` |
| No `@Security` on protected routes | Swagger UI shows no lock icon; testers send unauthenticated requests | Apply `@Security` to every authenticated endpoint |
| General info annotations in the wrong file | swag silently skips them; spec has no title/host | Use `-g <file>` flag or move annotations to `main.go` |
| Raw map in `@Success` (`{object} map[string]any`) | Generates, but as an anonymous `additionalProperties` object that documents nothing | Use a named struct when keys are known, a named map type (`type FeatureFlags map[string]bool`) when they are dynamic |
| Quoting `@Tags` or separating them with spaces | `@Tags` splits on commas only — quotes become part of the tag name, and `@Tags users admin` is one tag named "users admin" | `@Tags users,admin` |

## Cross-References

- → See `samber/cc-skills-golang@golang-security` for securing the Swagger UI endpoint in production (disable or gate with auth middleware).
- → See `samber/cc-skills-golang@golang-grpc` for gRPC — use grpc-gateway with its own OpenAPI generator instead of swag.

This skill is not exhaustive — refer to the swaggo/swag documentation and code examples for up-to-date API signatures and usage patterns:

- For Go package docs, symbols, versions, importers, and known vulnerabilities, → See `samber/cc-skills-golang@golang-pkg-go-dev` skill (`godig`), preferred over Context7 for Go package facts.
- To navigate this library's usage in your own code (definitions, call sites, diagnostics), → See `samber/cc-skills-golang@golang-gopls` skill (`gopls`).
- Context7 remains a fallback for docs not indexed on pkg.go.dev.

If you encounter a bug or unexpected behavior in swag, open an issue at <https://github.com/swaggo/swag/issues>.
