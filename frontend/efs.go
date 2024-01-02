package frontend

import "embed"

//go:embed "dist/*"
var Assets embed.FS

//go:embed "html/*"
var Templates embed.FS
