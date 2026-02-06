package coverage

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strconv"
	"strings"
)

// ParseEndpointsFromGoFiles extracts method+path pairs by statically inspecting
// internal/api code. It looks for calls like:
//
//	c.Get(ctx, "/orders", &out)
//	c.Get(ctx, fmt.Sprintf("/orders/%s", id), &out)
func ParseEndpointsFromGoFiles(globs []string) ([]Endpoint, error) {
	var files []string
	for _, g := range globs {
		m, err := filepath.Glob(g)
		if err != nil {
			return nil, err
		}
		files = append(files, m...)
	}

	var out []Endpoint
	fset := token.NewFileSet()
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil, fmt.Errorf("parse %s: %w", path, err)
		}

		for _, decl := range f.Decls {
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}

			pathVars := map[string]string{}
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch v := n.(type) {
				case *ast.AssignStmt:
					for i := 0; i < len(v.Lhs) && i < len(v.Rhs); i++ {
						lhs, ok := v.Lhs[i].(*ast.Ident)
						if !ok {
							continue
						}
						if tmpl, ok := evalPathTemplate(v.Rhs[i]); ok && strings.HasPrefix(tmpl, "/") {
							pathVars[lhs.Name] = tmpl
						}
					}
				case *ast.ValueSpec:
					for i, name := range v.Names {
						if name == nil || i >= len(v.Values) {
							continue
						}
						if tmpl, ok := evalPathTemplate(v.Values[i]); ok && strings.HasPrefix(tmpl, "/") {
							pathVars[name.Name] = tmpl
						}
					}
				case *ast.CallExpr:
					sel, ok := v.Fun.(*ast.SelectorExpr)
					if !ok {
						return true
					}
					name := sel.Sel.Name
					var method string
					switch name {
					case "Get":
						method = "GET"
					case "Post":
						method = "POST"
					case "Put":
						method = "PUT"
					case "Patch":
						method = "PATCH"
					case "Delete":
						method = "DELETE"
					case "DeleteWithBody":
						method = "DELETE"
					default:
						return true
					}
					// Signature: (ctx, path, ...)
					if len(v.Args) < 2 {
						return true
					}

					pathTmpl, ok := evalPathTemplate(v.Args[1])
					if !ok {
						ident, isIdent := v.Args[1].(*ast.Ident)
						if isIdent {
							if s, found := pathVars[ident.Name]; found {
								pathTmpl = s
								ok = true
							}
						}
					}
					if !ok || pathTmpl == "" || !strings.HasPrefix(pathTmpl, "/") {
						return true
					}

					out = append(out, Endpoint{
						Method: method,
						Path:   NormalizePath(pathTmpl),
						Source: path,
					})
				}
				return true
			})
		}
	}

	return out, nil
}

func evalPathTemplate(e ast.Expr) (string, bool) {
	switch v := e.(type) {
	case *ast.BasicLit:
		if v.Kind != token.STRING {
			return "", false
		}
		s, err := strconv.Unquote(v.Value)
		if err != nil {
			return "", false
		}
		return s, true
	case *ast.ParenExpr:
		return evalPathTemplate(v.X)
	case *ast.BinaryExpr:
		if v.Op != token.ADD {
			return "", false
		}
		l, ok := evalPathTemplate(v.X)
		r, ok2 := evalPathTemplate(v.Y)
		// If we have a literal prefix but can't evaluate the suffix (usually query builder),
		// keep the prefix.
		if ok && !ok2 {
			return l, true
		}
		if !ok && ok2 {
			return r, true
		}
		if !ok || !ok2 {
			return "", false
		}
		return l + r, true
	case *ast.CallExpr:
		// fmt.Sprintf("...", ...)
		if isFmtSprintf(v.Fun) && len(v.Args) > 0 {
			format, ok := evalPathTemplate(v.Args[0])
			if !ok {
				return "", false
			}
			return sprintfToTemplate(format), true
		}
		return "", false
	default:
		return "", false
	}
}

func isFmtSprintf(fun ast.Expr) bool {
	switch f := fun.(type) {
	case *ast.SelectorExpr:
		pkg, ok := f.X.(*ast.Ident)
		return ok && pkg.Name == "fmt" && f.Sel.Name == "Sprintf"
	case *ast.Ident:
		return f.Name == "Sprintf"
	default:
		return false
	}
}

// sprintfToTemplate converts a fmt.Sprintf path format string into a stable template.
// For example: "/orders/%s/items/%d" -> "/orders/{}/items/{}"
func sprintfToTemplate(format string) string {
	var b strings.Builder
	b.Grow(len(format))
	for i := 0; i < len(format); i++ {
		ch := format[i]
		if ch != '%' {
			b.WriteByte(ch)
			continue
		}
		// Escaped percent.
		if i+1 < len(format) && format[i+1] == '%' {
			b.WriteByte('%')
			i++
			continue
		}

		// Minimal parser: consume the verb letter if present.
		// Common cases for paths are %s, %d, %v.
		if i+1 < len(format) {
			verb := format[i+1]
			switch verb {
			case 's', 'd', 'v', 't':
				b.WriteString("{}")
				i++
				continue
			}
		}

		// Unknown formatting: keep '%' as-is so we don't silently corrupt.
		b.WriteByte(ch)
	}
	return b.String()
}
