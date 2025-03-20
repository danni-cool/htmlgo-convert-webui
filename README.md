# HTML to Go Converter

This project provides a web interface for converting HTML to Go code using the [htmlgo](https://github.com/theplant/htmlgo) library and [html2go](https://github.com/zhangshanwen/html2go) converter.

## Features

- Convert HTML to Go code with [htmlgo](https://github.com/theplant/htmlgo)
- Customizable package name prefix
- Web interface for easy interaction
- Example templates for common use cases

## Getting Started

### Prerequisites

- Go 1.22 or later

### Installation

1. Clone the repository:

```bash
git clone https://github.com/yourusername/htmlgo-convert-webui.git
cd htmlgo-convert-webui
```

2. Install dependencies:

```bash
go mod tidy
```

3. Run the server:

```bash
go run main.go
```

4. Open your browser and navigate to `http://localhost:8080`

## Usage

### Web Interface

The web interface allows you to:

1. Enter HTML code on the left panel
2. Configure the package prefix (default is "h")
3. Convert HTML to Go code by clicking the conversion button
4. View the generated Go code on the right panel

### Example Usage

Here's a basic example of converting HTML to Go code:

HTML:

```html
<div class="container">
  <h1 class="text-xl font-bold">Hello World</h1>
  <p class="text-gray-600">This is an example</p>
</div>
```

Generated Go code (with package prefix "h"):

```go
h.Div(
    h.Attr("class", "container"),
    h.H1(
        h.Attr("class", "text-xl font-bold"),
        h.Text("Hello World"),
    ),
    h.P(
        h.Attr("class", "text-gray-600"),
        h.Text("This is an example"),
    ),
)
```

## API Usage

The server provides a `/convert` endpoint for programmatic access:

```bash
curl -X POST -H "Content-Type: application/json" -d '{
  "html": "<div class=\"container\"><h1>Hello</h1></div>",
  "packagePrefix": "h",
  "direction": "html2go"
}' http://localhost:8080/convert
```

Response:

```json
{
  "code": "h.Div(\n    h.Attr(\"class\", \"container\"),\n    h.H1(\n        h.Text(\"Hello\"),\n    ),\n)"
}
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- [htmlgo](https://github.com/theplant/htmlgo) - The Go HTML builder library
- [html2go](https://github.com/zhangshanwen/html2go) - HTML to Go code converter
