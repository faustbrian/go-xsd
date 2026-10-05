package xsd_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faustbrian/go-xsd"
)

func TestParseOwnedNamespaceAdmission(t *testing.T) {
	source := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="a" type="xs:string" default="b"/></xs:schema>`)
	for _, maximum := range []int{1, 2} {
		document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxNamespaceEntries: maximum})
		if maximum == 1 {
			if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
				t.Fatal("expected namespace-copy limit refusal")
			}
		} else if err != nil || document.Elements[0].Default != "b" || document.Elements[0].Type != (xsd.QName{Namespace: xsd.Namespace, Local: "string"}) {
			t.Fatal("exact namespace allowance changed ordinary schema semantics")
		}
	}
	// A child without declarations or retained value scope aliases its parent's
	// namespace map and consumes no additional namespace entries.
	alias := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="a" type="xs:string"/></xs:schema>`)
	if _, err := xsd.Parse(context.Background(), alias, xsd.ParseOptions{MaxNamespaceEntries: 1}); err != nil {
		t.Fatal("aliased scope was charged as a copy")
	}
	// Rebinding still consumes an insertion at both independently owned grammar
	// scopes: root entry + element copy/rebind + body copy/rebind.
	rebound := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element xmlns:xs="http://www.w3.org/2001/XMLSchema" name="a" type="xs:string"/></xs:schema>`)
	for _, maximum := range []int{4, 5} {
		document, err := xsd.Parse(context.Background(), rebound, xsd.ParseOptions{MaxNamespaceEntries: maximum})
		if maximum == 4 {
			if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
				t.Fatal("rebinding work was not admitted")
			}
		} else if err != nil || document.Elements[0].Type.Namespace != xsd.Namespace {
			t.Fatal("rebinding allowance changed QName")
		}
	}
}

func TestParseOwnedAnnotationAdmission(t *testing.T) {
	source := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:annotation><xs:documentation>Read <b>this</b></xs:documentation><xs:appinfo><tool/></xs:appinfo></xs:annotation></xs:schema>`)
	// Namespace entry, captured documentation markup, wrapper copy work,
	// derived text builder work and retained text, then captured appinfo markup.
	const maximum = int64(2 + len(xsd.Namespace) + 2*len("Read <b>this</b>") + 2*len("Read this") + len("<tool/>"))
	for _, limit := range []int64{maximum - 1, maximum} {
		document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: limit})
		if limit < maximum {
			if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
				t.Fatal("expected annotation work limit refusal")
			}
		} else if err != nil || document.Annotations[0].Documentation[0].Markup != "Read <b>this</b>" || document.Annotations[0].Documentation[0].Content != "Read this" || document.Annotations[0].AppInformation[0].Content != "<tool/>" {
			t.Fatal("annotation accounting changed retained markup/text")
		}
	}
}

func TestParseOwnedAdmissionOptionsAndContext(t *testing.T) {
	source := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`)
	for _, options := range []xsd.ParseOptions{{MaxNamespaceEntries: -1}, {MaxModelBytes: -1}} {
		if document, err := xsd.Parse(context.Background(), source, options); document != nil || err == nil {
			t.Fatal("negative parse limit accepted")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	deadline, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	for _, ctx := range []context.Context{ctx, deadline} {
		if document, err := xsd.Parse(ctx, source, xsd.ParseOptions{}); document != nil || !errors.Is(err, ctx.Err()) {
			t.Fatal("parse changed real context cause")
		}
	}
	const maximum = int64(2 + len(xsd.Namespace) + 2*len("urn:root"))
	for _, limit := range []int64{maximum - 1, maximum} {
		document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{SystemID: "urn:root", MaxModelBytes: limit})
		if limit < maximum {
			if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
				t.Fatal("SystemID occurrences were not admitted")
			}
		} else if err != nil || document.SystemID != "urn:root" || document.BaseURI != "urn:root" {
			t.Fatal("SystemID exact allowance changed model")
		}
	}
}

func TestParseOwnedStringAdmission(t *testing.T) {
	source := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="a" type="xs:string" default="b"/></xs:schema>`)
	// Root namespace entry, element name, expanded type, default value,
	// and the independently owned namespace entry retained for that value.
	const maximum = int64((2 + len(xsd.Namespace)) + 1 + (len(xsd.Namespace) + 6) + 1 + (2 + len(xsd.Namespace)))
	for _, limit := range []int64{maximum - 1, maximum} {
		document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: limit})
		if limit < maximum {
			if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
				t.Fatal("expected retained model-string limit refusal")
			}
		} else if err != nil || document.Elements[0].Name != "a" {
			t.Fatal("exact model allowance changed schema semantics")
		}
	}
}

func TestParseOwnedDuplicateQNameAdmission(t *testing.T) {
	// Both namespace entries, the definition's name, and two retained QName
	// occurrences. Derivation additionally retains the expanded xs:string base.
	const ordinary = int64(2 + len(xsd.Namespace) + 1 + len("urn:t") + 1 + 2*(len("urn:t")+len("g")))
	for _, test := range []struct {
		name, body string
		maximum    int64
	}{
		{"complex", `<xs:complexType name="C"><xs:attributeGroup ref="t:g"/></xs:complexType>`, ordinary},
		{"derivation", `<xs:complexType name="C"><xs:simpleContent><xs:extension base="xs:string"><xs:attributeGroup ref="t:g"/></xs:extension></xs:simpleContent></xs:complexType>`, ordinary + int64(len(xsd.Namespace)+len("string"))},
		{"attribute group", `<xs:attributeGroup name="G"><xs:attributeGroup ref="t:g"/></xs:attributeGroup>`, ordinary},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:t="urn:t">` + test.body + `</xs:schema>`)
			for _, maximum := range []int64{test.maximum - 1, test.maximum} {
				document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: maximum})
				if maximum < test.maximum {
					if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
						t.Fatal("second retained QName occurrence was not admitted")
					}
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				var legacy xsd.QName
				var structured xsd.AttributeGroupReference
				if test.name == "attribute group" {
					legacy = document.AttributeGroups[0].References[0]
					structured = document.AttributeGroups[0].AttributeGroupReferences[0]
				} else {
					legacy = document.ComplexTypes[0].AttributeGroupRefs[0]
					structured = document.ComplexTypes[0].AttributeGroupReferences[0]
				}
				if legacy != (xsd.QName{Namespace: "urn:t", Local: "g"}) || structured.Ref != legacy {
					t.Fatal("accounting changed retained QName representations")
				}
			}
		})
	}
}
