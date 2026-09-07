# Gofreight Templates (GFT)

**Gofreight Templates** (`.gft`) is Gofreight's native view language. It provides layouts, partials, loops, conditionals, and CSRF helpers with its **own syntax**, and compiles to Go `html/template` at load time.

## File extension

| Extension | Description |
|-----------|-------------|
| `.gft` | Gofreight Template (recommended) |
| `.html` | Compiled when GFT directives are detected |

Store views in `app/views/`:

```
app/views/
├── layouts/
│   └── application.gft
├── partials/
│   └── flash.gft
└── posts/
    ├── index.gft
    ├── show.gft
    ├── new.gft
    └── edit.gft
```

## Output

| Syntax | Meaning |
|--------|---------|
| `{= .Title }` | Escaped output (HTML-safe) |
| `{! .HTML !}` | Raw / unescaped output |

## Layouts & slots

Child view:

```gft
#layout "layouts.application"

#slot "title"
All Posts
#endslot

#slot "content"
  <h1>Posts</h1>
#endslot
```

Layout (`app/views/layouts/application.gft`):

```gft
<!DOCTYPE html>
<html>
<head>
  <title>#place "title" "My App"</title>
</head>
<body>
  #partial "partials.flash"
  <main>#place "content"</main>
</body>
</html>
```

| Directive | Purpose |
|-----------|---------|
| `#layout "path"` | Inherit a layout (dots → slashes) |
| `#slot "name"` … `#endslot` | Define a content region |
| `#place "name"` | Render a slot in the layout |
| `#place "name" "default"` | Slot with fallback text |
| `#partial "path"` | Include a partial template |

## Conditionals

```gft
#when .Published
  <span>Live</span>
#orwhen .Scheduled
  <span>Scheduled</span>
#otherwise
  <span>Draft</span>
#endwhen

#unless .Deleted
  <article>...</article>
#endunless

#signedin
  <a href="/logout">Logout</a>
#endsignedin

#signedout
  <a href="/login">Login</a>
#endsignedout
```

## Loops

```gft
#each .Posts as post
  <h2>{= .Title }</h2>
  <p>{= .Body }</p>
#endeach

#eachor .Posts as post
  <li>{= .Title }</li>
#otherwise
  <li>No posts yet.</li>
#endeach
```

Inside `#each` / `#eachor`, use `{= .Field }` — the dot is the current item in the loop. The `as post` name is optional documentation; output always uses `.`.

## Forms & security

Use `#form` for CSRF-protected forms with optional PUT/PATCH/DELETE spoofing:

```gft
#form action="/posts" method="POST"
  #field "title" label="Title" type="text" value=".Item.Title"
  <button type="submit">Save</button>
#endform
```

`#field` renders label, input, old values, and validation errors. See **[Forms & validation](forms-validation.md)**.

Legacy CSRF token:

```gft
<form method="POST" action="/posts">
  #token
  <input name="title">
</form>
```

`#token` emits a CSRF hidden field (`authenticity_token`).

## Comments

```gft
{# This comment is stripped at compile time #}
```

## Helpers

Register custom helpers on the view engine:

```go
app.Views.RegisterFunc("money", func(n float64) string {
    return fmt.Sprintf("$%.2f", n)
})
```

Use in templates after compile: `{= money .Price }` → `{{money .Price}}`

Built-in helpers: `upper`, `lower`, `title`, `join`, `default`, `safeHTML`, `contains`, `trim`, `old`, `fieldErrors`, `hasError`.

## Rendering

```go
return base.RenderView("posts/index", map[string]any{
    "Posts": posts,
    "Flash": "Saved!",
})
```

Partial without layout:

```go
return base.RenderPartial("partials.flash", data)
```

## Design philosophy

GFT is Gofreight's own view language — not a port of another template engine. Directives use a `#` prefix; output uses `{= }` (escaped) and `{! !}` (raw).

| Concept | GFT directive |
|---------|---------------|
| Layout inheritance | `#layout` |
| Content regions | `#slot` / `#place` |
| Partials | `#partial` |
| Conditionals | `#when` / `#endwhen` |
| Loops | `#each` / `#endeach` |
| Empty fallback | `#eachor` / `#otherwise` |
| CSRF | `#token` / `#form` |
| Vite assets | `#vite "resources/js/app.js"` |
| Forms | `#form` / `#field` / `#error` |
| Validation helpers | `old`, `fieldErrors`, `hasError` |
| Output | `{= }` / `{! !}` |

GFT is designed to be readable on its own while fitting naturally into Go projects that already use `html/template` under the hood.

## Vite (frontend bundler)

When using [Vite](https://vitejs.dev/) with a `public/hot` file, enable assets in `bootstrap/app.go`:

```go
app.UseVite()
```

In layouts, load an entry point:

```html
#vite "resources/js/app.js"
```

In development (with `public/hot` present), Gofreight proxies `/@vite` paths to `VITE_DEV_SERVER_URL` and injects the Vite client scripts. In production, use a built manifest via `app.LoadAssetManifest("public/manifest.json")`.
