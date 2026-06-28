// Package cmd — chrome_bookmarks.go: export a profile's Bookmarks
// file to md|html|json. Defaults to md on stdout.
package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type bookmarkItem struct {
	Title    string
	URL      string
	Folder   string
	Children []bookmarkItem
}

func runChromeExportBookmarks(args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "chrome export-bookmarks: ERROR usage: gitmap chrome export-bookmarks <profile> [--format md|html|json] [--out <file>]")
		os.Exit(2)
	}
	profile, ok := resolveChromeProfile(args[0])
	if !ok {
		fmt.Fprintf(os.Stderr, "chrome export-bookmarks: ERROR profile %q not found\n", args[0])
		printAvailableChromeProfilesWithDisplay()
		os.Exit(1)
	}
	format, outPath := "md", ""
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--format", "-f":
			if i+1 < len(args) {
				format = args[i+1]
				i++
			}
		case "--out", "-o":
			if i+1 < len(args) {
				outPath = args[i+1]
				i++
			}
		}
	}
	roots := loadBookmarkRoots(profile.Path)
	body, err := renderBookmarks(roots, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "chrome export-bookmarks: ERROR %v\n", err)
		os.Exit(1)
	}
	if outPath == "" {
		fmt.Print(body)
		return
	}
	if err := os.WriteFile(outPath, []byte(body), 0o644); err != nil { //nolint:gosec
		fmt.Fprintf(os.Stderr, "chrome export-bookmarks: ERROR write: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("\033[1;92m✓ wrote\033[0m %s (%d bytes)\n", outPath, len(body))
}

func loadBookmarkRoots(profile string) []bookmarkItem {
	raw, err := os.ReadFile(filepath.Join(profile, "Bookmarks")) //nolint:gosec
	if err != nil {
		return nil
	}
	var doc struct {
		Roots map[string]json.RawMessage `json:"roots"`
	}
	if json.Unmarshal(raw, &doc) != nil {
		return nil
	}
	out := []bookmarkItem{}
	for name, r := range doc.Roots {
		item := parseBookmarkNode(r)
		item.Folder = name
		out = append(out, item)
	}
	return out
}

func parseBookmarkNode(raw json.RawMessage) bookmarkItem {
	var n struct {
		Name     string            `json:"name"`
		URL      string            `json:"url"`
		Children []json.RawMessage `json:"children"`
	}
	if json.Unmarshal(raw, &n) != nil {
		return bookmarkItem{}
	}
	item := bookmarkItem{Title: n.Name, URL: n.URL}
	for _, c := range n.Children {
		item.Children = append(item.Children, parseBookmarkNode(c))
	}
	return item
}

func renderBookmarks(roots []bookmarkItem, format string) (string, error) {
	switch format {
	case "json":
		b, err := json.MarshalIndent(roots, "", "  ")
		return string(b) + "\n", err
	case "html":
		var sb strings.Builder
		sb.WriteString("<!DOCTYPE html><html><body>\n")
		for _, r := range roots {
			renderBookmarkHTML(&sb, r, 0)
		}
		sb.WriteString("</body></html>\n")
		return sb.String(), nil
	default: // md
		var sb strings.Builder
		for _, r := range roots {
			renderBookmarkMD(&sb, r, 0)
		}
		return sb.String(), nil
	}
}

func renderBookmarkMD(sb *strings.Builder, n bookmarkItem, depth int) {
	indent := strings.Repeat("  ", depth)
	switch {
	case n.URL != "":
		fmt.Fprintf(sb, "%s- [%s](%s)\n", indent, fallback(n.Title, n.URL), n.URL)
	case n.Title != "" || n.Folder != "":
		fmt.Fprintf(sb, "%s- **%s/**\n", indent, fallback(n.Title, n.Folder))
	}
	for _, c := range n.Children {
		renderBookmarkMD(sb, c, depth+1)
	}
}

func renderBookmarkHTML(sb *strings.Builder, n bookmarkItem, depth int) {
	if n.URL != "" {
		fmt.Fprintf(sb, "<dt><a href=\"%s\">%s</a></dt>\n", n.URL, fallback(n.Title, n.URL))
		return
	}
	fmt.Fprintf(sb, "<dt><h3>%s</h3><dl>\n", fallback(n.Title, n.Folder))
	for _, c := range n.Children {
		renderBookmarkHTML(sb, c, depth+1)
	}
	sb.WriteString("</dl></dt>\n")
}

func fallback(a, b string) string {
	if a != "" {
		return a
	}
	return b
}
