package resolve_test

import (
	"context"
	"errors"
	"testing"

	"github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/compile"
	"github.com/faustbrian/go-xsd/v2/resolve"
)

func TestMemoryConstructorAdmission(t *testing.T) {
	resources := map[string][]byte{"urn:a": []byte("ab"), "urn:b": []byte("cd")}
	for name, options := range map[string]resolve.MemoryOptions{
		"count":      {MaxResources: 1},
		"content":    {MaxResourceBytes: 1},
		"cumulative": {MaxBytes: 13},
		"identity":   {MaxBytes: 4},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := resolve.NewMemoryWithOptions(resources, options)
			if r != nil || !errors.Is(err, resolve.ErrLimitExceeded) {
				t.Fatal("expected nil resolver and constructor limit refusal")
			}
			if err.Error() != "xsd resolve: failed" {
				t.Fatal("unsafe default error category")
			}
		})
	}
	r, err := resolve.NewMemoryWithOptions(resources, resolve.MemoryOptions{MaxResources: 2, MaxBytes: 14, MaxResourceBytes: 2})
	if err != nil {
		t.Fatal(err)
	}
	resources["urn:a"][0] = 'x'
	first, err := r.Resolve(context.Background(), resolve.Request{URI: "urn:a"})
	if err != nil || string(first.Content) != "ab" {
		t.Fatal("constructor retained caller storage")
	}
	first.Content[0] = 'y'
	second, err := r.Resolve(context.Background(), resolve.Request{URI: "urn:a"})
	if err != nil || string(second.Content) != "ab" {
		t.Fatal("resolver retained returned storage")
	}
}

func TestCatalogConstructorAdmission(t *testing.T) {
	mappings := map[string]string{"a": "urn:a", "b": "urn:b"}
	for name, options := range map[string]resolve.CatalogOptions{
		"count": {MaxMappings: 1}, "cumulative": {MaxBytes: 11}, "namespace": {MaxBytes: 5},
	} {
		t.Run(name, func(t *testing.T) {
			r, err := resolve.NewCatalogWithOptions(mappings, resolve.Deny(), options)
			if r != nil || !errors.Is(err, resolve.ErrLimitExceeded) {
				t.Fatal("expected nil catalog and constructor limit refusal")
			}
			if err.Error() != "xsd resolve: failed" {
				t.Fatal("unsafe default error category")
			}
		})
	}
	memory, err := resolve.NewMemory(map[string][]byte{"urn:a": []byte("ab")})
	if err != nil {
		t.Fatal(err)
	}
	r, err := resolve.NewCatalogWithOptions(mappings, memory, resolve.CatalogOptions{MaxMappings: 2, MaxBytes: 12})
	if err != nil {
		t.Fatal(err)
	}
	mappings["a"] = "urn:missing"
	got, err := r.Resolve(context.Background(), resolve.Request{Namespace: "a", Kind: resolve.KindImport})
	if err != nil || got.URI != "urn:a" || string(got.Content) != "ab" {
		t.Fatal("catalog retained caller mappings")
	}
}

func TestConstructorAdmissionOptions(t *testing.T) {
	for _, options := range []resolve.MemoryOptions{{MaxResources: -1}, {MaxBytes: -1}, {MaxResourceBytes: -1}} {
		if r, err := resolve.NewMemoryWithOptions(nil, options); r != nil || err == nil {
			t.Fatal("negative memory option accepted")
		}
	}
	for _, options := range []resolve.CatalogOptions{{MaxMappings: -1}, {MaxBytes: -1}} {
		if r, err := resolve.NewCatalogWithOptions(nil, resolve.Deny(), options); r != nil || err == nil {
			t.Fatal("negative catalog option accepted")
		}
	}
	if _, err := resolve.NewMemoryWithOptions(nil, resolve.MemoryOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := resolve.NewCatalogWithOptions(nil, resolve.Deny(), resolve.CatalogOptions{}); err != nil {
		t.Fatal(err)
	}
}

func TestConstructorAdmissionStringBytes(t *testing.T) {
	for _, maximum := range []int64{4, 5} {
		r, err := resolve.NewMemoryWithOptions(map[string][]byte{"urn:a": nil}, resolve.MemoryOptions{MaxBytes: maximum})
		if maximum == 4 {
			if r != nil || !errors.Is(err, resolve.ErrLimitExceeded) {
				t.Fatal("URI bytes were not admitted")
			}
		} else if err != nil {
			t.Fatal(err)
		}
	}
	for _, maximum := range []int64{5, 6} {
		r, err := resolve.NewCatalogWithOptions(map[string]string{"a": "urn:a"}, resolve.Deny(), resolve.CatalogOptions{MaxBytes: maximum})
		if maximum == 5 {
			if r != nil || !errors.Is(err, resolve.ErrLimitExceeded) {
				t.Fatal("namespace bytes were not admitted")
			}
		} else if err != nil {
			t.Fatal(err)
		}
	}
}

func TestAdmittedResolversCompileOrdinaryImport(t *testing.T) {
	const identity = "urn:schema:dependency"
	content := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:dependency"><xs:simpleType name="Code"><xs:restriction base="xs:string"/></xs:simpleType></xs:schema>`)
	memory, err := resolve.NewMemoryWithOptions(map[string][]byte{identity: content}, resolve.MemoryOptions{
		MaxResources: 1, MaxBytes: int64(len(identity) + len(content)), MaxResourceBytes: int64(len(content)),
	})
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := resolve.NewCatalogWithOptions(map[string]string{"urn:dependency": identity}, memory, resolve.CatalogOptions{
		MaxMappings: 1, MaxBytes: int64(len("urn:dependency") + len(identity)),
	})
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := compile.New(compile.Options{Resolver: catalog})
	if err != nil {
		t.Fatal(err)
	}
	set, err := compiler.Compile(context.Background(), compile.Source{URI: "urn:schema:root", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:dep="urn:dependency"><xs:import namespace="urn:dependency"/><xs:element name="root" type="dep:Code"/></xs:schema>`)})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := set.SimpleType(xsd.QName{Namespace: "urn:dependency", Local: "Code"}); !ok {
		t.Fatal("admitted import did not contribute type")
	}
}
