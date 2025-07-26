package reddit_v2

import (
	"embed"
)

//go:embed static/*
var StaticFiles embed.FS

//go:embed static/html/index.html
var IndexHTML []byte
