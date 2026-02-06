package coverage

import (
	"sort"
	"strings"
)

type CoverageReport struct {
	DocEndpoints  []Endpoint
	CodeEndpoints []Endpoint

	MissingInCode []Endpoint
	ExtraInCode   []Endpoint

	UnparsedDocPages []string
}

func BuildReport(docEndpoints []Endpoint, codeEndpoints []Endpoint, unparsedDocPages []string) CoverageReport {
	// Normalize to {method,path} keys
	docMap := make(map[string]Endpoint, len(docEndpoints))
	for _, e := range docEndpoints {
		k := NormalizeMethod(e.Method) + " " + NormalizePath(e.Path)
		e.Method = NormalizeMethod(e.Method)
		e.Path = NormalizePath(e.Path)
		docMap[k] = e
	}

	codeMap := make(map[string]Endpoint, len(codeEndpoints))
	for _, e := range codeEndpoints {
		k := NormalizeMethod(e.Method) + " " + NormalizePath(e.Path)
		e.Method = NormalizeMethod(e.Method)
		e.Path = NormalizePath(e.Path)
		codeMap[k] = e
	}

	var missing []Endpoint
	for k, e := range docMap {
		if _, ok := codeMap[k]; !ok {
			missing = append(missing, e)
		}
	}
	var extra []Endpoint
	for k, e := range codeMap {
		if _, ok := docMap[k]; !ok {
			extra = append(extra, e)
		}
	}

	sortEndpoints := func(es []Endpoint) {
		sort.Slice(es, func(i, j int) bool {
			if es[i].Path == es[j].Path {
				return es[i].Method < es[j].Method
			}
			return es[i].Path < es[j].Path
		})
	}
	sortEndpoints(docEndpoints)
	sortEndpoints(codeEndpoints)
	sortEndpoints(missing)
	sortEndpoints(extra)
	sort.Strings(unparsedDocPages)

	return CoverageReport{
		DocEndpoints:     docEndpoints,
		CodeEndpoints:    codeEndpoints,
		MissingInCode:    missing,
		ExtraInCode:      extra,
		UnparsedDocPages: unparsedDocPages,
	}
}

func (r CoverageReport) RenderMarkdown() string {
	var b strings.Builder
	b.WriteString("# Shopline OpenAPI Coverage Report\n\n")

	b.WriteString("## Summary\n\n")
	b.WriteString("- Docs endpoints parsed: ")
	b.WriteString(itoa(len(r.DocEndpoints)))
	b.WriteString("\n")
	b.WriteString("- Code endpoints detected: ")
	b.WriteString(itoa(len(r.CodeEndpoints)))
	b.WriteString("\n")
	b.WriteString("- Missing in code: ")
	b.WriteString(itoa(len(r.MissingInCode)))
	b.WriteString("\n")
	b.WriteString("- Extra in code: ")
	b.WriteString(itoa(len(r.ExtraInCode)))
	b.WriteString("\n")
	b.WriteString("- Unparsed doc pages: ")
	b.WriteString(itoa(len(r.UnparsedDocPages)))
	b.WriteString("\n\n")

	if len(r.UnparsedDocPages) > 0 {
		b.WriteString("## Unparsed Doc Pages\n\n")
		for _, p := range r.UnparsedDocPages {
			b.WriteString("- ")
			b.WriteString(p)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(r.MissingInCode) > 0 {
		b.WriteString("## Missing In Code\n\n")
		for _, e := range r.MissingInCode {
			b.WriteString("- `")
			b.WriteString(e.Key())
			b.WriteString("`")
			if e.DocURL != "" {
				b.WriteString(" (")
				b.WriteString(e.DocURL)
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if len(r.ExtraInCode) > 0 {
		b.WriteString("## Extra In Code\n\n")
		for _, e := range r.ExtraInCode {
			b.WriteString("- `")
			b.WriteString(e.Key())
			b.WriteString("`")
			if e.Source != "" {
				b.WriteString(" (")
				b.WriteString(e.Source)
				b.WriteString(")")
			}
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	return b.String()
}

func itoa(n int) string {
	// avoid pulling strconv in a file that's mostly strings
	if n == 0 {
		return "0"
	}
	var buf [32]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	return string(buf[i:])
}
