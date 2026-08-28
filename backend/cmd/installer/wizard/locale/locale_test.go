package locale

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
	"unicode"
)

var allowedEnglishConstants = map[string]struct{}{
	"CheckDockerAPI":                   {},
	"CheckDockerCompose":               {},
	"EULAFormName":                     {},
	"LLMProviderOpenAI":                {},
	"LLMProviderAnthropic":             {},
	"LLMProviderGemini":                {},
	"LLMProviderBedrock":               {},
	"LLMProviderOllama":                {},
	"LLMProviderDeepSeek":              {},
	"LLMProviderGLM":                   {},
	"LLMProviderKimi":                  {},
	"LLMProviderQwen":                  {},
	"MonitoringLangfuseFormName":       {},
	"ToolsSearchEnginesDuckDuckGoName": {},
	"ToolsSearchEnginesSploitusName":   {},
	"ToolsSearchEnginesPerplexityName": {},
	"ToolsSearchEnginesTavilyName":     {},
	"ToolsSearchEnginesTraversaalName": {},
	"ToolsSearchEnginesSearxngName":    {},
	"EmbedderProviderOpenAI":           {},
	"EmbedderProviderOllama":           {},
	"EmbedderProviderMistral":          {},
	"EmbedderProviderJina":             {},
	"EmbedderProviderHuggingFace":      {},
	"EmbedderProviderGoogleAI":         {},
	"EmbedderProviderVoyageAI":         {},
	"ProcessorComponentPentagi":        {},
	"ProcessorComponentLangfuse":       {},
}

func TestLocaleHasNoUnreviewedEnglishValues(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "locale.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	for _, decl := range file.Decls {
		gen, ok := decl.(*ast.GenDecl)
		if !ok || gen.Tok != token.CONST {
			continue
		}

		for _, spec := range gen.Specs {
			values, ok := spec.(*ast.ValueSpec)
			if !ok {
				continue
			}

			for i, expr := range values.Values {
				if i >= len(values.Names) {
					continue
				}
				name := values.Names[i].Name
				text, ok := stringLiteral(expr)
				if !ok || !containsEnglishWord(text) || containsHan(text) || isAllowedEnglishConstant(name) {
					continue
				}

				t.Errorf("%s: constant %s contains unreviewed English UI text %q", fset.Position(expr.Pos()), name, compact(text))
			}
		}
	}
}

func TestModelsHaveNoHardCodedEnglishUIStrings(t *testing.T) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, "../models", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	for _, pkg := range packages {
		for filename, file := range pkg.Files {
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok {
					return true
				}

				name := callName(call.Fun)
				var expressions []ast.Expr
				requireWhitespace := false
				switch {
				case name == "Render":
					expressions = call.Args
				case name == "append" && len(call.Args) > 1 && isIdentifier(call.Args[0], "sections"):
					expressions = call.Args[1:]
				case strings.HasPrefix(name, "create") && strings.HasSuffix(name, "Field"):
					expressions = call.Args
					requireWhitespace = true // Single-word arguments here are internal field keys.
				default:
					return true
				}

				for _, expr := range expressions {
					ast.Inspect(expr, func(child ast.Node) bool {
						literal, ok := child.(*ast.BasicLit)
						if !ok {
							return true
						}
						text, ok := stringLiteral(literal)
						if !ok || !containsEnglishWord(text) || containsHan(text) ||
							(requireWhitespace && !strings.ContainsAny(text, " \t\r\n")) {
							return true
						}
						t.Errorf("%s: hard-coded English UI text %q; move it to locale", fset.Position(literal.Pos()), compact(text))
						return true
					})
				}

				return true
			})
		}
	}
}

func TestInstallerRuntimeHasNoHardCodedEnglishUIStrings(t *testing.T) {
	assertFileHasNoEnglishSentences(t, "../../main.go")
	assertFileHasNoEnglishSentences(t, "../../processor/locale.go")
}

func TestProcessorHasNoHardCodedEnglishUIStrings(t *testing.T) {
	fset := token.NewFileSet()
	packages, err := parser.ParseDir(fset, "../../processor", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	for _, pkg := range packages {
		for filename, file := range pkg.Files {
			if strings.HasSuffix(filename, "_test.go") {
				continue
			}
			ast.Inspect(file, func(node ast.Node) bool {
				call, ok := node.(*ast.CallExpr)
				if !ok || callName(call.Fun) != "appendLog" {
					return true
				}
				for _, expr := range call.Args {
					checkExpressionForEnglish(t, fset, expr)
				}
				return true
			})
		}
	}
}

func assertFileHasNoEnglishSentences(t *testing.T, path string) {
	t.Helper()
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if !ok {
			return true
		}
		checkEnglishLiteral(t, fset, literal)
		return true
	})
}

func checkExpressionForEnglish(t *testing.T, fset *token.FileSet, expr ast.Expr) {
	t.Helper()
	ast.Inspect(expr, func(node ast.Node) bool {
		literal, ok := node.(*ast.BasicLit)
		if !ok {
			return true
		}
		checkEnglishLiteral(t, fset, literal)
		return true
	})
}

func checkEnglishLiteral(t *testing.T, fset *token.FileSet, literal *ast.BasicLit) {
	t.Helper()
	text, ok := stringLiteral(literal)
	if !ok || !containsEnglishWord(text) || containsHan(text) || !strings.ContainsAny(text, " \t\r\n") {
		return
	}
	t.Errorf("%s: unreviewed English UI text %q", fset.Position(literal.Pos()), compact(text))
}

func isAllowedEnglishConstant(name string) bool {
	if _, ok := allowedEnglishConstants[name]; ok {
		return true
	}
	return strings.HasPrefix(name, "EmbedderURLPlaceholder") ||
		strings.HasPrefix(name, "EmbedderModelPlaceholder") ||
		strings.HasPrefix(name, "EmbedderProviderID")
}

func stringLiteral(expr ast.Expr) (string, bool) {
	literal, ok := expr.(*ast.BasicLit)
	if !ok || literal.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(literal.Value)
	return value, err == nil
}

func callName(expr ast.Expr) string {
	switch value := expr.(type) {
	case *ast.Ident:
		return value.Name
	case *ast.SelectorExpr:
		return value.Sel.Name
	default:
		return ""
	}
}

func isIdentifier(expr ast.Expr, name string) bool {
	identifier, ok := expr.(*ast.Ident)
	return ok && identifier.Name == name
}

func containsEnglishWord(text string) bool {
	letters := 0
	for _, r := range text {
		if unicode.IsLetter(r) && r <= unicode.MaxASCII {
			letters++
			if letters >= 2 {
				return true
			}
		} else {
			letters = 0
		}
	}
	return false
}

func containsHan(text string) bool {
	for _, r := range text {
		if unicode.Is(unicode.Han, r) {
			return true
		}
	}
	return false
}

func compact(text string) string {
	return strings.Join(strings.Fields(text), " ")
}
