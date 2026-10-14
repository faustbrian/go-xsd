package builder_test

import (
	"context"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/builder"
)

// docs/builders.md promises copied inputs. Observe the published Build model,
// not the builder's private clone representation.
func TestR32RestrictionRetainsOwnedFacetSnapshot(t *testing.T) {
	for _, change := range []string{"namespace", "annotation", "slice_value"} {
		t.Run(change, func(t *testing.T) {
			annotation := &xsd.Annotation{Documentation: []xsd.Documentation{{Content: "original"}}}
			namespaces := map[string]string{"p": "urn:original"}
			facets := []xsd.Facet{{Kind: xsd.FacetEnumeration, Value: "p:accepted", Namespaces: namespaces, Annotation: annotation}}
			schema := builder.New("urn:r32")
			if err := schema.AddSimpleRestriction("Token", xsd.QName{Namespace: xsd.Namespace, Local: "QName"}, facets...); err != nil {
				t.Fatal(err)
			}
			switch change {
			case "namespace":
				namespaces["p"] = "urn:changed"
			case "annotation":
				annotation.Documentation[0].Content = "changed"
			case "slice_value":
				facets[0].Value = "p:changed"
			}
			document, err := schema.Build(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if len(document.SimpleTypes) != 1 || len(document.SimpleTypes[0].Facets) != 1 {
				t.Fatal("restriction snapshot missing")
			}
			got := document.SimpleTypes[0].Facets[0]
			if got.Value != "p:accepted" || got.Namespaces["p"] != "urn:original" || got.Annotation == nil || len(got.Annotation.Documentation) != 1 || got.Annotation.Documentation[0].Content != "original" {
				t.Fatalf("Build retained caller mutation: value=%q namespace=%q annotation=%#v", got.Value, got.Namespaces["p"], got.Annotation)
			}
			// A returned document is independently mutable; rebuilding keeps
			// the builder's retained snapshot, without replaying caller mutation.
			got.Namespaces["p"] = "urn:output-mutated"
			got.Annotation.Documentation[0].Content = "output-mutated"
			again, err := schema.Build(context.Background())
			if err != nil {
				t.Fatal(err)
			}
			if again.SimpleTypes[0].Facets[0].Namespaces["p"] != "urn:original" || again.SimpleTypes[0].Facets[0].Annotation.Documentation[0].Content != "original" {
				t.Fatal("returned model mutation poisoned rebuild")
			}
		})
	}
}

func TestR32FiniteRepeatedComplexInputIsCopied(t *testing.T) {
	leaf := &xsd.ModelGroup{Compositor: xsd.Sequence, Particles: []xsd.Particle{{MinOccurs: 1, MaxOccurs: 1, Element: &xsd.Element{Name: "leaf", Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}}}}
	content := &xsd.ModelGroup{Compositor: xsd.Sequence, Particles: []xsd.Particle{{MinOccurs: 1, MaxOccurs: 1, Group: leaf}, {MinOccurs: 1, MaxOccurs: 1, Group: leaf}}}
	schema := builder.New("urn:r32")
	if err := schema.AddComplexType("Pair", content); err != nil {
		t.Fatal(err)
	}
	leaf.Particles[0].Element.Name = "changed"
	document, err := schema.Build(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, particle := range document.ComplexTypes[0].Content.Particles {
		if particle.Group == nil || particle.Group.Particles[0].Element.Name != "leaf" {
			t.Fatal("finite repeated input retained an alias")
		}
	}
}
