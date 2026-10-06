package compile_test

import (
	"context"
	"errors"
	"testing"

	"github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/compile"
	"github.com/faustbrian/go-xsd/v2/resolve"
)

func TestCompilerPropagatesPerDocumentParseAdmission(t *testing.T) {
	child := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="a" type="xs:string" default="b"/></xs:schema>`)
	memory, err := resolve.NewMemory(map[string][]byte{"urn:child": child})
	if err != nil {
		t.Fatal(err)
	}
	for _, root := range []compile.Source{
		{URI: "urn:root", Content: child},
		{URI: "urn:root", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:include schemaLocation="urn:child"/></xs:schema>`)},
	} {
		for _, maximum := range []int{1, 2} {
			compiler, err := compile.New(compile.Options{Resolver: memory, Limits: compile.Limits{MaxParseNamespaceEntries: maximum}})
			if err != nil {
				t.Fatal(err)
			}
			set, err := compiler.Compile(context.Background(), root)
			if maximum == 1 {
				if set != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
					t.Fatal("compiler did not forward per-document parse limit")
				}
			} else if err != nil || set == nil {
				t.Fatal("per-document allowance became graph cumulative")
			}
		}
	}
	for _, limits := range []compile.Limits{{MaxParseNamespaceEntries: -1}, {MaxParseModelBytes: -1}} {
		if compiler, err := compile.New(compile.Options{Limits: limits}); compiler != nil || err == nil {
			t.Fatal("negative parser allowance accepted")
		}
	}
	for _, test := range []struct {
		root    compile.Source
		maximum int64
	}{
		{compile.Source{URI: "urn:root", Content: child}, 123},
		{compile.Source{URI: "urn:root", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:include schemaLocation="urn:child"/></xs:schema>`)}, 125},
	} {
		for _, maximum := range []int64{test.maximum, 126} {
			compiler, err := compile.New(compile.Options{Resolver: memory, Limits: compile.Limits{MaxParseModelBytes: maximum}})
			if err != nil {
				t.Fatal(err)
			}
			set, err := compiler.Compile(context.Background(), test.root)
			if maximum == test.maximum {
				if set != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
					t.Fatal("compiler did not forward model-string allowance")
				}
			} else if err != nil || set == nil {
				t.Fatal("model-string allowance became graph cumulative")
			}
		}
	}
}
