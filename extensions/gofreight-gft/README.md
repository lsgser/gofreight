<!--
| --------------------------------------------------------------------------
| README
| --------------------------------------------------------------------------
|
| Part of the Gofreight project source tree.
|

-->

<!--
| --------------------------------------------------------------------------
| Gofreight GFT — VS Code Extension
| --------------------------------------------------------------------------
|
| Documentation for the Gofreight framework.
|

-->

# Gofreight GFT — VS Code Extension

Syntax highlighting, snippets, and editor support for [Gofreight Template](https://lsgser.github.io/gofreight-web/docs/templating) (`.gft`) files.

## Features

- **Syntax highlighting** for GFT directives, output expressions, comments, and HTML
- **Snippets** for layouts, loops, conditionals, forms, and common directives
- **Block folding** for `#slot`, `#form`, `#each`, `#when`, and related blocks
- **Auto-closing pairs** for `{= }`, `{! !}`, `{# #}`, and HTML tags

## Supported syntax

| Construct | Example |
|-----------|---------|
| Output | `{= .Title }` |
| Raw HTML | `{! .Body !}` |
| Comment | `{# note #}` |
| Layout | `#layout "layouts.application"` |
| Slots | `#slot "content"` … `#endslot`, `#place "content"` |
| Partials | `#partial "partials/flash"` |
| Loops | `#each .Posts as post` … `#endeach` |
| Conditionals | `#when`, `#unless`, `#signedin`, `#signedout` |
| Forms | `#form`, `#field`, `#error`, `#token` |
| Assets | `#vite "resources/js/app.js"` |

## Install locally

From the repository root:

```bash
cd extensions/gofreight-gft
npm install -g @vscode/vsce   # once, if you don't have vsce
vsce package
code --install-extension gofreight-gft-0.1.0.vsix
```

Or install directly from the folder during development:

1. Open `extensions/gofreight-gft` in VS Code / Cursor
2. Press **F5** to launch an Extension Development Host

## Workspace recommendation

Add to your app's `.vscode/extensions.json`:

```json
{
  "recommendations": ["gofreight.gofreight-gft"]
}
```

After publishing to the Marketplace, install with:

```
ext install gofreight.gofreight-gft
```

## Related

- [GFT templating docs](https://lsgser.github.io/gofreight-web/docs/templating)
- [Forms & validation](https://lsgser.github.io/gofreight-web/docs/forms-validation)
