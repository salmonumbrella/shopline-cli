package coverage

import (
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
