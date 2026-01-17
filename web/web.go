package web

import "embed"

//go:embed static/* templates/* templates/*/*
var EmbedFS embed.FS
