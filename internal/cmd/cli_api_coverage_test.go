package cmd

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
)

func TestCLIApiCoverage(t *testing.T) {
	methods, err := parseAPIClientMethodsForTest(filepath.Join("..", "api", "interface.go"))
	if err != nil {
		t.Fatalf("parse APIClient methods: %v", err)
	}
	methods = filterInternalMethodsForTest(methods)

	cmdRefs, err := scanCmdForMethodRefsForTest(".", methods)
	if err != nil {
		t.Fatalf("scan internal/cmd for method refs: %v", err)
	}

	var missing []string
	for _, m := range methods {
		if !cmdRefs[m] {
			missing = append(missing, m)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 {
		t.Fatalf("CLI is missing coverage for %d APIClient methods:\n- %s", len(missing), strings.Join(missing, "\n- "))
	}
}

func parseAPIClientMethodsForTest(path string) ([]string, error) {
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

	sort.Strings(methods)
	return methods, nil
}

func filterInternalMethodsForTest(methods []string) []string {
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

func scanCmdForMethodRefsForTest(cmdDir string, methods []string) (map[string]bool, error) {
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
