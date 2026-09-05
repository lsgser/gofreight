# Forms & validation

Gofreight provides schema-based validation and GFT form components for HTML flows — validate in controllers, redirect back with errors, and repopulate fields automatically.

## Vine schemas

Define validation rules with composable field rules:

```go
import "github.com/lsgser/gofreight/vine"

var createPostValidator = vine.Object(map[string]vine.Rule{
    "title": vine.String().Required().MinLength(3).MaxLength(255),
    "email": vine.String().Required().Email(),
    "status": vine.String().In("draft", "published"),
})
```

Available rule chains:

| Rule | Methods |
|------|---------|
| `vine.String()` | `Required()`, `Email()`, `MinLength(n)`, `MaxLength(n)`, `In(...)` |
| `vine.Number()` | `Required()` |
| `vine.Boolean()` | `Required()` |

## Validate in controllers

```go
func (c PostController) Create(base controller.Base) error {
    payload, err := base.ValidateUsing(createPostValidator)
    if err != nil {
        return nil // JSON 422 or redirect back with errors
    }
    // use payload map[string]string
}
```

`ValidateUsing`:

- **JSON requests** — returns `422` with `{ "errors": { ... } }`
- **HTML forms** — flashes errors + old input, redirects to `Referer`

Helpers:

```go
base.ViewData(map[string]any{"Item": item}) // merges CSRF, errors, old input
base.RedirectBackWithErrors(errs, old)
base.HandleValidationFailure(errs, old)
```

## GFT form components

### `#form` — form wrapper with CSRF and method spoofing

```gft
#form action="/posts" method="POST"
  #field "title" label="Title" type="text" value=".Item.Title"
  <button type="submit">Create</button>
#endform
```

For updates, use `method="PUT"` — GFT emits `_method=PUT` and the framework rewrites the request method.

### `#field` — label, input, and inline errors

```gft
#field "email" label="Email" type="email" value=".Item.Email"
#field "body" label="Body" type="textarea" value=".Item.Body"
#field "status" label="Status" type="select" options="draft,published"
#field "published" label="Published" type="checkbox"
```

Each field renders:

- A label and input (or textarea/select/checkbox)
- Old input repopulation via the `old` helper
- Validation errors below the field

### `#error` — field errors only

```gft
#error "email"
```

### Manual helpers

```gft
{= old "title" . "" }
```

Template functions: `old`, `fieldErrors`, `hasError`.

## Session flash

On validation failure (HTML):

| Session key | Purpose |
|-------------|---------|
| `_validation_errors` | Field → messages (pulled into `.Errors` on next render) |
| `_old_input` | Submitted values (pulled into `.Old` on next render) |

`RenderView` automatically merges CSRF token, errors, old input, and flash messages via `ViewData` / `MergeRequestContext`.

## CSRF

Enable in bootstrap:

```go
app.UseCSRF()
```

`#form` and `#token` emit `authenticity_token` hidden fields. The CSRF middleware also accepts `_csrf` for compatibility.

## Generated scaffolds

`gofreight make:scaffold` generates:

- Vine validators in the controller
- `#form` / `#field` views
- `ValidateUsing` in create/update actions
- Redirect-back on validation failure

See also: [Templating](templating.md), [Routing](routing.md).
