package apidocs

import (
	"embed"
	"io/fs"
	"path"
	"strings"
)

//go:embed *.md admin/*.md
var content embed.FS

type Document struct {
	ID      string `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Documents is embedded in the server, never in the public frontend bundle.
var Documents = loadDocuments("*.md")

// AdminDocuments is only returned by the administrator-protected documentation route.
var AdminDocuments = loadDocuments("admin/*.md")

func loadDocuments(pattern string) []Document {
	names, err := fs.Glob(content, pattern)
	if err != nil {
		panic(err)
	}
	documents := make([]Document, 0, len(names))
	for _, name := range names {
		body, err := content.ReadFile(name)
		if err != nil {
			panic(err)
		}
		title, markdown, _ := strings.Cut(string(body), "\n")
		_, id, _ := strings.Cut(strings.TrimSuffix(strings.TrimSuffix(path.Base(name), ".md"), ".zh_CN"), "-")
		documents = append(documents, Document{
			ID: id, Title: strings.TrimPrefix(title, "# "), Content: strings.TrimSpace(markdown),
		})
	}
	return documents
}
