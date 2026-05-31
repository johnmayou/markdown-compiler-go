# markdown-compiler-go

A Markdown-to-HTML compiler written in Go.

## How It Works

The compiler follows a classic pipeline:

1. Lexer (Tokenizer): Scans the raw Markdown input and produces a list of tokens.
2. Parser: Processes tokens into an abstract syntax tree (AST) representing the document structure.
3. Code Generator: Converts the AST into the target language (HTML).

## Usage

Install:

```bash
go get github.com/johnmayou/markdown-compiler-go
```

Compile:

```go
import "github.com/johnmayou/markdown-compiler-go/pkg/markitdown"

html, err := markitdown.Compile("# Hello World")
// html => "<h1>Hello World</h1><hr>"
```

## Supported Syntax

> [!NOTE]
> Common wrapping tags omitted for brevity.

| Markdown                | HTML                                     |
| ----------------------- | ---------------------------------------- |
| `# Heading`             | `<h1>Heading</h1><hr>`                   |
| `## Heading`            | `<h2>Heading</h2><hr>`                   |
| `**bold**` / `__bold__` | `<b>bold</b>`                            |
| `*italic*` / `_italic_` | `<i>italic</i>`                          |
| `***bold italic***`     | `<b><i>bold italic</i></b>`              |
| `` `code` ``            | `<code>code</code>`                      |
| `[text](href)`          | `<a href="href">text</a>`                |
| `![alt](src)`           | `<img alt="alt" src="src"/>`             |
| `- item` / `* item`     | `<ul><li>item</li></ul>`                 |
| `1. item`               | `<ol><li>item</li></ol>`                 |
| `> quote`               | `<blockquote>quote</blockquote>`         |
| `---` / `***`           | `<hr>`                                   |
| ` ```lang ``` `         | `<pre><code class="lang">…</code></pre>` |

Unordered and ordered lists support nesting via 2-space indentation. Block quotes support nesting via `> >`.

## Development

### Prerequisites

- Go 1.26+
- [prettier](https://prettier.io) — used by the test suite to normalize HTML before golden-file comparison
- [golangci-lint](https://golangci-lint.run) v2.12.2 — installed automatically by `make lint` if missing

### Commands

| Command                       | Description                     |
| ----------------------------- | ------------------------------- |
| `make fmt`                    | Format all Go source files      |
| `make lint`                   | Run `go vet` + golangci-lint    |
| `make flint`                  | Format then lint                |
| `make test`                   | Run all tests                   |
| `make test-update`            | Rewrite golden files            |
| `make release VERSION=vX.Y.Z` | Tag and push to trigger release |

### Testing

Tests use a golden file strategy. Each fixture in `pkg/markitdown/testdata/` is a `.md` / `.golden.html` pair. `TestCompile` discovers all `.md` files automatically, compiles each one, prettifies the HTML output via `prettier`, and compares it against the corresponding `.golden.html`.

To add a new test case, drop a `.md` file into `testdata/` and run `make test-update` to generate its golden file. To update all golden files after an intentional output change, run `make test-update` and review the diff before committing.

## License

[MIT](LICENSE)
