package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/salmonumbrella/shopline-cli/internal/coverage"
)

type firecrawlPage struct {
	Markdown string `json:"markdown"`
	Metadata struct {
		SourceURL  string `json:"sourceURL"`
		StatusCode int    `json:"statusCode"`
	} `json:"metadata"`
	Warning string `json:"warning"`
}

func main() {
	var (
		docsDir = flag.String("docs-pages", "docs/shopline-openapi/pages_md/reference", "directory containing Shopline /reference/*.md pages (or firecrawl .json)")
		outDir  = flag.String("out", "docs/coverage", "output directory for coverage artifacts")
	)
	flag.Parse()

	docEndpoints, unparsed, err := loadDocEndpoints(*docsDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	codeEndpoints, err := coverage.ParseEndpointsFromGoFiles([]string{"internal/api/*.go"})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	// De-duplicate by {method,path} while keeping a stable metadata winner.
	docEndpoints = uniqByKey(docEndpoints)
	codeEndpoints = uniqByKey(codeEndpoints)

	rep := coverage.BuildReport(docEndpoints, codeEndpoints, unparsed)

	if err := os.MkdirAll(*outDir, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	if err := writeJSON(filepath.Join(*outDir, "openapi_endpoints.json"), rep.DocEndpoints); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := writeJSON(filepath.Join(*outDir, "code_endpoints.json"), rep.CodeEndpoints); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.WriteFile(filepath.Join(*outDir, "report.md"), []byte(rep.RenderMarkdown()), 0o644); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}

	fmt.Printf("wrote %s\n", filepath.Join(*outDir, "report.md"))
}

func loadDocEndpoints(root string) ([]coverage.Endpoint, []string, error) {
	var eps []coverage.Endpoint
	var unparsed []string

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}

		switch {
		case strings.HasSuffix(path, ".md"):
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			ep, ok := coverage.ParseEndpointFromDocMarkdown(string(raw))
			if !ok {
				unparsed = append(unparsed, path)
				return nil
			}
			ep.DocURL = inferDocURLFromPath(path)
			ep.Source = path
			eps = append(eps, ep)
			return nil

		case strings.HasSuffix(path, ".json"):
			raw, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			var page firecrawlPage
			if err := json.Unmarshal(raw, &page); err != nil {
				return fmt.Errorf("parse %s: %w", path, err)
			}

			ep, ok := coverage.ParseEndpointFromDocMarkdown(page.Markdown)
			if !ok {
				unparsed = append(unparsed, path)
				return nil
			}
			ep.DocURL = page.Metadata.SourceURL
			ep.Source = path
			eps = append(eps, ep)
			return nil
		default:
			return nil
		}
	})
	if err != nil {
		return nil, nil, err
	}

	sort.Slice(eps, func(i, j int) bool { return eps[i].Key() < eps[j].Key() })
	sort.Strings(unparsed)
	return eps, unparsed, nil
}

func inferDocURLFromPath(path string) string {
	p := filepath.ToSlash(path)
	i := strings.Index(p, "/reference/")
	if i < 0 {
		return ""
	}
	rel := p[i+1:] // "reference/..."
	rel = strings.TrimSuffix(rel, ".md")
	return "https://open-api.docs.shoplineapp.com/" + rel
}

func uniqByKey(in []coverage.Endpoint) []coverage.Endpoint {
	m := make(map[string]coverage.Endpoint, len(in))
	for _, e := range in {
		k := coverage.NormalizeMethod(e.Method) + " " + coverage.NormalizePath(e.Path)
		e.Method = coverage.NormalizeMethod(e.Method)
		e.Path = coverage.NormalizePath(e.Path)
		if _, ok := m[k]; !ok {
			m[k] = e
		}
	}
	out := make([]coverage.Endpoint, 0, len(m))
	for _, e := range m {
		out = append(out, e)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key() < out[j].Key() })
	return out
}

func writeJSON(path string, v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}
