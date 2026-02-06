package coverage

import (
	"encoding/json"
	"regexp"
	"strings"
)

var (
	// Shopline docs include two API base variants:
	// - Open API: https://open.shopline.io/v1/...
	// - Storefront API: https://{handle}.shoplineapp.com/storefront-api/v1/...
	//
	// We capture the path *after* "/v1".
	reShoplineV1URL = regexp.MustCompile(
		`https://open\.shopline\.io/v1(?P<open_path>/[^\s'")]+)|https://[^\s'")]+/storefront-api/v1(?P<storefront_path>/[^\s'")]+)`,
	)
	reMethodLine = regexp.MustCompile(`(?m)^(get|post|put|patch|del|delete)$`)
)

type openAPIDefinitionSnippet struct {
	Paths map[string]map[string]json.RawMessage `json:"paths"`
}

func parseEndpointFromOpenAPIDefinition(markdown string) (Endpoint, bool) {
	// ReadMe's `.../reference/<slug>.md` pages embed a JSON "OpenAPI definition"
	// code block. Unlike the HTML pages, these do not necessarily contain a
	// literal `https://open.shopline.io/v1/...` URL string.
	//
	// Example:
	//   servers: [{ "url": "https://open.shopline.io/v1" }]
	//   paths: { "/orders": { "get": { ... } } }
	const marker = "OpenAPI definition"
	markerIdx := strings.Index(markdown, marker)
	if markerIdx < 0 {
		return Endpoint{}, false
	}

	fenceIdx := strings.Index(markdown[markerIdx:], "```json")
	if fenceIdx < 0 {
		return Endpoint{}, false
	}
	fenceIdx += markerIdx
	start := fenceIdx + len("```json")

	// Skip whitespace/newlines after ```json
	for start < len(markdown) {
		switch markdown[start] {
		case ' ', '\t', '\r', '\n':
			start++
		default:
			goto findEnd
		}
	}

findEnd:
	endRel := strings.Index(markdown[start:], "```")
	if endRel < 0 {
		return Endpoint{}, false
	}
	end := start + endRel
	rawJSON := strings.TrimSpace(markdown[start:end])
	if rawJSON == "" {
		return Endpoint{}, false
	}

	var snip openAPIDefinitionSnippet
	if err := json.Unmarshal([]byte(rawJSON), &snip); err != nil {
		return Endpoint{}, false
	}
	if len(snip.Paths) != 1 {
		return Endpoint{}, false
	}

	var path string
	var methods map[string]json.RawMessage
	for p, m := range snip.Paths {
		path = p
		methods = m
	}
	if path == "" || len(methods) == 0 {
		return Endpoint{}, false
	}

	// The snippet should represent a single endpoint page, meaning a single HTTP
	// method. Be strict here to avoid mis-attributing coverage.
	known := []string{"get", "post", "put", "patch", "delete", "del"}
	found := ""
	for _, k := range known {
		if _, ok := methods[k]; ok {
			if found != "" {
				return Endpoint{}, false
			}
			found = k
		}
	}
	if found == "" {
		return Endpoint{}, false
	}

	return Endpoint{
		Method: NormalizeMethod(found),
		Path:   NormalizePath(path),
	}, true
}

// ParseEndpointFromDocMarkdown attempts to extract {method, path} from a firecrawl
// markdown scrape of a Shopline API reference page.
//
// This expects the "full" scrape (onlyMainContent=false). Those pages usually contain:
//   - a line with the method ("get", "post", ...)
//   - a line with the full URL ("https://open.shopline.io/v1/...").
func ParseEndpointFromDocMarkdown(markdown string) (Endpoint, bool) {
	if markdown == "" {
		return Endpoint{}, false
	}

	loc := reShoplineV1URL.FindStringSubmatchIndex(markdown)
	if loc == nil {
		// Fallback for official ReadMe plaintext: `/reference/<slug>.md`.
		if ep, ok := parseEndpointFromOpenAPIDefinition(markdown); ok {
			return ep, true
		}
		return Endpoint{}, false
	}

	urlMatch := reShoplineV1URL.FindStringSubmatch(markdown)
	// Full match + 2 capture groups.
	if len(urlMatch) < 3 {
		return Endpoint{}, false
	}
	path := urlMatch[1]
	if path == "" {
		path = urlMatch[2]
	}
	path = NormalizePath(path)

	// Find the nearest method line before the URL.
	prefix := markdown[:loc[0]]
	method := ""
	matches := reMethodLine.FindAllStringIndex(prefix, -1)
	if len(matches) > 0 {
		last := matches[len(matches)-1]
		method = strings.TrimSpace(prefix[last[0]:last[1]])
	}
	if method == "" {
		// Fallback: look around the URL for any method line.
		windowStart := loc[0] - 500
		if windowStart < 0 {
			windowStart = 0
		}
		windowEnd := loc[1] + 500
		if windowEnd > len(markdown) {
			windowEnd = len(markdown)
		}
		win := markdown[windowStart:windowEnd]
		if m := reMethodLine.FindStringSubmatch(win); len(m) == 2 {
			method = m[1]
		}
	}
	if method == "" {
		return Endpoint{}, false
	}

	return Endpoint{
		Method: NormalizeMethod(method),
		Path:   path,
	}, true
}
