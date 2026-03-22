package templates

import "embed"

// FS holds the embedded template files. The server imports this package
// so the binary is fully self-contained with no runtime filesystem dependency.
//
//go:embed index.html
var FS embed.FS
