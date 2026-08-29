package compile

import (
	"context"
	"errors"
	"testing"

	xsd "github.com/faustbrian/go-xsd"
)

func TestSetLookupsReportMissingComponents(t *testing.T) {
	set := &Set{
		elements:        map[xsd.QName]xsd.Element{},
		attributes:      map[xsd.QName]xsd.Attribute{},
		simpleTypes:     map[xsd.QName]xsd.SimpleType{},
		complexTypes:    map[xsd.QName]xsd.ComplexType{},
		modelGroups:     map[xsd.QName]xsd.ModelGroupDefinition{},
		attributeGroups: map[xsd.QName]xsd.AttributeGroup{},
		notations:       map[xsd.QName]xsd.Notation{},
	}
	name := xsd.QName{Namespace: "urn:missing", Local: "Missing"}
	if _, ok := set.Element(name); ok {
		t.Fatal("Element(missing) succeeded")
	}
	if _, ok := set.Attribute(name); ok {
		t.Fatal("Attribute(missing) succeeded")
	}
	if _, ok := set.SimpleType(name); ok {
		t.Fatal("SimpleType(missing) succeeded")
	}
	if _, ok := set.ComplexType(name); ok {
		t.Fatal("ComplexType(missing) succeeded")
	}
}

func TestSubstitutionValidationCoversRecursiveTransitiveAndBlockedMembers(t *testing.T) {
	state := emptyValidationState()
	state.substitutionHeads = map[xsd.QName]xsd.QName{}
	head := xsd.QName{Namespace: "urn:test", Local: "Head"}
	member := xsd.QName{Namespace: "urn:test", Local: "Member"}
	state.elements[head] = xsd.Element{Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}
	state.elements[member] = xsd.Element{Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}
	state.substitutionHeads[member] = head
	if err := state.validateSubstitution(member, map[xsd.QName]uint8{}); err != nil {
		t.Fatalf("validateSubstitution(valid) error = %v", err)
	}
	if err := state.validateSubstitution(member, map[xsd.QName]uint8{member: 1}); err == nil {
		t.Fatal("validateSubstitution(recursive) succeeded")
	}
	if err := state.validateSubstitution(member, map[xsd.QName]uint8{member: 2}); err != nil {
		t.Fatalf("validateSubstitution(already complete) error = %v", err)
	}

	transitive := emptyValidationState()
	transitive.substitutionHeads = map[xsd.QName]xsd.QName{}
	a := xsd.QName{Namespace: "urn:test", Local: "A"}
	b := xsd.QName{Namespace: "urn:test", Local: "B"}
	transitive.substitutionHeads[a] = b
	transitive.substitutionHeads[b] = head
	transitive.elements[a] = xsd.Element{Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}
	transitive.elements[b] = xsd.Element{Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}
	transitive.elements[head] = state.elements[head]
	if err := transitive.validateSubstitution(a, map[xsd.QName]uint8{}); err != nil {
		t.Fatalf("validateSubstitution(transitive) error = %v", err)
	}

	blocked := emptyValidationState()
	blocked.substitutionHeads = map[xsd.QName]xsd.QName{}
	blocked.elements[head] = state.elements[head]
	blocked.elements[member] = state.elements[member]
	blocked.substitutionHeads[member] = head
	parsed, err := xsd.Parse(context.Background(), []byte(`<schema xmlns="`+xsd.Namespace+`" blockDefault="restriction"/>`), xsd.ParseOptions{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if !parsed.BlockDefault.Contains(xsd.DerivationRestriction) {
		t.Fatalf("parsed block default = %s", parsed.BlockDefault)
	}
	baseType := xsd.QName{Namespace: "urn:test", Local: "BaseType"}
	derivedType := xsd.QName{Namespace: "urn:test", Local: "DerivedType"}
	blocked.simpleTypes[derivedType] = xsd.SimpleType{
		Variety: xsd.SimpleRestriction,
		Base:    baseType,
	}
	blocked.elements[head] = xsd.Element{Type: baseType, Final: parsed.BlockDefault}
	blocked.elements[member] = xsd.Element{Type: derivedType}
	if !blocked.elements[head].Final.Contains(xsd.DerivationRestriction) {
		t.Fatalf("blocked head final = %s", blocked.elements[head].Final)
	}
	methods, derived := blocked.elementTypeDerivationMethods(blocked.elements[member], blocked.elements[head].Type)
	if !derived || len(methods) != 1 || methods[0] != xsd.DerivationRestriction {
		t.Fatalf("substitution methods = %v, %t", methods, derived)
	}
	if !blocked.elements[head].Final.Contains(methods[0]) {
		t.Fatalf("blocked head does not contain method %s", methods[0])
	}
	if err := blocked.validateSubstitution(member, map[xsd.QName]uint8{}); err == nil {
		t.Fatal("validateSubstitution(blocked) succeeded")
	}
}

func TestCompilerAndAnonymousTypeBranches(t *testing.T) {
	if got := compileChildDepth(1); got != 2 {
		t.Fatalf("compileChildDepth(1) = %d, want 2", got)
	}

	compiler, err := New(Options{})
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	if _, err := compiler.Compile(context.Background(), Source{URI: "https://example.test/root.xsd", Content: []byte(`<schema`)}); err == nil {
		t.Fatal("Compile(malformed) succeeded")
	}
	if _, err := compiler.Compile(context.Background(), Source{
		URI:     "https://example.test/root.xsd",
		Content: []byte(`<schema xmlns="` + xsd.Namespace + `"><element name="member" substitutionGroup="missing"/></schema>`),
	}); !errors.Is(err, ErrUnresolvedComponent) {
		t.Fatalf("Compile(missing substitution head) error = %v", err)
	}

	state := emptyValidationState()
	if err := state.expandAnonymousComplexType(&xsd.ComplexType{
		Content: &xsd.ModelGroup{Particles: []xsd.Particle{{GroupRef: xsd.QName{Namespace: "urn:test", Local: "Missing"}}}},
	}); err == nil {
		t.Fatal("expandAnonymousComplexType(missing group) succeeded")
	}
	state.simpleTypes[xsd.QName{Namespace: xsd.Namespace, Local: "string"}] = xsd.SimpleType{}
	if err := state.expandAnonymousComplexType(&xsd.ComplexType{
		SimpleContent: true,
		Derivation:    xsd.DerivationExtension,
		Base:          xsd.QName{Namespace: xsd.Namespace, Local: "string"},
	}); err != nil {
		t.Fatalf("expandAnonymousComplexType(simple extension) error = %v", err)
	}
	if err := state.expandAnonymousComplexType(&xsd.ComplexType{
		Derivation: xsd.DerivationExtension,
		Base:       xsd.QName{Namespace: "urn:test", Local: "Missing"},
	}); err == nil {
		t.Fatal("expandAnonymousComplexType(missing base) succeeded")
	}
	if err := state.expandAnonymousComplexType(&xsd.ComplexType{
		SimpleContent: true,
		Derivation:    xsd.DerivationExtension,
		Base:          xsd.QName{Namespace: "urn:test", Local: "Missing"},
	}); err == nil {
		t.Fatal("expandAnonymousComplexType(missing simple base) succeeded")
	}

	base := xsd.QName{Namespace: "urn:test", Local: "Base"}
	state.complexTypes[base] = xsd.ComplexType{}
	if err := state.expandAnonymousComplexType(&xsd.ComplexType{
		Derivation: xsd.DerivationRestriction,
		Base:       base,
	}); err != nil {
		t.Fatalf("expandAnonymousComplexType(restriction) error = %v", err)
	}
	if got := extendContent(nil, &xsd.ModelGroup{}); got == nil {
		t.Fatal("extendContent(nil, extension) returned nil")
	}
	if got := extendContent(&xsd.ModelGroup{}, nil); got == nil {
		t.Fatal("extendContent(base, nil) returned nil")
	}
}

func TestCompilerCoversSetAndDerivationBranches(t *testing.T) {
	parsed, err := xsd.Parse(context.Background(), []byte(`<schema xmlns="`+xsd.Namespace+`" blockDefault="restriction"/>`), xsd.ParseOptions{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	base := xsd.QName{Namespace: "urn:test", Local: "Base"}
	derived := xsd.QName{Namespace: "urn:test", Local: "Derived"}
	set := &Set{
		elements: map[xsd.QName]xsd.Element{
			base:    {Type: xsd.QName{Namespace: xsd.Namespace, Local: "anyType"}},
			derived: {Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}},
		},
		complexTypes: map[xsd.QName]xsd.ComplexType{},
		simpleTypes: map[xsd.QName]xsd.SimpleType{derived: {
			Variety: xsd.SimpleRestriction,
			Base:    xsd.QName{Namespace: xsd.Namespace, Local: "anySimpleType"},
		}},
	}
	set.elements[base] = xsd.Element{Type: xsd.QName{Namespace: xsd.Namespace, Local: "anyType"}, Block: parsed.BlockDefault}
	set.elements[derived] = xsd.Element{Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}
	set.substitutionHeads = map[xsd.QName]xsd.QName{derived: base}
	if _, ok := set.SubstitutionMember(base, xsd.QName{Namespace: "urn:test", Local: "Missing"}); ok {
		t.Fatal("SubstitutionMember(missing) succeeded")
	}
	blockedSet := &Set{
		elements: map[xsd.QName]xsd.Element{
			base:    {Type: base, Block: parsed.BlockDefault},
			derived: {Type: derived},
		},
		simpleTypes:       map[xsd.QName]xsd.SimpleType{derived: {Variety: xsd.SimpleRestriction, Base: base}},
		complexTypes:      map[xsd.QName]xsd.ComplexType{},
		substitutionHeads: map[xsd.QName]xsd.QName{derived: base},
	}
	if _, ok := blockedSet.SubstitutionMember(base, derived); ok {
		t.Fatal("SubstitutionMember(blocked) succeeded")
	}

	state := emptyValidationState()
	state.substitutionHeads = map[xsd.QName]xsd.QName{}
	state.elements[base] = xsd.Element{Type: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}
	state.elements[derived] = xsd.Element{}
	state.substitutionHeads[derived] = base
	if err := state.validateSubstitution(derived, map[xsd.QName]uint8{}); err != nil {
		t.Fatalf("validateSubstitution(inherit type) error = %v", err)
	}

	state = emptyValidationState()
	state.simpleTypes[xsd.QName{Namespace: xsd.Namespace, Local: "string"}] = xsd.SimpleType{}
	complexName := xsd.QName{Namespace: "urn:test", Local: "Text"}
	state.complexTypes[complexName] = xsd.ComplexType{
		SimpleContent: true,
		Derivation:    xsd.DerivationExtension,
		Base:          xsd.QName{Namespace: xsd.Namespace, Local: "string"},
	}
	if err := state.compileComplexType(complexName, map[xsd.QName]uint8{}); err != nil {
		t.Fatalf("compileComplexType(simple extension) error = %v", err)
	}
	baseText := xsd.QName{Namespace: "urn:test", Local: "BaseText"}
	derivedText := xsd.QName{Namespace: "urn:test", Local: "DerivedText"}
	state.complexTypes[baseText] = xsd.ComplexType{SimpleContent: true, SimpleBase: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}
	state.complexTypes[derivedText] = xsd.ComplexType{SimpleContent: true, Derivation: xsd.DerivationExtension, Base: baseText}
	if err := state.compileComplexType(derivedText, map[xsd.QName]uint8{}); err != nil {
		t.Fatalf("compileComplexType(inherited simple extension) error = %v", err)
	}

	baseName := xsd.QName{Namespace: "urn:test", Local: "BaseComplex"}
	derivedName := xsd.QName{Namespace: "urn:test", Local: "DerivedComplex"}
	state.complexTypes[baseName] = xsd.ComplexType{AttributeWildcard: &xsd.Wildcard{Namespaces: []string{"##other"}}}
	state.complexTypes[derivedName] = xsd.ComplexType{
		Derivation: xsd.DerivationExtension,
		Base:       baseName,
	}
	if err := state.compileComplexType(derivedName, map[xsd.QName]uint8{}); err != nil {
		t.Fatalf("compileComplexType(wildcard inheritance) error = %v", err)
	}

	if err := state.applySimpleContentDerivation(
		&xsd.ComplexType{Derivation: xsd.DerivationRestriction, InlineSimpleType: &xsd.SimpleType{Variety: xsd.SimpleRestriction, Base: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}},
		xsd.ComplexType{},
	); err == nil {
		t.Fatal("applySimpleContentDerivation(non-mixed base) succeeded")
	}
	if err := state.compileSimpleContentValueType(
		&xsd.ComplexType{Derivation: xsd.DerivationRestriction, InlineSimpleType: &xsd.SimpleType{Variety: xsd.SimpleRestriction, Base: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}},
		nil,
	); err != nil {
		t.Fatalf("compileSimpleContentValueType(inline base) error = %v", err)
	}
}

func TestCompilerCoversValidationAndIndexingBranches(t *testing.T) {
	state := emptyValidationState()
	if err := state.validateComponents(); err != nil {
		t.Fatalf("validateComponents(empty) error = %v", err)
	}

	state = emptyValidationState()
	state.attributes[xsd.QName{Namespace: "urn:test", Local: "missing"}] = xsd.Attribute{Type: xsd.QName{Namespace: "urn:test", Local: "Missing"}}
	if err := state.validateComponents(); !errors.Is(err, ErrUnresolvedComponent) {
		t.Fatalf("validateComponents(missing attribute type) error = %v", err)
	}
	state = emptyValidationState()
	state.simpleTypes[xsd.QName{Namespace: "urn:test", Local: "bad"}] = xsd.SimpleType{Name: "not a name"}
	if err := state.validateComponents(); !errors.Is(err, ErrInvalidComponent) {
		t.Fatalf("validateComponents(invalid simple name) error = %v", err)
	}

	id := xsd.QName{Namespace: xsd.Namespace, Local: "ID"}
	state = emptyValidationState()
	state.typeKinds[id] = "simple"
	if err := state.validateAttributeUseSet([]xsd.AttributeUse{{Name: "first", Type: id}, {Name: "second", Type: id}}); err == nil {
		t.Fatal("validateAttributeUseSet(multiple IDs) succeeded")
	}
	if err := state.validateAttributeUseSet([]xsd.AttributeUse{
		{Name: "inline-first", InlineSimpleType: &xsd.SimpleType{Variety: xsd.SimpleRestriction, Base: id}},
		{Name: "inline-second", InlineSimpleType: &xsd.SimpleType{Variety: xsd.SimpleRestriction, Base: id}},
	}); err == nil {
		t.Fatal("validateAttributeUseSet(multiple inline IDs) succeeded")
	}
	if err := state.validateSimpleTypeDefinition(xsd.SimpleType{Variety: xsd.SimpleRestriction, Base: xsd.QName{Namespace: "urn:test", Local: "Missing"}}); err == nil {
		t.Fatal("validateSimpleTypeDefinition(missing base) succeeded")
	}
	state.attributes[xsd.QName{Namespace: "urn:test", Local: "both"}] = xsd.Attribute{DefaultSet: true, FixedSet: true}
	if err := state.validateComponents(); err == nil {
		t.Fatal("validateComponents(attribute default and fixed) succeeded")
	}

	state = emptyValidationState()
	particles := 0
	if err := state.validateModelGroup(&xsd.ModelGroup{Compositor: xsd.All, Particles: []xsd.Particle{{Wildcard: &xsd.Wildcard{}}}}, "", &particles); err == nil {
		t.Fatal("validateModelGroup(all wildcard) succeeded")
	}
	if err := state.validateModelGroup(&xsd.ModelGroup{Particles: []xsd.Particle{{Element: &xsd.Element{}}}}, "", &particles); err == nil {
		t.Fatal("validateModelGroup(unnamed element) succeeded")
	}
	if err := state.validateModelGroup(&xsd.ModelGroup{Particles: []xsd.Particle{{Element: &xsd.Element{Ref: xsd.QName{Namespace: "urn:test", Local: "missing"}}}}}, "", &particles); err == nil {
		t.Fatal("validateModelGroup(missing element ref) succeeded")
	}

	if err := state.validateElementValueConstraint(xsd.Element{
		DefaultSet: true,
		Default:    "invalid",
		InlineComplexType: &xsd.ComplexType{
			SimpleContent:    true,
			InlineSimpleType: &xsd.SimpleType{Variety: xsd.SimpleRestriction, Base: xsd.QName{Namespace: xsd.Namespace, Local: "int"}},
		},
	}); err == nil {
		t.Fatal("validateElementValueConstraint(invalid inline complex value) succeeded")
	}
	if err := state.validateElementValueConstraint(xsd.Element{
		DefaultSet: true,
		Default:    "value",
		InlineComplexType: &xsd.ComplexType{
			SimpleContent:    true,
			InlineSimpleType: &xsd.SimpleType{Variety: xsd.SimpleRestriction, Base: xsd.QName{Namespace: xsd.Namespace, Local: "string"}},
		},
	}); err != nil {
		t.Fatalf("validateElementValueConstraint(valid inline complex value) error = %v", err)
	}
	complexValue := xsd.QName{Namespace: "urn:test", Local: "Value"}
	state.complexTypes[complexValue] = xsd.ComplexType{
		SimpleContent:    true,
		InlineSimpleType: &xsd.SimpleType{Variety: xsd.SimpleRestriction, Base: xsd.QName{Namespace: xsd.Namespace, Local: "string"}},
	}
	if err := state.validateElementValueConstraint(xsd.Element{Type: complexValue, DefaultSet: true, Default: "value"}); err != nil {
		t.Fatalf("validateElementValueConstraint(named inline complex value) error = %v", err)
	}
	if err := state.validateAttributeDeclarationValueConstraint(xsd.Attribute{FixedSet: true, Fixed: "value"}); err != nil {
		t.Fatalf("validateAttributeDeclarationValueConstraint(fixed) error = %v", err)
	}

	state = emptyValidationState()
	name := xsd.QName{Namespace: "urn:test", Local: "Name"}
	state.modelGroups[name] = xsd.ModelGroupDefinition{}
	state.simpleTypes[name] = xsd.SimpleType{Base: xsd.QName{Namespace: xsd.Namespace, Local: "string"}}
	document := &xsd.Document{
		SimpleTypes: []xsd.SimpleType{{Name: "Name", Variety: xsd.SimpleRestriction, Base: name}},
		ModelGroups: []xsd.ModelGroupDefinition{{Name: "Name", Content: &xsd.ModelGroup{}}},
	}
	if err := state.applyRedefinition(xsd.Redefinition{SimpleTypes: document.SimpleTypes, ModelGroups: document.ModelGroups}, document, name.Namespace, false); err != nil {
		t.Fatalf("applyRedefinition(simple and model) error = %v", err)
	}
}

func TestCompilerCoversNormalizationBranches(t *testing.T) {
	name := xsd.QName{Local: "Name"}
	group := xsd.ModelGroup{Particles: []xsd.Particle{{GroupRef: name}}}
	if err := (&compileState{compiler: &Compiler{limits: Limits{MaxComponents: 100}}, elements: map[xsd.QName]xsd.Element{}, attributes: map[xsd.QName]xsd.Attribute{}, simpleTypes: map[xsd.QName]xsd.SimpleType{}, complexTypes: map[xsd.QName]xsd.ComplexType{}, modelGroups: map[xsd.QName]xsd.ModelGroupDefinition{}, attributeGroups: map[xsd.QName]xsd.AttributeGroup{}, notations: map[xsd.QName]xsd.Notation{}, typeKinds: map[xsd.QName]string{}}).indexComponents(&xsd.Document{
		Elements: []xsd.Element{{Name: "same"}, {Name: "same"}},
	}, "urn:test", false); err == nil {
		t.Fatal("indexComponents(duplicate elements) succeeded")
	}
	if err := (&compileState{compiler: &Compiler{limits: Limits{MaxComponents: 100}}, elements: map[xsd.QName]xsd.Element{}, attributes: map[xsd.QName]xsd.Attribute{}, simpleTypes: map[xsd.QName]xsd.SimpleType{}, complexTypes: map[xsd.QName]xsd.ComplexType{}, modelGroups: map[xsd.QName]xsd.ModelGroupDefinition{}, attributeGroups: map[xsd.QName]xsd.AttributeGroup{}, notations: map[xsd.QName]xsd.Notation{}, typeKinds: map[xsd.QName]string{}}).indexComponents(&xsd.Document{
		Attributes: []xsd.Attribute{{Name: "code", Type: name}},
	}, "urn:target", true); err != nil {
		t.Fatalf("indexComponents(chameleon attribute) error = %v", err)
	}
	if err := (&compileState{compiler: &Compiler{limits: Limits{MaxComponents: 100}}, elements: map[xsd.QName]xsd.Element{}, attributes: map[xsd.QName]xsd.Attribute{}, simpleTypes: map[xsd.QName]xsd.SimpleType{}, complexTypes: map[xsd.QName]xsd.ComplexType{}, modelGroups: map[xsd.QName]xsd.ModelGroupDefinition{}, attributeGroups: map[xsd.QName]xsd.AttributeGroup{}, notations: map[xsd.QName]xsd.Notation{}, typeKinds: map[xsd.QName]string{}}).indexComponents(&xsd.Document{
		SimpleTypes: []xsd.SimpleType{{Name: "same"}, {Name: "same"}},
	}, "urn:test", false); err == nil {
		t.Fatal("indexComponents(duplicate simple types) succeeded")
	}
	if err := (&compileState{compiler: &Compiler{limits: Limits{MaxComponents: 100}}, elements: map[xsd.QName]xsd.Element{}, attributes: map[xsd.QName]xsd.Attribute{}, simpleTypes: map[xsd.QName]xsd.SimpleType{}, complexTypes: map[xsd.QName]xsd.ComplexType{}, modelGroups: map[xsd.QName]xsd.ModelGroupDefinition{}, attributeGroups: map[xsd.QName]xsd.AttributeGroup{}, notations: map[xsd.QName]xsd.Notation{}, typeKinds: map[xsd.QName]string{}}).indexComponents(&xsd.Document{
		ModelGroups: []xsd.ModelGroupDefinition{{Name: "same"}, {Name: "same"}},
	}, "urn:test", false); err == nil {
		t.Fatal("indexComponents(duplicate model groups) succeeded")
	}

	if got := normalizeComplexType(xsd.ComplexType{AttributeGroupRefs: []xsd.QName{name}, Base: name}, xsd.FormQualified, xsd.FormQualified, "urn:target", true); got.Base.Namespace != "urn:target" || got.AttributeGroupRefs[0].Namespace != "urn:target" {
		t.Fatalf("normalizeComplexType(chameleon) = %#v", got)
	}
	if got := normalizeElementTypes(xsd.Element{InlineSimpleType: &xsd.SimpleType{InlineBase: &xsd.SimpleType{}, InlineItem: &xsd.SimpleType{}, InlineMembers: []xsd.SimpleType{{}}}}, xsd.FormQualified, xsd.FormQualified, "urn:target", true); got.InlineSimpleType == nil {
		t.Fatal("normalizeElementTypes(inline simple) dropped type")
	}
	if got := normalizeInlineSimpleType(xsd.SimpleType{InlineBase: &xsd.SimpleType{}, InlineItem: &xsd.SimpleType{}, InlineMembers: []xsd.SimpleType{{}}, Base: name, ItemType: name, MemberTypes: []xsd.QName{name}}, "urn:target", true); got.Base.Namespace != "urn:target" || got.ItemType.Namespace != "urn:target" || got.MemberTypes[0].Namespace != "urn:target" {
		t.Fatalf("normalizeInlineSimpleType(chameleon) = %#v", got)
	}
	if got := normalizeAttributeUses([]xsd.AttributeUse{{Form: xsd.FormQualified, InlineSimpleType: &xsd.SimpleType{}}}, xsd.FormUnqualified, "urn:target", false); got[0].Namespace != "urn:target" {
		t.Fatalf("normalizeAttributeUses(qualified) = %#v", got)
	}
	if got := normalizeAttributeUses([]xsd.AttributeUse{{InlineSimpleType: &xsd.SimpleType{}}}, xsd.FormQualified, "urn:target", false); got[0].InlineSimpleType == nil {
		t.Fatal("normalizeAttributeUses(inline simple) dropped type")
	}
	if group.Particles[0].GroupRef != name {
		t.Fatal("normalization fixture changed unexpectedly")
	}
	if got := normalizeConstraintWhitespace("a\tb\nc", "replace"); got != "a b c" {
		t.Fatalf("normalizeConstraintWhitespace(replace) = %q", got)
	}
	if _, ok := notationFacetName(xsd.Facet{Value: "Name", Namespaces: map[string]string{"": "urn:test"}}); !ok {
		t.Fatal("notationFacetName(unqualified) failed")
	}
}
