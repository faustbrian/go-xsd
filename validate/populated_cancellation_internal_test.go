package validate

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/compile"
)

func TestPopulatedFacetsStopAfterCancellation(t *testing.T) {
	validator := populatedCancellationValidator(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state := validationState{validator: validator, ctx: ctx}
	definition := xsd.SimpleType{
		Variety: xsd.SimpleRestriction,
		Base:    xsd.QName{Namespace: xsd.Namespace, Local: "string"},
		Facets: []xsd.Facet{
			{Kind: xsd.FacetMinLength, Value: "1"},
			{Kind: xsd.FacetMaxLength, Value: "4"},
			{Kind: xsd.FacetEnumeration, Value: "ok"},
		},
	}
	if !state.facetsValid(definition, "ok") || state.facetsValid(definition, "bad") {
		t.Fatal("live populated facets lost ordinary acceptance or rejection")
	}
	cancel()
	if state.facetsValid(definition, "ok") || !errors.Is(state.contextError(), context.Canceled) {
		t.Fatal("canceled populated facets accepted an otherwise valid value or lost cause")
	}
	if len(state.diagnostics) != 0 {
		t.Fatal("cancellation became an instance diagnostic")
	}
}

func TestPopulatedIdentityStateStopsAfterCancellation(t *testing.T) {
	validator := populatedCancellationValidator(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	state := validationState{validator: validator, ctx: ctx}
	root := identityTestNode("root")
	root.Name = xsd.QName{Local: "root"}
	for _, value := range []string{"a", "b"} {
		child := identityTestNode(value)
		child.Attributes[xsd.QName{Local: "id"}] = value
		root.Children = append(root.Children, child)
	}
	constraints := []xsd.IdentityConstraint{{
		Kind: xsd.IdentityKey, Name: "key", Selector: "child", Fields: []string{"@id"},
	}}
	if err := state.validateIdentityConstraints(root, constraints, "/root"); err != nil {
		t.Fatal(err)
	}
	key := xsd.QName{Local: "key"}
	table := maps.Clone(state.identityTables[root][key])
	if len(table) != 2 || state.xpathSteps == 0 || state.identityValues != 2 || len(state.diagnostics) != 0 {
		t.Fatal("live identity assessment did not populate the expected unique-key table")
	}
	steps, values := state.xpathSteps, state.identityValues
	cancel()
	if err := state.validateIdentityConstraints(root, constraints, "/root"); !errors.Is(err, context.Canceled) {
		t.Fatalf("canceled populated identity assessment = %v", err)
	}
	if state.xpathSteps != steps || state.identityValues != values ||
		!reflect.DeepEqual(state.identityTables[root][key], table) || len(state.diagnostics) != 0 {
		t.Fatal("canceled identity reassessment changed prior state or consumed work")
	}
	if nodes := state.identityDescendants(root); len(nodes) != 0 {
		t.Fatal("canceled populated identity traversal continued")
	}
}

func TestCompletedTreeCloneRejectsCanceledAssessment(t *testing.T) {
	validator := populatedCancellationValidator(t)
	source := Node{
		Name:       xsd.QName{Local: "root"},
		Namespaces: map[string]string{"xs": xsd.Namespace},
		Children: []Node{{
			Name: xsd.QName{Local: "child"}, Text: "ok",
			Attributes: map[xsd.QName]string{{Local: "id"}: "a"},
		}},
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	liveClone, err := (&treeCloneState{validator: validator}).clone(ctx, source, 1)
	if err != nil {
		t.Fatal(err)
	}
	result, err := validator.validateRoot(ctx, liveClone)
	if err != nil || !result.Valid || len(result.Diagnostics) != 0 {
		t.Fatalf("ordinary completed clone assessment = %#v, %v", result, err)
	}
	clone, err := (&treeCloneState{validator: validator}).clone(ctx, source, 1)
	if err != nil || len(clone.Children) != 1 || clone.Children[0].Text != "ok" {
		t.Fatalf("populated clone = %#v, %v", clone, err)
	}
	cancel()
	result, err = validator.validateRoot(ctx, clone)
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(result, Result{}) {
		t.Fatalf("clone-to-assessment cancellation = %#v, %v", result, err)
	}
	for _, node := range []*instanceNode{clone, clone.Children[0]} {
		if node.Type != (xsd.QName{}) || node.Nillable || len(node.AttributeTypes) != 0 {
			t.Fatal("canceled assessment published partial clone annotations")
		}
	}
	if source.Children[0].Text != "ok" || source.Children[0].Attributes[xsd.QName{Local: "id"}] != "a" ||
		source.Namespaces["xs"] != xsd.Namespace {
		t.Fatal("clone or assessment modified caller-owned payload")
	}
}

func populatedCancellationValidator(t *testing.T) *Validator {
	t.Helper()
	compiler, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	set, err := compiler.Compile(context.Background(), compile.Source{
		URI: "urn:populated-cancellation",
		Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
 <xs:element name="root"><xs:complexType><xs:sequence>
  <xs:element name="child"><xs:complexType><xs:simpleContent>
   <xs:extension base="xs:string"><xs:attribute name="id" type="xs:string"/></xs:extension>
  </xs:simpleContent></xs:complexType></xs:element>
 </xs:sequence></xs:complexType></xs:element>
</xs:schema>`),
	})
	if err != nil {
		t.Fatal(err)
	}
	validator, err := New(set, Options{})
	if err != nil {
		t.Fatal(err)
	}
	return validator
}
