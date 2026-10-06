package compile

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	xsd "github.com/faustbrian/go-xsd/v2"
)

func TestNullableModelGroupHonorsCompilationOwnerCancellation(t *testing.T) {
	for _, compositor := range []xsd.Compositor{xsd.Sequence, xsd.Choice} {
		group := &xsd.ModelGroup{
			Compositor: compositor,
			Particles: []xsd.Particle{{
				MinOccurs: 1, MaxOccurs: 1,
				Group: &xsd.ModelGroup{Compositor: xsd.Sequence, Particles: []xsd.Particle{{
					MinOccurs: 0, MaxOccurs: 1,
					Element: &xsd.Element{Name: "optional", Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}},
				}}},
			}},
		}
		ctx, cancel := context.WithCancel(context.Background())
		state := compileState{ctx: ctx}
		if !modelGroupNullable(group) || !modelGroupNullable(group, &state) {
			cancel()
			t.Fatal("ordinary nullable group was rejected")
		}
		cancel()
		if modelGroupNullable(group, &state) {
			t.Fatal("canceled compiler owner continued nullable-group assessment")
		}
		if !errors.Is(state.contextError(), context.Canceled) {
			t.Fatal("nullable-group assessment lost the real cancellation cause")
		}
		if !modelGroupNullable(group) {
			t.Fatal("owner cancellation changed context-free nullable-group semantics")
		}
	}
}

func TestCompileCancellationAfterGraphLoading(t *testing.T) {
	const uri = "urn:cancellation"
	const schema = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 xmlns:t="urn:test" targetNamespace="urn:test">
 <xs:group name="Leaf"><xs:sequence><xs:element name="value" type="xs:string"/></xs:sequence></xs:group>
 <xs:group name="Outer"><xs:sequence><xs:group ref="t:Leaf"/></xs:sequence></xs:group>
 <xs:element name="root" type="xs:string"/>
</xs:schema>`
	outer := xsd.QName{Namespace: "urn:test", Local: "Outer"}
	for _, canceled := range []bool{false, true} {
		ctx, cancel := context.WithCancel(context.Background())
		document, err := xsd.Parse(ctx, []byte(schema), xsd.ParseOptions{SystemID: uri})
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		state := emptyValidationState()
		state.compiler, err = New(Options{})
		if err != nil {
			cancel()
			t.Fatal(err)
		}
		state.resources = map[string]resourceDocument{uri: {document: document}}
		state.instances = map[instanceKey]*Document{}
		if err := state.compileDocument(ctx, uri, document.TargetNamespace, 1); err != nil {
			cancel()
			t.Fatal(err)
		}
		before := cloneModelGroup(state.modelGroups[outer].Content)
		if canceled {
			cancel()
		}
		err = state.expandGroups()
		cancel()
		if canceled {
			if !errors.Is(err, context.Canceled) {
				t.Fatalf("expansion after graph-loading cancellation = %v, want context.Canceled", err)
			}
			if !reflect.DeepEqual(before, state.modelGroups[outer].Content) {
				t.Fatal("canceled expansion changed indexed model group")
			}
			for _, phase := range []struct {
				name string
				run  func() error
			}{
				{"complex extensions", state.compileComplexExtensions},
				{"anonymous types", state.compileAnonymousComplexTypes},
				{"substitution", state.compileSubstitutions},
				{"components", state.validateComponents},
				{"identity", state.validateIdentityConstraints},
				{"UPA", func() error { return state.validateUniqueParticleAttribution(before, "urn:test") }},
				{"facets", func() error { return state.validateRestrictionFacets(xsd.SimpleType{}) }},
			} {
				if err := phase.run(); !errors.Is(err, context.Canceled) {
					t.Fatalf("%s after graph-loading cancellation = %v", phase.name, err)
				}
			}
		} else {
			if err != nil {
				t.Fatal(err)
			}
			particle := state.modelGroups[outer].Content.Particles[0]
			if particle.GroupRef.Local != "" || particle.Group == nil || len(particle.Group.Particles) != 1 {
				t.Fatalf("ordinary expanded group = %#v", particle)
			}
		}
	}
}

func TestCompileCancellationDoesNotPublishSetOrPoisonCompiler(t *testing.T) {
	compiler, err := New(Options{})
	if err != nil {
		t.Fatal(err)
	}
	source := Source{URI: "urn:ordinary", Content: []byte(`<xs:schema
 xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="root" type="xs:string"/></xs:schema>`)}
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, stop := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer stop()
	for _, test := range []struct {
		ctx context.Context
		err error
	}{{canceled, context.Canceled}, {expired, context.DeadlineExceeded}} {
		set, err := compiler.Compile(test.ctx, source)
		if set != nil || !errors.Is(err, test.err) {
			t.Fatalf("canceled Compile = %v, %v; want nil Set, %v", set, err, test.err)
		}
		// Cancellation is invocation-owned; the same immutable Compiler remains usable.
		set, err = compiler.Compile(context.Background(), source)
		if err != nil || set == nil || len(set.ElementNames()) != 1 {
			t.Fatalf("ordinary subsequent Compile = %v, %v", set, err)
		}
	}
}
