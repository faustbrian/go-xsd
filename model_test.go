package xsd

import (
	"context"
	"encoding/xml"
	"testing"
	"time"
)

func TestDerivationSetReportsMembershipAndLexicalValue(t *testing.T) {
	document, err := Parse(context.Background(), []byte(`<?xml version="1.0"?>
<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 blockDefault="extension restriction" finalDefault="#all"/>`), ParseOptions{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !document.BlockDefault.Contains(DerivationExtension) ||
		document.BlockDefault.Contains(DerivationUnion) {
		t.Fatalf("BlockDefault membership = %v", document.BlockDefault)
	}
	if got := document.BlockDefault.String(); got != "extension restriction" {
		t.Fatalf("BlockDefault.String() = %q", got)
	}
	if !document.FinalDefault.All() ||
		!document.FinalDefault.Contains(DerivationUnion) ||
		document.FinalDefault.String() != "#all" {
		t.Fatalf("FinalDefault = %v", document.FinalDefault)
	}
}

func TestParserAndSerializerCoverNestedQNameBranches(t *testing.T) {
	element, err := parseElement(xml.StartElement{Attr: []xml.Attr{
		{Name: xml.Name{Local: "form"}, Value: "qualified"},
	}}, nil)
	if err != nil || element.Form != FormQualified {
		t.Fatalf("parseElement() = %#v, %v", element, err)
	}

	attribute, err := parseAttributeUse(xml.StartElement{Attr: []xml.Attr{
		{Name: xml.Name{Space: "xmlns", Local: "t"}, Value: "urn:test"},
		{Name: xml.Name{Local: "ref"}, Value: "t:code"},
	}}, nil)
	if err != nil || attribute.Ref != (QName{Namespace: "urn:test", Local: "code"}) {
		t.Fatalf("parseAttributeUse() = %#v, %v", attribute, err)
	}

	decoder, start := decoderAtStart(t, `<element xmlns="`+Namespace+`"><complexType><sequence></complexType></element>`)
	done := make(chan error, 1)
	go func() {
		done <- parseElementBody(decoder, start, &Element{}, nil)
	}()
	var parseErr error
	select {
	case parseErr = <-done:
	case <-time.After(time.Second):
		t.Fatal("parseElementBody() did not terminate")
	}
	if parseErr == nil {
		t.Fatal("parseElementBody() accepted an invalid inline complex type")
	}

	namespaces := documentQNameNamespaces(&Document{Elements: []Element{{
		InlineComplexType: &ComplexType{Base: QName{Namespace: "urn:test", Local: "Record"}},
	}}})
	if _, ok := namespaces["urn:test"]; !ok {
		t.Fatalf("documentQNameNamespaces() = %#v", namespaces)
	}
}
