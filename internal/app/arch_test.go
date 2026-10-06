package app_test

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

// modulePrefix is this repository's import path for internal packages.
const modulePrefix = "github.com/weavatar/weavatar/internal/"

// escapeFixture holds the escapes TestRequestContextDoesNotEscape reports,
// next to look-alikes it must accept.
const escapeFixture = `package fixture

func handler(c fiber.Ctx) error {
	ctx := c.Context()
	go use(c)
	go func() { use(ctx) }()
	q.Push(func(context.Context) error { return use(c.Context()) })
	q.Push(func(ctx context.Context) error { return use(ctx) })
	use(c, ctx)
	return nil
}

func usecase(ctx context.Context) {
	go use(ctx)
	go func(ctx context.Context) { use(ctx) }(context.Background())
	go func() {
		ctx := context.Background()
		use(ctx)
	}()
}

func job(c *cron.Cron) {
	go use(c)
}
`

// nonModulePackages are the fixed top-level packages under internal/;
// every other directory is a business module, discovered automatically.
var nonModulePackages = map[string]bool{
	"app":        true,
	"migrations": true,
	"mocks":      true,
	"platform":   true,
	"shared":     true,
}

func TestModuleBoundaries(t *testing.T) {
	internalDir := filepath.Join("..", "..", "internal")

	entries, err := os.ReadDir(internalDir)
	if err != nil {
		t.Fatalf("read internal/: %v", err)
	}
	modules := map[string]bool{}
	for _, e := range entries {
		if e.IsDir() && !nonModulePackages[e.Name()] {
			modules[e.Name()] = true
		}
	}
	if len(modules) == 0 {
		t.Fatal("no business modules discovered under internal/")
	}

	fset := token.NewFileSet()
	err = filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}

		rel, err := filepath.Rel(internalDir, path)
		if err != nil {
			return err
		}
		segs := strings.Split(filepath.ToSlash(rel), "/")
		ownerTop := segs[0]
		ownerSub := ""
		if len(segs) > 2 { // internal/<top>/<sub>/file.go
			ownerSub = segs[1]
		}

		f, err := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range f.Imports {
			target := strings.Trim(imp.Path.Value, `"`)
			if !strings.HasPrefix(target, modulePrefix) {
				continue
			}
			if msg := violation(modules, ownerTop, ownerSub, strings.TrimPrefix(target, modulePrefix)); msg != "" {
				t.Errorf("%s imports %s: %s", rel, target, msg)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/: %v", err)
	}
}

// TestRequestContextDoesNotEscape guards service and biz packages: Fiber
// recycles the Ctx after the handler, so goroutines and queued jobs must not
// capture the request or its context.
func TestRequestContextDoesNotEscape(t *testing.T) {
	fset := token.NewFileSet()

	// the detector must catch the fixture's escapes before its silence on the
	// repository means anything
	fixture, err := parser.ParseFile(fset, "fixture.go", escapeFixture, parser.SkipObjectResolution)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	var caught []string
	for _, id := range requestEscapes(fixture) {
		caught = append(caught, fmt.Sprintf("%d:%s", fset.Position(id.Pos()).Line, id.Name))
	}
	if want := []string{"5:c", "6:ctx", "7:c", "14:ctx"}; !slices.Equal(caught, want) {
		t.Fatalf("fixture escapes = %v, want %v", caught, want)
	}

	internalDir := filepath.Join("..", "..", "internal")
	scanned := 0
	err = filepath.WalkDir(internalDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}

		rel, err := filepath.Rel(internalDir, path)
		if err != nil {
			return err
		}
		segs := strings.Split(filepath.ToSlash(rel), "/")
		if len(segs) != 3 || nonModulePackages[segs[0]] || (segs[1] != "service" && segs[1] != "biz") {
			return nil
		}

		f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
		if err != nil {
			return err
		}
		scanned++
		for _, id := range requestEscapes(f) {
			t.Errorf("%s:%d: %s reaches a goroutine or queued job; give background work its own ctx and copied values",
				rel, fset.Position(id.Pos()).Line, id.Name)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk internal/: %v", err)
	}
	if scanned == 0 {
		t.Fatal("no service or biz package found under internal/")
	}
}

// violation returns why this import edge is illegal, or "".
func violation(modules map[string]bool, ownerTop, ownerSub, target string) string {
	targetSegs := strings.Split(target, "/")
	targetTop := targetSegs[0]
	targetSub := ""
	if len(targetSegs) > 1 {
		targetSub = targetSegs[1]
	}

	switch {
	case ownerTop == "app":
		return "" // the assembly imports everything

	case ownerTop == "mocks":
		return "" // generated mirrors of the mocked interfaces

	case ownerTop == "migrations":
		return "migrations declares schema only and imports no internal package"

	case ownerTop == "shared":
		if targetTop == "shared" {
			return ""
		}
		return "shared holds the bottom-layer contracts and cannot depend on layers above"

	case ownerTop == "platform":
		switch {
		case ownerSub == "conf":
			return "platform/conf is the bottom layer and imports no internal package"
		case targetTop == "shared":
			return ""
		case targetTop == "platform" && (targetSub == ownerSub || targetSub == "conf"):
			return "" // own subtree, or the configuration everything reads
		}
		return "platform assembles infrastructure and must not know business modules"

	case modules[ownerTop]:
		switch {
		case targetTop == ownerTop:
			if ownerSub == "biz" && targetSub != "biz" {
				return "biz is the core and cannot import its own data/service adapters"
			}
			return ""
		case targetTop == "shared":
			return ""
		case modules[targetTop]:
			if targetSub == "biz" {
				return ""
			}
			return "modules reach each other only through the other module's biz package"
		default:
			return "modules depend on shared contracts and other modules' biz packages only"
		}
	}
	return ""
}

// requestEscapes returns the request-scoped identifiers in f that a go
// statement or a closure handed to Push references.
func requestEscapes(f *ast.File) []*ast.Ident {
	var found []*ast.Ident
	for _, decl := range f.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			if fn.Body != nil {
				walkEscapes(fn.Body, requestParams(nil, fn.Type, false), false, &found)
			}
			continue
		}
		walkEscapes(decl, map[string]bool{}, false, &found)
	}
	return found
}

// walkEscapes visits n in source order. names holds the identifiers bound to
// request-scoped values; inside a goroutine or a queued job (spawned), every
// reference to one of them is an escape.
func walkEscapes(n ast.Node, names map[string]bool, spawned bool, found *[]*ast.Ident) {
	ast.Inspect(n, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncLit:
			walkEscapes(n.Body, requestParams(names, n.Type, spawned), spawned, found)
			return false
		case *ast.GoStmt:
			walkEscapes(n.Call, names, true, found)
			return false
		case *ast.CallExpr:
			if !isPush(n.Fun) {
				return true
			}
			walkEscapes(n.Fun, names, spawned, found)
			for _, arg := range n.Args {
				walkEscapes(arg, names, true, found)
			}
			return false
		case *ast.AssignStmt:
			for _, rhs := range n.Rhs {
				walkEscapes(rhs, names, spawned, found)
			}
			for i, lhs := range n.Lhs {
				id, ok := lhs.(*ast.Ident)
				switch {
				case !ok:
					walkEscapes(lhs, names, spawned, found)
				case !spawned && len(n.Lhs) == len(n.Rhs) && isRequestContext(n.Rhs[i], names):
					names[id.Name] = true
				case n.Tok == token.DEFINE:
					delete(names, id.Name)
				}
			}
			return false
		case *ast.ValueSpec:
			for _, value := range n.Values {
				walkEscapes(value, names, spawned, found)
			}
			for _, id := range n.Names {
				delete(names, id.Name)
			}
			return false
		case *ast.Field:
			return false // a parameter or struct field declaration, not a use
		case *ast.SelectorExpr:
			walkEscapes(n.X, names, spawned, found) // Sel names a field or method
			return false
		case *ast.KeyValueExpr:
			if _, ok := n.Key.(*ast.Ident); ok {
				walkEscapes(n.Value, names, spawned, found) // the key is a field name
				return false
			}
		case *ast.Ident:
			if spawned && names[n.Name] {
				*found = append(*found, n)
			}
		}
		return true
	})
}

// requestParams derives a function's names from the enclosing ones: its
// parameters shadow them, and outside a spawned closure a fiber.Ctx or
// context.Context parameter is request-scoped itself.
func requestParams(outer map[string]bool, typ *ast.FuncType, spawned bool) map[string]bool {
	names := map[string]bool{}
	maps.Copy(names, outer)
	for _, field := range typ.Params.List {
		request := !spawned && (isSelector(field.Type, "fiber", "Ctx") || isSelector(field.Type, "context", "Context"))
		for _, id := range field.Names {
			if request {
				names[id.Name] = true
			} else {
				delete(names, id.Name)
			}
		}
	}
	return names
}

// isRequestContext reports whether expr is X.Context() on a request-scoped X.
func isRequestContext(expr ast.Expr, names map[string]bool) bool {
	call, ok := expr.(*ast.CallExpr)
	if !ok || len(call.Args) > 0 {
		return false
	}
	sel, ok := call.Fun.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != "Context" {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	return ok && names[x.Name]
}

func isPush(fun ast.Expr) bool {
	switch fun := fun.(type) {
	case *ast.SelectorExpr:
		return fun.Sel.Name == "Push"
	case *ast.Ident:
		return fun.Name == "Push"
	default:
		return false
	}
}

func isSelector(expr ast.Expr, pkg, name string) bool {
	sel, ok := expr.(*ast.SelectorExpr)
	if !ok || sel.Sel.Name != name {
		return false
	}
	x, ok := sel.X.(*ast.Ident)
	return ok && x.Name == pkg
}
