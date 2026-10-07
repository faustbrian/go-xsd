package xsd_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/faustbrian/go-xsd/v2"
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

func TestParseOwnedWorkAdmissionBoundaries(t *testing.T) {
	const namespaceBytes = int64(2 + len(xsd.Namespace))
	for _, test := range []struct {
		name, body string
		refusals   []int64
		maximum    int64
		valid      func(*xsd.Document) bool
	}{
		{
			name: "union member splitting", body: `<xs:simpleType name="U"><xs:union memberTypes="xs:string"/></xs:simpleType>`,
			refusals: []int64{namespaceBytes + 1 + int64(len("xs:string")) - 1},
			maximum:  namespaceBytes + 1 + int64(len("xs:string")+len(xsd.Namespace)+len("string")),
			valid: func(d *xsd.Document) bool {
				return len(d.SimpleTypes) == 1 && d.SimpleTypes[0].Name == "U" && len(d.SimpleTypes[0].MemberTypes) == 1 && d.SimpleTypes[0].MemberTypes[0] == (xsd.QName{Namespace: xsd.Namespace, Local: "string"})
			},
		},
		{
			name: "wildcard namespace splitting", body: `<xs:complexType name="C"><xs:anyAttribute namespace="urn:a"/></xs:complexType>`,
			refusals: []int64{namespaceBytes + 1 + int64(len("urn:a")) - 1}, maximum: namespaceBytes + 1 + int64(len("urn:a")),
			valid: func(d *xsd.Document) bool {
				return len(d.ComplexTypes) == 1 && d.ComplexTypes[0].Name == "C" && d.ComplexTypes[0].AttributeWildcard != nil && len(d.ComplexTypes[0].AttributeWildcard.Namespaces) == 1 && d.ComplexTypes[0].AttributeWildcard.Namespaces[0] == "urn:a"
			},
		},
		{
			// Location retention, three-byte URI-work envelope, resolved URI.
			name: "URI work", body: `<xs:include schemaLocation="urn:a"/>`,
			refusals: []int64{namespaceBytes + 2*int64(len("urn:a")) - 1}, maximum: namespaceBytes + 5*int64(len("urn:a")),
			valid: func(d *xsd.Document) bool {
				return len(d.References) == 1 && d.References[0].Kind == xsd.ReferenceInclude && d.References[0].Location == "urn:a" && d.References[0].URI == "urn:a"
			},
		},
		{
			// Redefinitions retain another Location and URI occurrence.
			name: "retained redefine reference", body: `<xs:redefine schemaLocation="urn:a"/>`,
			refusals: []int64{namespaceBytes + 7*int64(len("urn:a")) - 1}, maximum: namespaceBytes + 7*int64(len("urn:a")),
			valid: func(d *xsd.Document) bool {
				return len(d.References) == 1 && len(d.Redefinitions) == 1 && d.References[0].URI == "urn:a" && d.Redefinitions[0].Reference.URI == "urn:a" && d.Redefinitions[0].Reference.Location == "urn:a"
			},
		},
		{
			name: "appinfo metadata before capture", body: `<xs:annotation><xs:appinfo source="urn:a"><tool/></xs:appinfo></xs:annotation>`,
			refusals: []int64{namespaceBytes + int64(len("urn:a")) - 1}, maximum: namespaceBytes + int64(len("urn:a")+len("<tool/>")),
			valid: func(d *xsd.Document) bool {
				return len(d.Annotations) == 1 && len(d.Annotations[0].AppInformation) == 1 && d.Annotations[0].AppInformation[0].Source == "urn:a" && d.Annotations[0].AppInformation[0].Content == "<tool/>"
			},
		},
		{
			// Raw capture, wrapper copy, builder write, retained trimmed text.
			name: "documentation copy and builder work", body: `<xs:annotation><xs:documentation>a</xs:documentation></xs:annotation>`,
			refusals: []int64{namespaceBytes + 1, namespaceBytes + 2}, maximum: namespaceBytes + 4,
			valid: func(d *xsd.Document) bool {
				return len(d.Annotations) == 1 && len(d.Annotations[0].Documentation) == 1 && d.Annotations[0].Documentation[0].Markup == "a" && d.Annotations[0].Documentation[0].Content == "a"
			},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">` + test.body + `</xs:schema>`)
			for _, maximum := range test.refusals {
				document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: maximum})
				if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
					t.Fatal("expected owned-work refusal without a document")
				}
			}
			document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: test.maximum})
			if err != nil || document == nil || !test.valid(document) {
				t.Fatal("exact owned-work allowance changed literal schema semantics")
			}
		})
	}
}

func TestParseOwnedNamespaceWorkBoundaries(t *testing.T) {
	for _, test := range []struct {
		source  string
		refused int
	}{
		{`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element xmlns:t="urn:t" name="a"/></xs:schema>`, 1},
		{`<schema xmlns="http://www.w3.org/2001/XMLSchema"><element xmlns="http://www.w3.org/2001/XMLSchema" name="a"/></schema>`, 2},
		{`<schema xmlns="http://www.w3.org/2001/XMLSchema"><element xmlns="http://www.w3.org/2001/XMLSchema" name="a"/></schema>`, 4},
	} {
		document, err := xsd.Parse(context.Background(), []byte(test.source), xsd.ParseOptions{MaxNamespaceEntries: test.refused})
		if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
			t.Fatal("expected scope-copy/declaration refusal")
		}
		// Root insertion plus two separately owned parent-copy/declaration pairs.
		document, err = xsd.Parse(context.Background(), []byte(test.source), xsd.ParseOptions{MaxNamespaceEntries: 5})
		if err != nil || document == nil || len(document.Elements) != 1 || document.Elements[0].Name != "a" {
			t.Fatal("exact scope allowance changed element identity")
		}
	}
	const namespaceBytes = int64(len(xsd.Namespace))
	source := []byte(`<schema xmlns="http://www.w3.org/2001/XMLSchema"/>`)
	document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: namespaceBytes - 1})
	if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
		t.Fatal("default namespace URI was not admitted")
	}
	document, err = xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: namespaceBytes})
	if err != nil || document == nil || document.Namespaces[""] != xsd.Namespace {
		t.Fatal("exact default namespace allowance changed binding")
	}

	// Refused name retention remains sticky through default retention and the
	// subsequent value-scope copy; no partial model may escape those owners.
	source = []byte(`<schema xmlns="http://www.w3.org/2001/XMLSchema"><element name="a" default="b"/></schema>`)
	document, err = xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: namespaceBytes})
	if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
		t.Fatal("retention failure did not remain a limit refusal")
	}
	document, err = xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxModelBytes: 2*namespaceBytes + 2})
	if err != nil || document == nil || len(document.Elements) != 1 || document.Elements[0].Name != "a" || document.Elements[0].Default != "b" || document.Elements[0].ValueNamespaces[""] != xsd.Namespace {
		t.Fatal("exact retained value allowance changed default/scope")
	}
}

func TestParseRootMetadataIsNotANamespaceDeclaration(t *testing.T) {
	source := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" version="a"/>`)
	const maximum = int64(2 + len(xsd.Namespace) + 1)
	for _, limit := range []int64{maximum - 1, maximum} {
		document, err := xsd.Parse(context.Background(), source, xsd.ParseOptions{MaxNamespaceEntries: 1, MaxModelBytes: limit})
		if limit < maximum {
			if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
				t.Fatal("root version metadata bypassed its model allowance")
			}
		} else if err != nil || document == nil || document.Version != "a" || document.Namespaces["xs"] != xsd.Namespace {
			t.Fatal("ordinary root metadata consumed a namespace entry or changed identity")
		}
	}
}
