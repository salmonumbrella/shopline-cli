package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type report struct {
	TotalMethods int      `json:"total_methods"`
	Covered      []string `json:"covered"`
	Missing      []string `json:"missing"`
}

func main() {
	var outMD string
	var outJSON string
	flag.StringVar(&outMD, "out-md", "docs/coverage/cli_api_report.md", "Path to write markdown report")
	flag.StringVar(&outJSON, "out-json", "docs/coverage/cli_api_report.json", "Path to write JSON report")
	flag.Parse()

	methods, err := parseAPIClientMethods("internal/api/interface.go")
	if err != nil {
		fatal(err)
	}
	methods = filterInternalMethods(methods)

	cmdRefs, err := scanCmdForMethodRefs("internal/cmd", methods)
	if err != nil {
		fatal(err)
	}

	var covered, missing []string
	for _, m := range methods {
		if cmdRefs[m] {
			covered = append(covered, m)
		} else {
			missing = append(missing, m)
		}
	}
	sort.Strings(covered)
	sort.Strings(missing)

	r := report{
		TotalMethods: len(methods),
		Covered:      covered,
		Missing:      missing,
	}

	if err := os.MkdirAll(filepath.Dir(outMD), 0o755); err != nil {
		fatal(err)
	}
	if err := writeMD(outMD, r); err != nil {
		fatal(err)
	}
	if err := writeJSON(outJSON, r); err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	_, _ = fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}

func parseAPIClientMethods(path string) ([]string, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		return nil, err
	}

	var methods []string
	for _, decl := range f.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.TYPE {
			continue
		}
		for _, spec := range gen.Specs {
			ts, ok := spec.(*ast.TypeSpec)
			if !ok || ts.Name == nil || ts.Name.Name != "APIClient" {
				continue
			}
			iface, ok := ts.Type.(*ast.InterfaceType)
			if !ok || iface.Methods == nil {
				continue
			}
			for _, m := range iface.Methods.List {
				// Interface embedding, skip.
				if len(m.Names) == 0 {
					continue
				}
				for _, name := range m.Names {
					if name == nil || name.Name == "" {
						continue
					}
					methods = append(methods, name.Name)
				}
			}
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no APIClient methods found in %s", path)
	}
	sort.Strings(methods)
	return methods, nil
}

func filterInternalMethods(methods []string) []string {
	ignore := map[string]struct{}{
		"Post":           {},
		"Put":            {},
		"DeleteWithBody": {},
	}
	out := make([]string, 0, len(methods))
	for _, m := range methods {
		if _, ok := ignore[m]; ok {
			continue
		}
		out = append(out, m)
	}
	return out
}

func scanCmdForMethodRefs(cmdDir string, methods []string) (map[string]bool, error) {
	found := make(map[string]bool, len(methods))
	for _, m := range methods {
		found[m] = false
	}

	err := filepath.WalkDir(cmdDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		s := string(b)
		for _, m := range methods {
			if found[m] {
				continue
			}
			// Cheap heuristic: ".MethodName(" somewhere in the file.
			if strings.Contains(s, "."+m+"(") {
				found[m] = true
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

func writeJSON(path string, r report) error {
	b, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return err
	}
	b = append(b, '\n')
	return os.WriteFile(path, b, 0o644)
}

func writeMD(path string, r report) error {
	var sb strings.Builder
	sb.WriteString("# Shopline CLI API Coverage (APIClient -> internal/cmd)\n\n")
	sb.WriteString(fmt.Sprintf("- APIClient methods: %d\n", r.TotalMethods))
	sb.WriteString(fmt.Sprintf("- Covered by internal/cmd: %d\n", len(r.Covered)))
	sb.WriteString(fmt.Sprintf("- Missing in internal/cmd: %d\n\n", len(r.Missing)))

	if len(r.Missing) > 0 {
		sb.WriteString("## Missing\n\n")
		for _, m := range r.Missing {
			sb.WriteString("- `")
			sb.WriteString(m)
			sb.WriteString("`\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Covered\n\n")
	for _, m := range r.Covered {
		sb.WriteString("- `")
		sb.WriteString(m)
		sb.WriteString("`\n")
	}
	sb.WriteString("\n")

	return os.WriteFile(path, []byte(sb.String()), 0o644)
}
