package compile_test

import (
	"context"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/compile"
)

// The selected WSDL component surface is a resolved projection, not a
// lossless source document. Keep the source model when lexical IDs matter.
func TestR32ProjectionPreservesResolvedComponentMeaning(t *testing.T) {
	const source = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:t="urn:r32" targetNamespace="urn:r32"><xs:group name="Fields"><xs:sequence id="source-sequence"><xs:element name="leaf" type="xs:string"/></xs:sequence></xs:group><xs:complexType name="Root"><xs:group ref="t:Fields"/></xs:complexType><xs:element name="root" type="t:Root"/></xs:schema>`
	ctx := context.Background()
	raw, err := xsd.Parse(ctx, []byte(source), xsd.ParseOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if raw.ModelGroups[0].Content.ID != "source-sequence" {
		t.Fatal("source model lost represented lexical metadata")
	}
	compiler, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	set, err := compiler.Compile(ctx, compile.Source{URI: "urn:r32:projection", Content: []byte(source)})
	if err != nil {
		t.Fatal(err)
	}
	name := xsd.QName{Namespace: "urn:r32", Local: "Root"}
	projected, ok := set.ComplexType(name)
	if !ok || projected.Content == nil || len(projected.Content.Particles) != 1 || projected.Content.Particles[0].Group == nil {
		t.Fatal("resolved group projection missing")
	}
	group := projected.Content.Particles[0].Group
	if len(group.Particles) != 1 || group.Particles[0].Element == nil || group.Particles[0].Element.Type != (xsd.QName{Namespace: xsd.Namespace, Local: "string"}) || group.Particles[0].GroupRef.Local != "" {
		t.Fatal("resolved component meaning changed")
	}
	group.Particles[0].Element.Name = "changed"
	again, ok := set.ComplexType(name)
	if !ok || again.Content.Particles[0].Group.Particles[0].Element.Name != "leaf" {
		t.Fatal("projection mutation changed immutable set")
	}
	if raw.ModelGroups[0].Content.ID != "source-sequence" {
		t.Fatal("compilation/projection mutated source model")
	}
}
