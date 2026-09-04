# Gofreight Templates (GFT)

**Gofreight Templates** (`.gft`) is Gofreight's native view language. It takes inspiration from Laravel and Rails ergonomics — layouts, partials, loops, conditionals — but uses its **own syntax and naming**. GFT compiles to Go `html/template` at load time.

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

```gft
<form method="POST" action="/posts">
  #token
  <input name="title">
</form>
```

`#token` emits a CSRF hidden field (`{{.CSRFToken}}` after compile).

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

Built-in helpers: `upper`, `lower`, `title`, `join`, `default`, `safeHTML`, `contains`, `trim`.

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

GFT is **not** Blade or ERB. It is Gofreight's own language:

| Concept | GFT | Inspired by |
|---------|-----|-------------|
| Layout inheritance | `#layout` | Laravel `@extends` |
| Content regions | `#slot` / `#place` | `@section` / `@yield` |
| Partials | `#partial` | `@include` |
| Conditionals | `#when` / `#endwhen` | `@if` / `@endif` |
| Loops | `#each` / `#endeach` | `@foreach` |
| Empty fallback | `#eachor` / `#otherwise` | `@forelse` |
| CSRF | `#token` | `@csrf` |
| Output | `{= }` / `{! !}` | `{{ }}` / `{!! !!}` |

The `#` prefix and `{= }` output delimiters keep GFT visually distinct while remaining easy to learn if you know Laravel or Rails.
