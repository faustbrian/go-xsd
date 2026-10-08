package compile_test

import (
	"context"
	"errors"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/compile"
	"github.com/faustbrian/go-xsd/v2/resolve"
)

func TestCompilePropagatesRedefinitionFailure(t *testing.T) {
	const root = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:t="urn:test" targetNamespace="urn:test"><xs:redefine schemaLocation="base.xsd"><xs:simpleType name="Value"><xs:restriction base="t:Value"><xs:maxLength value="4"/></xs:restriction></xs:simpleType></xs:redefine></xs:schema>`
	for _, exists := range []bool{false, true} {
		definition := ""
		if exists {
			definition = `<xs:simpleType name="Value"><xs:restriction base="xs:string"><xs:minLength value="1"/></xs:restriction></xs:simpleType>`
		}
		memory, err := resolve.NewMemory(map[string][]byte{
			"https://example.test/base.xsd": []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:test">` + definition + `</xs:schema>`),
		})
		if err != nil {
			t.Fatal(err)
		}
		compiler, err := compile.New(compile.Options{Resolver: memory})
		if err != nil {
			t.Fatal(err)
		}
		set, err := compiler.Compile(t.Context(), compile.Source{URI: "https://example.test/root.xsd", Content: []byte(root)})
		if !exists {
			if set != nil || !errors.Is(err, compile.ErrUnresolvedComponent) {
				t.Fatal("missing redefinition target must fail without publishing a Set")
			}
			continue
		}
		if err != nil || set == nil {
			t.Fatal("valid self-restriction failed")
		}
		value, ok := set.SimpleType(xsd.QName{Namespace: "urn:test", Local: "Value"})
		if !ok || len(value.Facets) != 2 || value.Facets[0].Value != "1" || value.Facets[1].Value != "4" {
			t.Fatal("self-restriction lost original or added facets")
		}
	}
}

func TestPreCanceledCompileSkipsURIAdmission(t *testing.T) {
	compiler, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	measure := func(uri string) float64 {
		return testing.AllocsPerRun(1, func() {
			set, err := compiler.Compile(ctx, compile.Source{URI: uri})
			if set != nil || !errors.Is(err, context.Canceled) {
				t.Fatal("pre-canceled compilation lost its cause or published a Set")
			}
		})
	}
	ordinary, malformed := measure("urn:ordinary"), measure("%")
	if malformed != ordinary {
		t.Fatalf("canceled malformed URI performed admission work: %.0f allocations versus %.0f", malformed, ordinary)
	}
	if set, err := compiler.Compile(t.Context(), compile.Source{URI: "%"}); set != nil || err == nil || errors.Is(err, context.Canceled) {
		t.Fatal("fresh-context control did not reach malformed URI validation")
	}
}
