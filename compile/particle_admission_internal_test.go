package compile

import (
	"context"
	"errors"
	"reflect"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/resolve"
)

func TestExpandedParticleFinalCountAndCopyIsolation(t *testing.T) {
	for _, test := range []struct {
		second bool
		limit  int
		accept bool
	}{{false, 3, true}, {true, 5, true}, {false, 2, false}, {true, 4, false}} {
		const uri = "urn:particle-admission"
		extra := ""
		if test.second {
			extra = `<xs:complexType name="Second"><xs:group ref="t:Leaf"/></xs:complexType>`
		}
		source := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 xmlns:t="urn:test" targetNamespace="urn:test">
 <xs:group name="Leaf"><xs:sequence><xs:element name="value" type="xs:string"/></xs:sequence></xs:group>
 <xs:complexType name="First"><xs:group ref="t:Leaf"/></xs:complexType>` + extra + `</xs:schema>`)
		ctx := context.Background()
		document, err := xsd.Parse(ctx, source, xsd.ParseOptions{SystemID: uri})
		if err != nil {
			t.Fatal(err)
		}
		state := emptyValidationState()
		state.compiler, err = New(Options{Limits: Limits{MaxParticles: test.limit}})
		if err != nil {
			t.Fatal(err)
		}
		state.resources = map[string]resourceDocument{uri: {document: document}}
		state.instances = map[instanceKey]*Document{}
		if err := state.compileDocument(ctx, uri, document.TargetNamespace, 1); err != nil {
			t.Fatal(err)
		}
		first := xsd.QName{Namespace: "urn:test", Local: "First"}
		before := cloneComplexType(state.complexTypes[first])
		err = state.expandGroups()
		if !test.accept {
			_, err = state.compiler.Compile(ctx, Source{URI: uri, Content: source})
			if !errors.Is(err, ErrLimitExceeded) {
				t.Fatalf("expandGroups(limit %d, second %t) = %v; want admission error", test.limit, test.second, err)
			}
			continue
		}
		if err != nil {
			t.Fatalf("exact family limit %d: %v", test.limit, err)
		}
		if reflect.DeepEqual(before, state.complexTypes[first]) {
			t.Fatal("accepted reference was not expanded")
		}
		copy := state.complexTypes[first].Content.Particles[0].Group
		if copy == nil || copy.Particles[0].Element.Name != "value" {
			t.Fatal("accepted reference was not expanded")
		}
		copy.Particles[0].Element.Name = "changed"
		leaf := xsd.QName{Namespace: "urn:test", Local: "Leaf"}
		if state.modelGroups[leaf].Content.Particles[0].Element.Name != "value" {
			t.Fatal("expanded copy aliases retained definition")
		}
		if test.second {
			second := xsd.QName{Namespace: "urn:test", Local: "Second"}
			if state.complexTypes[second].Content.Particles[0].Group.Particles[0].Element.Name != "value" {
				t.Fatal("expanded references alias each other")
			}
		}
		if _, err := state.compiler.Compile(ctx, Source{URI: uri, Content: source}); err != nil {
			t.Fatalf("public exact family limit %d: %v", test.limit, err)
		}
	}
}

func TestFinalParticleLimitAllowsReducingRedefinition(t *testing.T) {
	resolver, err := resolve.NewMemory(map[string][]byte{
		"https://example.test/base.xsd": []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:test">
 <xs:complexType name="Value"><xs:sequence><xs:element name="first" type="xs:string" minOccurs="0"/></xs:sequence></xs:complexType>
</xs:schema>`),
		"https://example.test/middle.xsd": []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:t="urn:test" targetNamespace="urn:test">
 <xs:redefine schemaLocation="base.xsd"><xs:complexType name="Value"><xs:complexContent><xs:extension base="t:Value">
 <xs:sequence><xs:element name="second" type="xs:string" minOccurs="0"/></xs:sequence>
 </xs:extension></xs:complexContent></xs:complexType></xs:redefine>
</xs:schema>`),
	})
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := New(Options{Resolver: resolver, Limits: Limits{MaxParticles: 1, MaxParticleCopies: 100}})
	if err != nil {
		t.Fatal(err)
	}
	set, err := compiler.Compile(context.Background(), Source{
		URI: "https://example.test/root.xsd",
		Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:t="urn:test" targetNamespace="urn:test">
 <xs:redefine schemaLocation="middle.xsd"><xs:complexType name="Value"><xs:complexContent><xs:restriction base="t:Value">
 <xs:sequence/>
 </xs:restriction></xs:complexContent></xs:complexType></xs:redefine>
</xs:schema>`),
	})
	if err != nil {
		t.Fatal(err)
	}
	value, ok := set.ComplexType(xsd.QName{Namespace: "urn:test", Local: "Value"})
	if !ok || value.Content == nil || len(value.Content.Particles) != 0 {
		t.Fatalf("final reduced content = %#v", value)
	}
}

func TestParticleCopyAdmission(t *testing.T) {
	compiler, err := New(Options{Limits: Limits{MaxParticleCopies: 1}})
	if err != nil {
		t.Fatal(err)
	}
	state := &compileState{compiler: compiler, ctx: context.Background()}
	group := &xsd.ModelGroup{Compositor: xsd.Sequence, Particles: []xsd.Particle{
		{Element: &xsd.Element{Name: "first"}}, {Element: &xsd.Element{Name: "second"}},
	}}
	if copy := cloneModelGroup(group, state); copy != nil {
		t.Fatal("over-budget clone allocated and returned particle storage")
	}
	if !errors.Is(state.contextError(), ErrLimitExceeded) {
		t.Fatalf("owner error = %v", state.contextError())
	}
	if cloneModelGroup(&xsd.ModelGroup{}, state) != nil {
		t.Fatal("sticky rejection allowed later copy work")
	}
	ctx, cancel := context.WithCancel(context.Background())
	state.ctx = ctx
	cancel()
	if !errors.Is(state.contextError(), context.Canceled) {
		t.Fatal("copy rejection replaced real cancellation")
	}
}

func TestParticleCopiesCountWithoutCompiler(t *testing.T) {
	state := &compileState{ctx: context.Background()}
	group := &xsd.ModelGroup{Particles: []xsd.Particle{{Element: &xsd.Element{Name: "leaf"}}}}
	copy := cloneModelGroup(group, state)
	if copy == nil || len(copy.Particles) != 1 || state.particleCopies != 1 {
		t.Fatal("compiler-free owner did not copy and account for its particle")
	}
	copy.Particles[0].Element.Name = "changed"
	if group.Particles[0].Element.Name != "leaf" {
		t.Fatal("compiler-free owner copy aliases input")
	}
}

func TestParticleCopyBudgetCumulativeAndWrappers(t *testing.T) {
	compiler, err := New(Options{Limits: Limits{MaxParticleCopies: 2}})
	if err != nil {
		t.Fatal(err)
	}
	state := &compileState{compiler: compiler, ctx: context.Background()}
	empty := &xsd.ModelGroup{Compositor: xsd.Sequence}
	if group := extendContent(empty, empty, state); group == nil || len(group.Particles) != 2 {
		t.Fatal("exact wrapper budget rejected")
	}
	leaf := &xsd.ModelGroup{Particles: []xsd.Particle{{Element: &xsd.Element{Name: "leaf"}}}}
	if cloneModelGroup(leaf, state) != nil || !errors.Is(state.contextError(), ErrLimitExceeded) {
		t.Fatal("extension wrappers did not consume cumulative copy budget")
	}
	compiler, err = New(Options{Limits: Limits{MaxParticleCopies: 1}})
	if err != nil {
		t.Fatal(err)
	}
	state = &compileState{compiler: compiler, ctx: context.Background()}
	if extendContent(empty, empty, state) != nil || !errors.Is(state.contextError(), ErrLimitExceeded) {
		t.Fatal("over-budget wrappers were returned")
	}
}

func TestParticleCopyBudgetPublicControls(t *testing.T) {
	source := Source{URI: "urn:copies", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
 <xs:complexType name="Value"><xs:sequence><xs:element name="leaf" type="xs:string"/></xs:sequence></xs:complexType>
</xs:schema>`)}
	for _, limit := range []int{0, 2, 3} {
		compiler, err := New(Options{Limits: Limits{MaxParticles: 1, MaxParticleCopies: limit}})
		if err != nil {
			t.Fatal(err)
		}
		for repeat := 0; repeat < 2; repeat++ {
			set, err := compiler.Compile(context.Background(), source)
			if limit == 2 {
				if set != nil || !errors.Is(err, ErrLimitExceeded) {
					t.Fatalf("budget2 = %v, %v", set, err)
				}
			} else if set == nil || err != nil {
				t.Fatalf("budget%d = %v, %v", limit, set, err)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if set, err := compiler.Compile(ctx, source); set != nil || !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled budget%d = %v, %v", limit, set, err)
		}
	}
	if _, err := New(Options{Limits: Limits{MaxParticleCopies: -1}}); err == nil {
		t.Fatal("negative copy policy accepted")
	}
}
