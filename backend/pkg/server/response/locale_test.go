package response

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"testing"
)

func TestZhCNErrorMessagesCoverCatalog(t *testing.T) {
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, "errors.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}

	codes := make(map[string]struct{})
	ast.Inspect(file, func(node ast.Node) bool {
		call, ok := node.(*ast.CallExpr)
		if !ok || len(call.Args) < 2 {
			return true
		}

		function, ok := call.Fun.(*ast.Ident)
		if !ok || function.Name != "NewHttpError" {
			return true
		}

		literal, ok := call.Args[1].(*ast.BasicLit)
		if !ok || literal.Kind != token.STRING {
			t.Errorf("%s: error code must be a string literal", fset.Position(call.Args[1].Pos()))
			return true
		}

		code, unquoteErr := strconv.Unquote(literal.Value)
		if unquoteErr != nil {
			t.Errorf("%s: invalid error code: %v", fset.Position(literal.Pos()), unquoteErr)
			return true
		}

		codes[code] = struct{}{}
		if message := zhCNErrorMessages[code]; message == "" {
			t.Errorf("%s: missing Simplified Chinese message for %q", fset.Position(literal.Pos()), code)
		}
		return true
	})

	for code := range zhCNErrorMessages {
		if _, ok := codes[code]; !ok {
			t.Errorf("Simplified Chinese message has no matching error code: %q", code)
		}
	}
}

func TestPreferredResponseLanguage(t *testing.T) {
	tests := []struct {
		name           string
		acceptLanguage string
		want           string
	}{
		{name: "default", want: responseLanguageEnglish},
		{name: "simplified Chinese", acceptLanguage: "zh-CN", want: responseLanguageChinese},
		{name: "Chinese variant", acceptLanguage: "zh-Hans-SG, en;q=0.8", want: responseLanguageChinese},
		{name: "English preferred", acceptLanguage: "en-US, zh-CN;q=0.8", want: responseLanguageEnglish},
		{name: "weighted Chinese", acceptLanguage: "en;q=0.5, zh-CN;q=0.9", want: responseLanguageChinese},
		{name: "invalid", acceptLanguage: "not a language", want: responseLanguageEnglish},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := preferredResponseLanguage(test.acceptLanguage); got != test.want {
				t.Fatalf("preferredResponseLanguage(%q) = %q, want %q", test.acceptLanguage, got, test.want)
			}
		})
	}
}
