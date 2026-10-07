package xsd

import (
	"context"
	"encoding/xml"
	"errors"
	"strings"
	"testing"
)

// Keep this representative parser contract nonparallel so fail-fast mutation
// runs reject broken token handling before scheduling broader parallel suites.
func TestRepresentativeSchemaParsesBeforeParallelCampaigns(t *testing.T) {
	// Check the token boundary before Parse can loop on a nil token and nil
	// error. Refusal must not consume input, while a live owner must progress.
	decoder := xml.NewDecoder(strings.NewReader(`<schema/>`))
	owner := newSchemaParser(context.Background(), nil, ParseOptions{})
	token, err := owner.token(decoder)
	start, ok := token.(xml.StartElement)
	if err != nil || !ok || start.Name.Local != "schema" || decoder.InputOffset() == 0 {
		t.Fatal("live parser owner did not deliver the first XML token")
	}

	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	decoder = xml.NewDecoder(strings.NewReader(`<schema/>`))
	owner = newSchemaParser(canceled, nil, ParseOptions{})
	if token, err := owner.token(decoder); token != nil || !errors.Is(err, context.Canceled) || decoder.InputOffset() != 0 {
		t.Fatal("canceled parser owner did not refuse before consuming input")
	}

	decoder = xml.NewDecoder(strings.NewReader(`<schema/>`))
	owner = newSchemaParser(context.Background(), nil, ParseOptions{MaxModelBytes: 1})
	if owner.chargeBytes(2) || !errors.Is(owner.err, ErrLimitExceeded) {
		t.Fatal("model allowance did not establish an admission refusal")
	}
	if token, err := owner.token(decoder); token != nil || err != owner.err || decoder.InputOffset() != 0 {
		t.Fatal("sticky parser owner did not preserve refusal before consuming input")
	}

	source := []byte(`<schema xmlns="` + Namespace + `">
<simpleType name="Code"><restriction base="string"/></simpleType>
<element name="root"><key name="identity"><selector xpath="."/><field xpath="@id"/></key></element>
</schema>`)
	document, err := Parse(context.Background(), source, ParseOptions{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if document == nil || len(document.SimpleTypes) != 1 ||
		document.SimpleTypes[0].Name != "Code" || document.SimpleTypes[0].Variety != SimpleRestriction ||
		len(document.Elements) != 1 ||
		len(document.Elements[0].IdentityConstraints) != 1 {
		t.Fatalf("Parse() document = %#v", document)
	}
}
