<!--
| --------------------------------------------------------------------------
| Localization
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
| Localization (i18n)
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Localization (i18n)

Gofreight includes a lightweight translation system for multi-language applications.

## Quick start

**1. Translation files** — JSON in `config/locales/`:

```json
// config/locales/en.json
{
  "welcome": "Welcome to Gofreight",
  "posts.count": "You have :count posts"
}
```

```json
// config/locales/fr.json
{
  "welcome": "Bienvenue sur Gofreight",
  "posts.count": "Vous avez :count articles"
}
```

**2. Bootstrap:**

```go
_ = app.LoadLocales("config/locales")
app.UseLocale()
```

**3. Translate in controllers or views:**

```go
msg := app.I18n.T("welcome", nil)
msg := app.I18n.T("posts.count", map[string]string{"count": "5"})
// → "You have 5 posts"
```

## Configuration

| Variable | Default | Purpose |
|----------|---------|---------|
| `APP_LOCALE` | `en` | Default locale |
| `APP_FALLBACK_LOCALE` | `en` | Fallback when key missing |

Set in `.env`:

```env
APP_LOCALE=en
APP_FALLBACK_LOCALE=en
```

## Translator API

```go
translator := i18n.New("en", "en")
_ = translator.LoadDir("config/locales")

translator.SetLocale("fr")
translator.Locale() // "fr"

translator.T("welcome", nil)                    // translated string
translator.T("missing.key", nil)                // returns key if not found
translator.T("posts.count", map[string]string{   // placeholder replacement
    "count": "12",
})
```

Placeholders use `:name` syntax in JSON values and are replaced via the `replacements` map.

## Locale middleware

`app.UseLocale()` adds middleware that detects locale from:

1. Query parameter `?lang=fr`
2. `Accept-Language` header
3. Session key (when configured)
4. Falls back to `APP_LOCALE`

The active locale is stored on the request context for the duration of the request.

## Using translations in GFT views

Register a template helper in bootstrap:

```go
app.Views.RegisterFunc("t", func(key string, replacements ...map[string]string) string {
    reps := map[string]string{}
    if len(replacements) > 0 {
        reps = replacements[0]
    }
    return app.I18n.T(key, reps)
})
```

In templates:

```html
<h1>{= t "welcome" }</h1>
<p>{= t "posts.count" (dict "count" "3") }</p>
```

Or pass translated strings from the controller via `ViewData`.

## File format

- One JSON file per locale: `en.json`, `fr.json`, `de.json`
- Flat key-value pairs (use dot notation for namespacing: `auth.login.title`)
- UTF-8 encoding

## Testing translations

```go
func TestWelcomeFrench(t *testing.T) {
    app := gftest.NewApp(t)
    _ = app.App.LoadLocales("config/locales")
    app.App.I18n.SetLocale("fr")
    gftest.Expect(app.App.I18n.T("welcome", nil)).ToContain("Bienvenue")
}
```

## Limitations

| Feature | Status |
|---------|--------|
| JSON file catalogs | Supported |
| Pluralization rules | Not built-in — use separate keys |
| ICU message format | Not supported |
| Auto-extraction from code | Not supported |

For complex i18n needs, integrate an external library via the service container.

## Related

- [Configuration](configuration.md) — `APP_LOCALE`
- [Templating](templating.md) — custom template helpers
- [Application wiring](application-wiring.md) — `LoadLocales`, `UseLocale`
