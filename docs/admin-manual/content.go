package adminmanual

import (
	"embed"
	"strings"
)

// Content is served only by the authenticated administrator-manual endpoints.
// It must never be imported by the customer documentation or frontend build.
//
//go:embed *.md
var content embed.FS

type Document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

var Documents []Document

func init() {
	entries, err := content.ReadDir(".")
	if err != nil {
		panic(err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".zh_CN.md") {
			continue
		}
		body, err := content.ReadFile(name)
		if err != nil {
			panic(err)
		}
		title, markdown, _ := strings.Cut(string(body), "\n")
		Documents = append(Documents, Document{
			ID: strings.TrimSuffix(name[3:], ".zh_CN.md"), Title: strings.TrimPrefix(title, "# "), Content: strings.TrimSpace(markdown),
		})
	}
}
