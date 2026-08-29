package validate

import (
	"context"
	"testing"

	xsd "github.com/faustbrian/go-xsd"
	"github.com/faustbrian/go-xsd/compile"
)

func TestValidatorCoversElementTypeAndNilBranches(t *testing.T) {
	t.Parallel()

	set := advancedValidationSet(t)
	state := validationState{validator: &Validator{set: set, limits: coverageLimits()}}

	requireDiagnostic(t, &state, func() error {
		return state.validateElement(contentNode("", nil), xsd.Element{Abstract: true}, "/root")
	}, "validateElement(abstract)")
	requireDiagnostic(t, &state, func() error {
		nilledChildren := contentNode("", map[xsd.QName]string{
			{Namespace: schemaInstanceNamespace, Local: "nil"}: "true",
		})
		nilledChildren.Children = []*instanceNode{contentNode("", nil)}
		return state.validateElementContent(
			nilledChildren,
			xsd.Element{Nillable: true},
			"/root",
		)
	}, "validateElementContent(nilled children)")

	unknownType := contentNode("", map[xsd.QName]string{
		{Namespace: schemaInstanceNamespace, Local: "type"}: "t:Missing",
	})
	unknownType.Namespaces["t"] = "urn:test"
	requireDiagnostic(t, &state, func() error {
		return state.validateElementContent(unknownType, xsd.Element{}, "/root")
	}, "validateElementContent(unknown xsi:type)")

	abstractType := contentNode("", map[xsd.QName]string{
		{Namespace: schemaInstanceNamespace, Local: "type"}: "t:Abstract",
	})
	abstractType.Namespaces["t"] = "urn:test"
	requireDiagnostic(t, &state, func() error {
		return state.validateElementContent(abstractType, xsd.Element{
			Type: xsd.QName{Namespace: "urn:test", Local: "Base"},
		}, "/root")
	}, "validateElementContent(abstract xsi:type)")

	unrelatedType := contentNode("", map[xsd.QName]string{
		{Namespace: schemaInstanceNamespace, Local: "type"}: "t:Other",
	})
	unrelatedType.Namespaces["t"] = "urn:test"
	requireDiagnostic(t, &state, func() error {
		return state.validateElementContent(unrelatedType, xsd.Element{
			Type: xsd.QName{Namespace: "urn:test", Local: "Base"},
		}, "/root")
	}, "validateElementContent(unrelated xsi:type)")

	nilledInline := contentNode("", map[xsd.QName]string{
		{Namespace: schemaInstanceNamespace, Local: "nil"}: "true",
	})
	if err := state.validateElementContent(nilledInline, xsd.Element{
		Nillable:          true,
		InlineComplexType: &xsd.ComplexType{},
	}, "/root"); err != nil {
		t.Fatalf("validateElementContent(nilled inline complex) error = %v", err)
	}

	if !state.typeExists(xsd.QName{Namespace: "urn:test", Local: "Union"}) {
		t.Fatal("typeExists(simple type) returned false")
	}
}

func TestValidatorCoversSimpleIdentityAndFacetBranches(t *testing.T) {
	t.Parallel()

	set := advancedValidationSet(t)
	state := validationState{validator: &Validator{set: set, limits: coverageLimits()}}
	list := xsd.SimpleType{
		Variety:  xsd.SimpleList,
		ItemType: xsd.QName{Namespace: xsd.Namespace, Local: "boolean"},
	}
	if state.inlineSimpleLexicalValid(list, "") {
		t.Fatal("inlineSimpleLexicalValid(empty list) succeeded")
	}
	if state.inlineSimpleLexicalValid(list, "invalid") {
		t.Fatal("inlineSimpleLexicalValid(invalid item) succeeded")
	}

	nonKey := xsd.IdentityConstraint{Kind: xsd.IdentityUnique, Fields: []string{"@missing"}}
	if err := state.addIdentityDefinitionValue(
		identityTestNode("value"), nonKey, "/root", identityNodeTable{}, map[string]struct{}{},
	); err != nil {
		t.Fatalf("addIdentityDefinitionValue(incomplete unique) error = %v", err)
	}

	leaf := identityTestNode("leaf")
	leaf.Name = xsd.QName{Namespace: "urn:test", Local: "leaf"}
	item := identityTestNode("item")
	item.Name = xsd.QName{Namespace: "urn:test", Local: "item"}
	item.Children = []*instanceNode{leaf}
	root := identityTestNode("root")
	root.Children = []*instanceNode{item}
	selected := selectIdentityNodes(root, xsd.IdentityConstraint{
		Selector:   ".//t:item/t:leaf",
		Namespaces: map[string]string{"t": "urn:test"},
	})
	if len(selected) != 1 || selected[0] != leaf {
		t.Fatalf("selectIdentityNodes(descendant) = %#v", selected)
	}

	root.Attributes = map[xsd.QName]string{{Local: "id"}: "one", {Local: "other"}: "two"}
	values := state.identityFieldValues(root, "@id | @other | @id", nil)
	if len(values) != 2 {
		t.Fatalf("identityFieldValues(union) = %#v", values)
	}

	equal, err := state.simpleValuesEqual(
		xsd.QName{Namespace: "urn:test", Local: "Union"}, "left", "right",
	)
	if err != nil || equal {
		t.Fatalf("simpleValuesEqual(unmatched union) = %t, %v", equal, err)
	}
	requireDiagnostic(t, &state, func() error {
		return state.validateSimple(nodeWithChild(contentNode("", nil)), xsd.QName{
			Namespace: xsd.Namespace, Local: "string",
		}, "/root")
	}, "validateSimple(child)")
	if state.simpleLexicalValid(xsd.QName{Namespace: xsd.Namespace, Local: "int"}, "invalid") {
		t.Fatal("simpleLexicalValid(invalid integer) succeeded")
	}

	inlineListRestriction := xsd.SimpleType{
		Base: xsd.QName{Namespace: xsd.Namespace, Local: "string"},
		InlineBase: &xsd.SimpleType{
			Variety:  xsd.SimpleList,
			ItemType: xsd.QName{Namespace: xsd.Namespace, Local: "string"},
		},
		Facets: []xsd.Facet{{Kind: xsd.FacetLength, Value: "2"}},
	}
	if !state.facetsValid(inlineListRestriction, "one two") {
		t.Fatal("facetsValid(inline list) returned false")
	}
	nmtokens := xsd.SimpleType{
		Base:   xsd.QName{Namespace: xsd.Namespace, Local: "NMTOKENS"},
		Facets: []xsd.Facet{{Kind: xsd.FacetLength, Value: "2"}},
	}
	if !state.facetsValid(nmtokens, "one two") {
		t.Fatal("facetsValid(NMTOKENS) returned false")
	}
	if state.simpleTypeUsesNamespaceContext(xsd.QName{Namespace: "urn:test", Local: "Union"}) {
		t.Fatal("simpleTypeUsesNamespaceContext(string union) returned true")
	}

	if !state.numericFacetValid(xsd.QName{Namespace: xsd.Namespace, Local: "decimal"}, "1.23", xsd.Facet{
		Kind: xsd.FacetFractionDigits, Value: "2",
	}) {
		t.Fatal("numericFacetValid(decimal fraction) returned false")
	}
	if !state.numericFacetValid(xsd.QName{Namespace: xsd.Namespace, Local: "double"}, "1", xsd.Facet{
		Kind: xsd.FacetMinInclusive, Value: "0",
	}) {
		t.Fatal("numericFacetValid(double) returned false")
	}
	if !state.numericFacetValid(xsd.QName{Namespace: xsd.Namespace, Local: "duration"}, "P1D", xsd.Facet{
		Kind: xsd.FacetMaxInclusive, Value: "P2D",
	}) {
		t.Fatal("numericFacetValid(duration) returned false")
	}
	if !state.numericFacetValid(xsd.QName{Namespace: xsd.Namespace, Local: "date"}, "2026-01-01", xsd.Facet{
		Kind: xsd.FacetMinInclusive, Value: "2025-01-01",
	}) {
		t.Fatal("numericFacetValid(date) returned false")
	}
	if durationValuesEqual("P1Y", "P1M") {
		t.Fatal("durationValuesEqual(different months) returned true")
	}
	if got := state.normalizeRestrictionLexical(xsd.SimpleType{
		Base:       xsd.QName{Namespace: xsd.Namespace, Local: "string"},
		InlineBase: &list,
	}, " one\t two "); got != "one two" {
		t.Fatalf("normalizeRestrictionLexical(inline list) = %q", got)
	}
}

func TestValidatorCoversWildcardAndNullableParticleBranches(t *testing.T) {
	t.Parallel()

	set := advancedValidationSet(t)
	state := validationState{validator: &Validator{set: set, limits: coverageLimits()}}
	attribute := contentNode("", map[xsd.QName]string{{Namespace: "urn:other", Local: "missing"}: "value"})
	if err := state.validateAdditionalAttribute(attribute, xsd.QName{
		Namespace: "urn:other", Local: "missing",
	}, "urn:test", &xsd.Wildcard{
		Namespaces: []string{"##other"}, ProcessContents: xsd.ProcessLax,
	}, "/root"); err != nil {
		t.Fatalf("validateAdditionalAttribute(lax) error = %v", err)
	}

	child := contentNode("true", nil)
	child.Name = xsd.QName{Namespace: "urn:test", Local: "known"}
	unknown := contentNode("", nil)
	unknown.Name = xsd.QName{Namespace: "urn:other", Local: "unknown"}
	for _, test := range []struct {
		name     string
		child    *instanceNode
		contents xsd.ProcessContents
		wantErr  bool
	}{
		{name: "skip", child: unknown, contents: xsd.ProcessSkip},
		{name: "lax", child: unknown, contents: xsd.ProcessLax},
		{name: "strict undeclared", child: unknown, contents: xsd.ProcessStrict, wantErr: true},
		{name: "strict declared", child: child, contents: xsd.ProcessStrict},
	} {
		test := test
		t.Run(test.name, func(t *testing.T) {
			particle := xsd.Particle{Wildcard: &xsd.Wildcard{
				Namespaces: []string{"##any"}, ProcessContents: test.contents,
			}}
			before := len(state.diagnostics)
			next, matched, err := state.matchParticleOnce(particle, []*instanceNode{test.child}, 0, "urn:test", "/root")
			if !matched || next != 1 || err != nil || (len(state.diagnostics) > before) != test.wantErr {
				t.Fatalf("matchParticleOnce(%s) = %d, %t, %v", test.name, next, matched, err)
			}
		})
	}

	particle := xsd.Particle{
		Group:     &xsd.ModelGroup{Compositor: xsd.Sequence},
		MinOccurs: 0,
		MaxOccurs: 1,
	}
	next, matched, err := state.matchParticle(particle, nil, 0, "urn:test", "/root")
	if err != nil || !matched || next != 0 {
		t.Fatalf("matchParticle(nullable group) = %d, %t, %v", next, matched, err)
	}
}

func coverageLimits() Limits {
	return Limits{MaxDiagnostics: 20, MaxXPathSteps: 100, MaxIdentityValues: 100}
}

func requireDiagnostic(t *testing.T, state *validationState, operation func() error, name string) {
	t.Helper()

	before := len(state.diagnostics)
	if err := operation(); err != nil {
		t.Fatalf("%s error = %v", name, err)
	}
	if len(state.diagnostics) == before {
		t.Fatalf("%s produced no diagnostic", name)
	}
}

func advancedValidationSet(t *testing.T) *compile.Set {
	t.Helper()

	compiler, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	set, err := compiler.Compile(context.Background(), compile.Source{
		URI: "https://example.test/advanced.xsd",
		Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:t="urn:test" targetNamespace="urn:test">
 <xs:simpleType name="Union"><xs:union memberTypes="xs:string xs:int"/></xs:simpleType>
 <xs:complexType name="Base"/>
 <xs:complexType name="Abstract" abstract="true"/>
 <xs:complexType name="Other"/>
 <xs:element name="known" type="xs:string"/>
 <xs:attribute name="known" type="xs:string"/>
</xs:schema>`),
	})
	if err != nil {
		t.Fatal(err)
	}
	return set
}
