<!--
| --------------------------------------------------------------------------
| Tutorial-html-crud
| --------------------------------------------------------------------------
|
| Official Gofreight documentation consumed by the docs site and linked
| from the CLI welcome page.
|
| Keep examples aligned with the current CLI and version; sync copies to
| gofreight-web when publishing.
|
| Covers concepts, tutorials, and reference material for this topic.
|

-->

<!--
| --------------------------------------------------------------------------
| Tutorial-html-crud
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

<p align="center">
  <img src="assets/gofreight-icon.png" alt="Gofreight" width="48">
</p>

# HTML CRUD with GFT

This tutorial builds a browser-based CRUD interface using Gofreight Templates (GFT), Vine validation, and form helpers.

## What you'll build

A posts manager with list, create, edit, and delete pages — including server-side validation and flash error messages.

## Prerequisites

Start from an app with migrations applied (see **[Your First App](tutorial-first-app.md)**):

```bash
gofreight make:scaffold Post title:string body:text published:boolean
gofreight migrate
gofreight serve
```

Visit **http://localhost:5000/posts** — the generator creates routes, controller actions, and GFT views.

## Step 1 — Web resource routes

In `routes/web.go`, a full HTML resource includes `new` and `edit` pages:

```go
r.Resources("posts", router.ResourceHandlers{
    Index:   controller.Handler(c.Index),
    Show:    controller.Handler(c.Show),
    Create:  controller.Handler(c.Create),
    Update:  controller.Handler(c.Update),
    Destroy: controller.Handler(c.Destroy),
})
```

| Method | Path | Page |
|--------|------|------|
| GET | `/posts` | Index (list) |
| GET | `/posts/new` | Create form |
| POST | `/posts` | Store |
| GET | `/posts/:id` | Show |
| GET | `/posts/:id/edit` | Edit form |
| PUT/PATCH | `/posts/:id` | Update |
| DELETE | `/posts/:id` | Destroy |

HTML forms use `_method=PUT` or `_method=DELETE` for updates and deletes (handled automatically by Gofreight middleware).

## Step 2 — GFT templates

Views live in `views/posts/`. A typical index template:

```html
<!DOCTYPE html>
<html>
<head><title>Posts</title></head>
<body>
  <h1>Posts</h1>
  <a href="/posts/new">New post</a>
  <ul>
    {{ range .Posts }}
    <li>
      <a href="/posts/{{ .ID }}">{{ .Title }}</a>
    </li>
    {{ end }}
  </ul>
</body>
</html>
```

Controllers pass data via `controller.ViewData`:

```go
controller.Render(w, r, "posts/index", controller.ViewData(r, map[string]any{
    "Posts": posts,
}))
```

## Step 3 — Form helpers

Gofreight provides `#form`, `#field`, `#token`, and `#error` for consistent forms with CSRF protection:

```html
#layout "layouts/application"

#form action="/posts" method="POST"
  #field "title" label="Title" type="text" value=".Post.Title"
  #error "title"

  #field "body" label="Body" type="textarea" value=".Post.Body"
  #error "body"

  #field "published" label="Published" type="checkbox"
  #token
  <button type="submit">Save</button>
#endform
```

`#token` emits the CSRF hidden field (`authenticity_token`).

## Step 4 — Vine validation

Define a schema and validate in your controller:

```go
var postSchema = vine.Object(map[string]vine.Rule{
    "title": vine.String().Required().MinLength(3).MaxLength(255),
    "body":  vine.String().Required(),
})

func (c PostsController) Store(base controller.Base) error {
    data, err := base.ValidateUsing(postSchema)
    if err != nil {
        return nil // ValidateUsing responds with redirect or 422
    }

    // create post from data...
    http.Redirect(w, r, "/posts", http.StatusSeeOther)
}
```

When validation fails, errors and old input are flashed to the session and shown on the form via `{{ error "title" }}`.

## Step 5 — Redirect back on failure

`RedirectBackWithErrors` sends the user back to the form with:

- Field errors under each input (`{{ error "field" }}`)
- Old input repopulated (`{{ field "title" .Old.Title "text" }}`)

Validation errors and old input are preserved across redirects — idiomatic Go with a smooth form UX.

## Step 6 — Flash messages

Show a success notice after creating a record:

```go
session.Flash(r, "success", "Post created!")
http.Redirect(w, r, "/posts", http.StatusSeeOther)
```

In your layout template, render flashed messages when present.

## Generators save time

The scaffold command generates validation, forms, and views for you:

```bash
gofreight make:scaffold Article title:string body:text status:enum:draft,published
```

Customize the generated controller and views from there.

## Next steps

- **[Forms & Validation](../docs/forms-validation.md)** — full Vine schema reference
- **[Templating (GFT)](../docs/templating.md)** — template syntax and helpers
- **[JWT Authentication](tutorial-auth-jwt.md)** — add an API alongside your HTML app
