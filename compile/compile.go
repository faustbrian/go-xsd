package compile

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"sort"
	"strings"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/datatype"
	"github.com/faustbrian/go-xsd/v2/internal/errsafe"
	"github.com/faustbrian/go-xsd/v2/resolve"
)

var (
	ErrLimitExceeded       = errors.New("xsd compile: resource limit exceeded")
	ErrNamespace           = errors.New("xsd compile: namespace mismatch")
	ErrResourceIdentity    = errors.New("xsd compile: resource identity mismatch")
	ErrDuplicateComponent  = errors.New("xsd compile: duplicate component")
	ErrInvalidComponent    = errors.New("xsd compile: invalid component")
	ErrUnresolvedComponent = errors.New("xsd compile: unresolved component")
)

const (
	defaultMaxSchemas        = 256
	defaultMaxDepth          = 64
	defaultMaxReferences     = 4096
	defaultMaxBytes          = 64 << 20
	defaultMaxURIBytes       = 64 << 10
	defaultMaxComponents     = 100000
	defaultMaxParticles      = 1000000
	defaultMaxParticleCopies = 1000000
)

func compileDepthExceeded(depth int) bool {
	return depth > defaultMaxDepth
}

func compileChildDepth(depth int) int {
	return depth + 1
}

// Limits bounds graph construction. Zero values select conservative defaults.
type Limits struct {
	MaxSchemas    int
	MaxDepth      int
	MaxReferences int
	MaxBytes      int64
	// MaxURIBytes bounds each root, reference and resolved resource URI before
	// identity parsing, cache lookup or resolver dispatch. Zero selects 64 KiB.
	// It is independent of MaxBytes, which counts cumulative schema content.
	MaxURIBytes   int64
	MaxComponents int
	MaxParticles  int
	// MaxParticleCopies bounds cumulative compiler-owned particle slots copied or
	// synthesized per Compile call, including temporary and redefined content.
	// Zero selects 1,000,000 independently of MaxParticles (the final count).
	MaxParticleCopies int
	// MaxParseNamespaceEntries and MaxParseModelBytes independently bound owned
	// parsing work per document, not cumulatively across the schema graph.
	// Zero selects parser defaults (1,000,000 entries and 64 MiB).
	MaxParseNamespaceEntries int
	MaxParseModelBytes       int64
}

// Options configures a Compiler. A nil Resolver denies every external load.
type Options struct {
	Resolver resolve.Resolver
	Limits   Limits
}

// Source is the caller-owned root schema resource.
type Source struct {
	URI     string
	Content []byte
}

// Compiler is immutable and safe for concurrent use.
type Compiler struct {
	resolver resolve.Resolver
	limits   Limits
}

// New validates options and creates a reusable compiler.
func New(options Options) (*Compiler, error) {
	limits := options.Limits
	if err := validateLimits(limits); err != nil {
		return nil, err
	}
	if limits.MaxSchemas == 0 {
		limits.MaxSchemas = defaultMaxSchemas
	}
	if limits.MaxDepth == 0 {
		limits.MaxDepth = defaultMaxDepth
	}
	if limits.MaxReferences == 0 {
		limits.MaxReferences = defaultMaxReferences
	}
	if limits.MaxBytes == 0 {
		limits.MaxBytes = defaultMaxBytes
	}
	if limits.MaxURIBytes == 0 {
		limits.MaxURIBytes = defaultMaxURIBytes
	}
	if limits.MaxComponents == 0 {
		limits.MaxComponents = defaultMaxComponents
	}
	if limits.MaxParticles == 0 {
		limits.MaxParticles = defaultMaxParticles
	}
	if limits.MaxParticleCopies == 0 {
		limits.MaxParticleCopies = defaultMaxParticleCopies
	}
	resolver := options.Resolver
	if resolver == nil {
		resolver = resolve.Deny()
	}
	return &Compiler{resolver: resolver, limits: limits}, nil
}

func validateLimits(limits Limits) error {
	if limits.MaxURIBytes < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	if limits.MaxParseNamespaceEntries < 0 || limits.MaxParseModelBytes < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	if limits.MaxSchemas < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	if limits.MaxDepth < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	if limits.MaxReferences < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	if limits.MaxBytes < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	if limits.MaxComponents < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	if limits.MaxParticles < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	if limits.MaxParticleCopies < 0 {
		return fmt.Errorf("xsd compile: limits must not be negative")
	}
	return nil
}

// Document describes one schema document in one effective namespace.
type Document struct {
	URI          string
	Namespace    string
	Chameleon    bool
	Dependencies []string
}

// Set is an immutable, concurrency-safe schema document graph.
type Set struct {
	documents         []Document
	elements          map[xsd.QName]xsd.Element
	attributes        map[xsd.QName]xsd.Attribute
	simpleTypes       map[xsd.QName]xsd.SimpleType
	complexTypes      map[xsd.QName]xsd.ComplexType
	modelGroups       map[xsd.QName]xsd.ModelGroupDefinition
	attributeGroups   map[xsd.QName]xsd.AttributeGroup
	notations         map[xsd.QName]xsd.Notation
	substitutionHeads map[xsd.QName]xsd.QName
}

// ElementNames returns global element names in expanded-name order.
func (s *Set) ElementNames() []xsd.QName { return sortedComponentNames(s.elements) }

// AttributeNames returns global attribute names in expanded-name order.
func (s *Set) AttributeNames() []xsd.QName { return sortedComponentNames(s.attributes) }

// SimpleTypeNames returns global simple type names in expanded-name order.
func (s *Set) SimpleTypeNames() []xsd.QName { return sortedComponentNames(s.simpleTypes) }

// ComplexTypeNames returns global complex type names in expanded-name order.
func (s *Set) ComplexTypeNames() []xsd.QName { return sortedComponentNames(s.complexTypes) }

// ModelGroupNames returns global model group names in expanded-name order.
func (s *Set) ModelGroupNames() []xsd.QName { return sortedComponentNames(s.modelGroups) }

// AttributeGroupNames returns global attribute group names in expanded-name order.
func (s *Set) AttributeGroupNames() []xsd.QName { return sortedComponentNames(s.attributeGroups) }

// NotationNames returns global notation names in expanded-name order.
func (s *Set) NotationNames() []xsd.QName { return sortedComponentNames(s.notations) }

func sortedComponentNames[T any](components map[xsd.QName]T, owner ...*compileState) []xsd.QName {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	names := make([]xsd.QName, 0, len(components))
	for name := range components {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		names = append(names, name)
	}
	sort.Slice(names, func(left, right int) bool {
		return expandedNameLess(names[left], names[right])
	})
	return names
}

func expandedNameLess(left, right xsd.QName) bool {
	if comparison := strings.Compare(left.Namespace, right.Namespace); comparison != 0 {
		return comparison == -1
	}
	return strings.Compare(left.Local, right.Local) == -1
}

// ModelGroup returns a named model group by expanded name.
func (s *Set) ModelGroup(name xsd.QName) (xsd.ModelGroupDefinition, bool) {
	group, ok := s.modelGroups[name]
	if !ok {
		return xsd.ModelGroupDefinition{}, false
	}
	group.Content = cloneModelGroup(group.Content)
	group.Annotation = cloneAnnotation(group.Annotation)
	return group, true
}

// AttributeGroup returns a named attribute group by expanded name.
func (s *Set) AttributeGroup(name xsd.QName) (xsd.AttributeGroup, bool) {
	group, ok := s.attributeGroups[name]
	if !ok {
		return xsd.AttributeGroup{}, false
	}
	return cloneAttributeGroup(group), true
}

// Notation returns a global notation declaration by expanded name.
func (s *Set) Notation(name xsd.QName) (xsd.Notation, bool) {
	notation, ok := s.notations[name]
	if !ok {
		return xsd.Notation{}, false
	}
	notation.Annotation = cloneAnnotation(notation.Annotation)
	return notation, true
}

// SubstitutionMember resolves member when it may substitute for head.
func (s *Set) SubstitutionMember(head xsd.QName, member xsd.QName) (xsd.Element, bool) {
	return substitutionMember(s, head, member)
}

func substitutionMember(s *Set, head xsd.QName, member xsd.QName, owner ...*compileState) (xsd.Element, bool) {
	if compileOwnerError(owner) != nil {
		return xsd.Element{}, false
	}
	headDeclaration, ok := s.elements[head]
	if !ok || headDeclaration.Block.Contains(xsd.DerivationSubstitution) {
		return xsd.Element{}, false
	}
	current := member
	for current.Local != "" {
		if compileOwnerError(owner) != nil {
			return xsd.Element{}, false
		}
		direct, affiliated := s.substitutionHeads[current]
		if !affiliated {
			return xsd.Element{}, false
		}
		if direct == head {
			declaration, exists := s.elements[member]
			if !exists {
				return xsd.Element{}, false
			}
			methods, derived := setElementTypeDerivationMethods(s, declaration, headDeclaration.Type, owner...)
			if !derived {
				return xsd.Element{}, false
			}
			var typeBlock xsd.DerivationSet
			if headType, exists := s.complexTypes[headDeclaration.Type]; exists {
				typeBlock = headType.Block
			}
			for _, method := range methods {
				if compileOwnerError(owner) != nil {
					return xsd.Element{}, false
				}
				if headDeclaration.Block.Contains(method) || typeBlock.Contains(method) {
					return xsd.Element{}, false
				}
			}
			return cloneElement(declaration, owner...), true
		}
		directDeclaration, exists := s.elements[direct]
		if !exists || directDeclaration.Block.Contains(xsd.DerivationSubstitution) {
			return xsd.Element{}, false
		}
		current = direct
	}
	return xsd.Element{}, false
}

func setElementTypeDerivationMethods(
	s *Set,
	element xsd.Element,
	base xsd.QName,

	owner ...*compileState,
) ([]xsd.Derivation, bool) {
	if err := compileOwnerError(owner); err != nil {
		return nil, false
	}

	if element.InlineComplexType != nil {
		rest, ok := setTypeDerivationMethods(s, element.InlineComplexType.Base, base, owner...)
		return append([]xsd.Derivation{element.InlineComplexType.Derivation}, rest...), ok
	}
	if element.InlineSimpleType != nil {
		rest, ok := setTypeDerivationMethods(s, element.InlineSimpleType.Base, base, owner...)
		return append([]xsd.Derivation{xsd.DerivationRestriction}, rest...), ok
	}
	return setTypeDerivationMethods(s, element.Type, base, owner...)
}

func setTypeDerivationMethods(
	s *Set,
	derived xsd.QName,
	base xsd.QName,

	owner ...*compileState,
) ([]xsd.Derivation, bool) {
	if err := compileOwnerError(owner); err != nil {
		return nil, false
	}

	state := compileState{simpleTypes: s.simpleTypes, complexTypes: s.complexTypes}
	if len(owner) != 0 {
		state.ctx = owner[0].ctx
	}
	return state.typeDerivationMethods(derived, base)
}

func cloneAttributeGroup(group xsd.AttributeGroup, owner ...*compileState) xsd.AttributeGroup {
	if err := compileOwnerError(owner); err != nil {
		return xsd.AttributeGroup{}
	}

	group.Attributes = cloneAttributeUses(group.Attributes, owner...)
	group.References = append([]xsd.QName(nil), group.References...)
	group.AttributeGroupReferences = append(
		[]xsd.AttributeGroupReference(nil),
		group.AttributeGroupReferences...,
	)
	for index := range group.AttributeGroupReferences {
		if err := compileOwnerError(owner); err != nil {
			return xsd.AttributeGroup{}
		}
		group.AttributeGroupReferences[index].Annotation = cloneAnnotation(
			group.AttributeGroupReferences[index].Annotation,
			owner...)
	}
	group.Wildcard = cloneWildcard(group.Wildcard, owner...)
	group.Annotation = cloneAnnotation(group.Annotation, owner...)
	return group
}

func cloneWildcard(wildcard *xsd.Wildcard, owner ...*compileState) *xsd.Wildcard {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	if wildcard == nil {
		return nil
	}
	clone := *wildcard
	clone.Namespaces = append([]string(nil), wildcard.Namespaces...)
	clone.Annotation = cloneAnnotation(wildcard.Annotation, owner...)
	return &clone
}

// Documents returns a deep copy in deterministic URI and namespace order.
func (s *Set) Documents() []Document {
	result := make([]Document, len(s.documents))
	for index, document := range s.documents {
		result[index] = cloneDocument(document)
	}
	return result
}

// Document returns the first compiled document with the resource URI. A
// chameleon resource compiled into multiple namespaces is returned in sorted
// namespace order and remains visible in full through Documents.
func (s *Set) Document(uri string) (Document, bool) {
	for _, document := range s.documents {
		if document.URI == uri {
			return cloneDocument(document), true
		}
	}
	return Document{}, false
}

// Element returns a global element declaration by expanded name.
func (s *Set) Element(name xsd.QName) (xsd.Element, bool) {
	element, ok := s.elements[name]
	if !ok {
		return xsd.Element{}, false
	}
	return cloneElement(element), true
}

func cloneElement(element xsd.Element, owner ...*compileState) xsd.Element {
	if err := compileOwnerError(owner); err != nil {
		return xsd.Element{}
	}

	element.Annotation = cloneAnnotation(element.Annotation, owner...)
	element.ValueNamespaces = cloneNamespaceMap(element.ValueNamespaces, owner...)
	constraints := element.IdentityConstraints
	element.IdentityConstraints = make(
		[]xsd.IdentityConstraint,
		len(constraints),
	)
	for index, constraint := range constraints {
		if err := compileOwnerError(owner); err != nil {
			return xsd.Element{}
		}
		constraint.Fields = append([]string(nil), constraint.Fields...)
		constraint.FieldIDs = append([]string(nil), constraint.FieldIDs...)
		constraint.Namespaces = cloneNamespaceMap(constraint.Namespaces, owner...)
		constraint.Annotation = cloneAnnotation(constraint.Annotation, owner...)
		constraint.SelectorAnnotation = cloneAnnotation(constraint.SelectorAnnotation, owner...)
		constraint.FieldAnnotations = append(
			[]*xsd.Annotation(nil),
			constraint.FieldAnnotations...,
		)
		for fieldIndex := range constraint.FieldAnnotations {
			if err := compileOwnerError(owner); err != nil {
				return xsd.Element{}
			}
			constraint.FieldAnnotations[fieldIndex] = cloneAnnotation(
				constraint.FieldAnnotations[fieldIndex],
				owner...)
		}
		element.IdentityConstraints[index] = constraint
	}
	if element.InlineSimpleType != nil {
		typeDefinition := cloneSimpleType(*element.InlineSimpleType, owner...)
		element.InlineSimpleType = &typeDefinition
	}
	if element.InlineComplexType != nil {
		typeDefinition := cloneComplexType(*element.InlineComplexType, owner...)
		element.InlineComplexType = &typeDefinition
	}
	return element
}

func cloneNamespaceMap(namespaces map[string]string, owner ...*compileState) map[string]string {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	clone := make(map[string]string, len(namespaces))
	for prefix, namespace := range namespaces {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		clone[prefix] = namespace
	}
	return clone
}

// Attribute returns a global attribute declaration by expanded name.
func (s *Set) Attribute(name xsd.QName) (xsd.Attribute, bool) {
	attribute, ok := s.attributes[name]
	if !ok {
		return xsd.Attribute{}, false
	}
	return cloneAttribute(attribute), true
}

func cloneAttribute(attribute xsd.Attribute, owner ...*compileState) xsd.Attribute {
	if err := compileOwnerError(owner); err != nil {
		return xsd.Attribute{}
	}

	attribute.Annotation = cloneAnnotation(attribute.Annotation, owner...)
	attribute.ValueNamespaces = cloneNamespaceMap(attribute.ValueNamespaces, owner...)
	if attribute.InlineSimpleType != nil {
		typeDefinition := cloneSimpleType(*attribute.InlineSimpleType, owner...)
		attribute.InlineSimpleType = &typeDefinition
	}
	return attribute
}

func cloneAttributeUses(attributes []xsd.AttributeUse, owner ...*compileState) []xsd.AttributeUse {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	clone := make([]xsd.AttributeUse, len(attributes))
	for index, attribute := range attributes {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		clone[index] = attribute
		clone[index].ValueNamespaces = cloneNamespaceMap(attribute.ValueNamespaces, owner...)
		clone[index].Annotation = cloneAnnotation(attribute.Annotation, owner...)
		if attribute.InlineSimpleType != nil {
			typeDefinition := cloneSimpleType(*attribute.InlineSimpleType, owner...)
			clone[index].InlineSimpleType = &typeDefinition
		}
	}
	return clone
}

// SimpleType returns a global simple type definition by expanded name.
func (s *Set) SimpleType(name xsd.QName) (xsd.SimpleType, bool) {
	simpleType, ok := s.simpleTypes[name]
	if !ok {
		return xsd.SimpleType{}, false
	}
	return cloneSimpleType(simpleType), true
}

func cloneSimpleType(simpleType xsd.SimpleType, owner ...*compileState) xsd.SimpleType {
	if err := compileOwnerError(owner); err != nil {
		return xsd.SimpleType{}
	}

	simpleType.Annotation = cloneAnnotation(simpleType.Annotation, owner...)
	simpleType.VarietyAnnotation = cloneAnnotation(simpleType.VarietyAnnotation, owner...)
	simpleType.Facets = append([]xsd.Facet(nil), simpleType.Facets...)
	for index := range simpleType.Facets {
		if err := compileOwnerError(owner); err != nil {
			return xsd.SimpleType{}
		}
		simpleType.Facets[index].Namespaces = cloneNamespaceMap(
			simpleType.Facets[index].Namespaces,
			owner...)
		simpleType.Facets[index].Annotation = cloneAnnotation(
			simpleType.Facets[index].Annotation,
			owner...)
	}
	simpleType.MemberTypes = append([]xsd.QName(nil), simpleType.MemberTypes...)
	if simpleType.InlineBase != nil {
		base := cloneSimpleType(*simpleType.InlineBase, owner...)
		simpleType.InlineBase = &base
	}
	if simpleType.InlineItem != nil {
		item := cloneSimpleType(*simpleType.InlineItem, owner...)
		simpleType.InlineItem = &item
	}
	simpleType.InlineMembers = append([]xsd.SimpleType(nil), simpleType.InlineMembers...)
	for index := range simpleType.InlineMembers {
		if err := compileOwnerError(owner); err != nil {
			return xsd.SimpleType{}
		}
		simpleType.InlineMembers[index] = cloneSimpleType(simpleType.InlineMembers[index], owner...)
	}
	return simpleType
}

// ComplexType returns a global complex type definition by expanded name.
func (s *Set) ComplexType(name xsd.QName) (xsd.ComplexType, bool) {
	complexType, ok := s.complexTypes[name]
	if !ok {
		return xsd.ComplexType{}, false
	}
	return cloneComplexType(complexType), true
}

func cloneComplexType(complexType xsd.ComplexType, owner ...*compileState) xsd.ComplexType {
	if err := compileOwnerError(owner); err != nil {
		return xsd.ComplexType{}
	}

	complexType.Annotation = cloneAnnotation(complexType.Annotation, owner...)
	complexType.ContentAnnotation = cloneAnnotation(complexType.ContentAnnotation, owner...)
	complexType.DerivationAnnotation = cloneAnnotation(complexType.DerivationAnnotation, owner...)
	if complexType.InlineSimpleType != nil {
		typeDefinition := cloneSimpleType(*complexType.InlineSimpleType, owner...)
		complexType.InlineSimpleType = &typeDefinition
	}
	complexType.SimpleFacets = append([]xsd.Facet(nil), complexType.SimpleFacets...)
	for index := range complexType.SimpleFacets {
		if err := compileOwnerError(owner); err != nil {
			return xsd.ComplexType{}
		}
		complexType.SimpleFacets[index].Namespaces = cloneNamespaceMap(
			complexType.SimpleFacets[index].Namespaces,
			owner...)
		complexType.SimpleFacets[index].Annotation = cloneAnnotation(
			complexType.SimpleFacets[index].Annotation,
			owner...)
	}
	complexType.Attributes = cloneAttributeUses(complexType.Attributes, owner...)
	complexType.Content = cloneModelGroup(complexType.Content, owner...)
	complexType.AttributeGroupRefs = append(
		[]xsd.QName(nil),
		complexType.AttributeGroupRefs...,
	)
	complexType.AttributeGroupReferences = append(
		[]xsd.AttributeGroupReference(nil),
		complexType.AttributeGroupReferences...,
	)
	for index := range complexType.AttributeGroupReferences {
		if err := compileOwnerError(owner); err != nil {
			return xsd.ComplexType{}
		}
		complexType.AttributeGroupReferences[index].Annotation = cloneAnnotation(
			complexType.AttributeGroupReferences[index].Annotation,
			owner...)
	}
	complexType.AttributeWildcard = cloneWildcard(complexType.AttributeWildcard, owner...)
	return complexType
}

func cloneModelGroup(group *xsd.ModelGroup, owner ...*compileState) *xsd.ModelGroup {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	if group == nil {
		return nil
	}
	clone := &xsd.ModelGroup{
		Compositor: group.Compositor,
		MinOccurs:  group.MinOccurs,
		MaxOccurs:  group.MaxOccurs,
		Unbounded:  group.Unbounded,
		OccursSet:  group.OccursSet,
		Annotation: cloneAnnotation(group.Annotation, owner...),
	}
	if !admitParticleCopies(len(group.Particles), owner) {
		return nil
	}
	clone.Particles = make([]xsd.Particle, len(group.Particles))
	for index, particle := range group.Particles {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		clone.Particles[index] = particle
		clone.Particles[index].Annotation = cloneAnnotation(particle.Annotation, owner...)
		if particle.Element != nil {
			element := cloneElement(*particle.Element, owner...)
			clone.Particles[index].Element = &element
		}
		clone.Particles[index].Group = cloneModelGroup(particle.Group, owner...)
		clone.Particles[index].Wildcard = cloneWildcard(particle.Wildcard, owner...)
	}
	return clone
}

func cloneAnnotation(annotation *xsd.Annotation, owner ...*compileState) *xsd.Annotation {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	if annotation == nil {
		return nil
	}
	clone := *annotation
	clone.Documentation = append([]xsd.Documentation(nil), annotation.Documentation...)
	clone.AppInformation = append([]xsd.AppInfo(nil), annotation.AppInformation...)
	return &clone
}

func cloneDocument(document Document, owner ...*compileState) Document {
	if err := compileOwnerError(owner); err != nil {
		return Document{}
	}

	document.Dependencies = append([]string(nil), document.Dependencies...)
	return document
}

// Compile parses and resolves a complete bounded schema graph.
func (c *Compiler) Compile(ctx context.Context, root Source) (set *Set, err error) {
	defer func() {
		if err != nil {
			set = nil
			err = errsafe.Wrap("xsd compile: failed", err)
		}
	}()
	defer func() {
		if ctx != nil {
			if canceled := ctx.Err(); canceled != nil {
				set, err = nil, canceled
			}
		}
	}()
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if err := c.admitURI(root.URI); err != nil {
		return nil, err
	}
	if err := validateIdentity(root.URI); err != nil {
		return nil, err
	}
	if int64(len(root.Content)) > c.limits.MaxBytes {
		return nil, fmt.Errorf("%w: schema bytes exceed %d", ErrLimitExceeded, c.limits.MaxBytes)
	}
	state := compileState{
		ctx:               ctx,
		compiler:          c,
		resources:         map[string]resourceDocument{},
		instances:         map[instanceKey]*Document{},
		elements:          map[xsd.QName]xsd.Element{},
		attributes:        map[xsd.QName]xsd.Attribute{},
		simpleTypes:       map[xsd.QName]xsd.SimpleType{},
		complexTypes:      map[xsd.QName]xsd.ComplexType{},
		modelGroups:       map[xsd.QName]xsd.ModelGroupDefinition{},
		attributeGroups:   map[xsd.QName]xsd.AttributeGroup{},
		notations:         map[xsd.QName]xsd.Notation{},
		substitutionHeads: map[xsd.QName]xsd.QName{},
		typeKinds:         map[xsd.QName]string{},
		bytes:             int64(len(root.Content)),
	}
	defer func() {
		if stopped := state.contextError(); stopped != nil {
			set, err = nil, stopped
		}
	}()
	document, err := xsd.Parse(ctx, root.Content, xsd.ParseOptions{
		SystemID:            root.URI,
		MaxDocumentBytes:    c.limits.MaxBytes,
		MaxNamespaceEntries: c.limits.MaxParseNamespaceEntries,
		MaxModelBytes:       c.limits.MaxParseModelBytes,
	})
	if err != nil {
		return nil, err
	}
	state.resources[root.URI] = resourceDocument{document: document}
	if err := state.compileDocument(ctx, root.URI, document.TargetNamespace, 1); err != nil {
		return nil, err
	}
	if err := state.expandGroups(); err != nil {
		return nil, err
	}
	if err := state.compileComplexExtensions(); err != nil {
		return nil, err
	}
	if err := state.compileAnonymousComplexTypes(); err != nil {
		return nil, err
	}
	if err := state.compileSubstitutions(); err != nil {
		return nil, err
	}
	if err := state.validateComponents(); err != nil {
		return nil, err
	}

	documents := make([]Document, 0, len(state.instances))
	for _, document := range state.instances {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		documents = append(documents, cloneDocument(*document, &state))
	}
	sort.Slice(documents, func(left, right int) bool {
		return documentLess(documents[left], documents[right])
	})
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &Set{
		documents:         documents,
		elements:          state.elements,
		attributes:        state.attributes,
		simpleTypes:       state.simpleTypes,
		complexTypes:      state.complexTypes,
		modelGroups:       state.modelGroups,
		attributeGroups:   state.attributeGroups,
		notations:         state.notations,
		substitutionHeads: state.substitutionHeads,
	}, nil
}

func documentLess(left, right Document) bool {
	if comparison := strings.Compare(left.URI, right.URI); comparison != 0 {
		return comparison == -1
	}
	return strings.Compare(left.Namespace, right.Namespace) == -1
}

var builtInTypes = map[string]string{
	"anyType":       "complex",
	"anySimpleType": "simple",
	"string":        "simple", "boolean": "simple", "decimal": "simple",
	"float": "simple", "double": "simple", "duration": "simple",
	"dateTime": "simple", "time": "simple", "date": "simple",
	"gYearMonth": "simple", "gYear": "simple", "gMonthDay": "simple",
	"gDay": "simple", "gMonth": "simple", "hexBinary": "simple",
	"base64Binary": "simple", "anyURI": "simple", "QName": "simple",
	"NOTATION": "simple", "normalizedString": "simple", "token": "simple",
	"language": "simple", "Name": "simple", "NCName": "simple",
	"ID": "simple", "IDREF": "simple", "IDREFS": "simple",
	"ENTITY": "simple", "ENTITIES": "simple", "NMTOKEN": "simple",
	"NMTOKENS": "simple", "integer": "simple", "nonPositiveInteger": "simple",
	"negativeInteger": "simple", "long": "simple", "int": "simple",
	"short": "simple", "byte": "simple", "nonNegativeInteger": "simple",
	"unsignedLong": "simple", "unsignedInt": "simple", "unsignedShort": "simple",
	"unsignedByte": "simple", "positiveInteger": "simple",
}

func (s *compileState) compileSubstitutions() error {
	if err := s.contextError(); err != nil {
		return err
	}

	for member, declaration := range s.elements {
		if err := s.contextError(); err != nil {
			return err
		}
		if declaration.SubstitutionGroup.Local == "" {
			continue
		}
		if _, ok := s.elements[declaration.SubstitutionGroup]; !ok {
			return unresolvedComponent("substitution group head", declaration.SubstitutionGroup)
		}
		s.substitutionHeads[member] = declaration.SubstitutionGroup
	}
	colors := make(map[xsd.QName]uint8, len(s.substitutionHeads))
	for member := range s.substitutionHeads {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := s.validateSubstitution(member, colors); err != nil {
			return err
		}
	}
	return nil
}

func (s *compileState) validateSubstitution(
	member xsd.QName,
	colors map[xsd.QName]uint8,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	switch colors[member] {
	case 1:
		return invalidComponent("element", member, "substitution affiliation is recursive")
	case 2:
		return nil
	}
	colors[member] = 1
	head := s.substitutionHeads[member]
	if _, transitive := s.substitutionHeads[head]; transitive {
		return s.validateSubstitution(head, colors)
	}
	memberDeclaration := s.elements[member]
	headDeclaration := s.elements[head]
	if memberDeclaration.Type.Local == "" && memberDeclaration.InlineSimpleType == nil &&
		memberDeclaration.InlineComplexType == nil {
		memberDeclaration.Type = headDeclaration.Type
		memberDeclaration.InlineSimpleType = headDeclaration.InlineSimpleType
		memberDeclaration.InlineComplexType = headDeclaration.InlineComplexType
		s.elements[member] = memberDeclaration
	}
	headType := headDeclaration.Type
	methods, derived := s.elementTypeDerivationMethods(memberDeclaration, headType)
	if !derived {
		return invalidComponent(
			"element",
			member,
			"type is not validly derived from its substitution group head",
		)
	}
	for _, method := range methods {
		if err := s.contextError(); err != nil {
			return err
		}
		if headDeclaration.Final.Contains(method) {
			return invalidComponent("element", member, "head excludes the type derivation method")
		}
	}
	colors[member] = 2
	return nil
}

func (s *compileState) elementTypeDerivationMethods(
	element xsd.Element,
	head xsd.QName,
) ([]xsd.Derivation, bool) {
	if err := s.contextError(); err != nil {
		return nil, false
	}

	if element.InlineComplexType != nil {
		methods := []xsd.Derivation{element.InlineComplexType.Derivation}
		rest, ok := s.typeDerivationMethods(element.InlineComplexType.Base, head)
		return append(methods, rest...), ok
	}
	if element.InlineSimpleType != nil {
		base := element.InlineSimpleType.Base
		if element.InlineSimpleType.InlineBase != nil {
			base = element.InlineSimpleType.InlineBase.Base
		}
		rest, ok := s.typeDerivationMethods(base, head)
		return append([]xsd.Derivation{xsd.DerivationRestriction}, rest...), ok
	}
	return s.typeDerivationMethods(element.Type, head)
}

func (s *compileState) typeDerivationMethods(
	derived xsd.QName,
	base xsd.QName,
) ([]xsd.Derivation, bool) {
	if err := s.contextError(); err != nil {
		return nil, false
	}

	if base.Local == "" || base == (xsd.QName{Namespace: xsd.Namespace, Local: "anyType"}) {
		return nil, true
	}
	methods := make([]xsd.Derivation, 0)
	seen := make(map[xsd.QName]struct{})
	for derived != base {
		if err := s.contextError(); err != nil {
			return nil, false
		}
		if _, duplicate := seen[derived]; duplicate {
			return nil, false
		}
		seen[derived] = struct{}{}
		if complexType, ok := s.complexTypes[derived]; ok {
			methods = append(methods, complexType.Derivation)
			derived = complexType.Base
			continue
		}
		if simpleType, ok := s.simpleTypes[derived]; ok {
			method := xsd.Derivation(simpleType.Variety)
			next := simpleType.Base
			if simpleType.Variety == xsd.SimpleList ||
				simpleType.Variety == xsd.SimpleUnion {
				next = xsd.QName{Namespace: xsd.Namespace, Local: "anySimpleType"}
			}
			methods = append(methods, method)
			derived = next
			continue
		}
		if derived.Namespace == xsd.Namespace {
			parent, method, ok := datatype.BuiltInDerivation(derived.Local)
			if !ok {
				return nil, false
			}
			methods = append(methods, xsd.Derivation(method))
			derived = xsd.QName{Namespace: xsd.Namespace, Local: parent}
			continue
		}
		return nil, false
	}
	return methods, true
}

func (s *compileState) expandGroups() error {
	if err := s.contextError(); err != nil {
		return err
	}

	modelColors := make(map[xsd.QName]uint8, len(s.modelGroups))
	for name := range s.modelGroups {
		if err := s.contextError(); err != nil {
			return err
		}
		if _, err := s.expandModelGroup(name, modelColors); err != nil {
			return err
		}
	}
	attributeColors := make(map[xsd.QName]uint8, len(s.attributeGroups))
	for name := range s.attributeGroups {
		if err := s.contextError(); err != nil {
			return err
		}
		if _, err := s.expandAttributeGroup(name, attributeColors); err != nil {
			return err
		}
	}
	for name, typeDefinition := range s.complexTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		content, err := s.expandModelGroupContent(typeDefinition.Content, modelColors)
		if err != nil {
			return err
		}
		typeDefinition.Content = content
		wildcardSeen := typeDefinition.AttributeWildcard != nil
		for _, reference := range typeDefinition.AttributeGroupRefs {
			if err := s.contextError(); err != nil {
				return err
			}
			group, err := s.expandAttributeGroup(reference, attributeColors)
			if err != nil {
				return err
			}
			typeDefinition.Attributes = append(typeDefinition.Attributes, group.Attributes...)
			if group.Wildcard != nil {
				if wildcardSeen {
					typeDefinition.AttributeWildcard = intersectWildcards(
						typeDefinition.AttributeWildcard,
						group.Wildcard,
						s)
				} else {
					typeDefinition.AttributeWildcard = cloneWildcard(group.Wildcard, s)
					wildcardSeen = true
				}
			}
		}
		typeDefinition.AttributeGroupRefs = nil
		s.complexTypes[name] = typeDefinition
	}
	return nil
}

func (s *compileState) expandModelGroup(
	name xsd.QName,
	colors map[xsd.QName]uint8,
) (xsd.ModelGroupDefinition, error) {
	if err := s.contextError(); err != nil {
		return xsd.ModelGroupDefinition{}, err
	}

	switch colors[name] {
	case 1:
		return xsd.ModelGroupDefinition{}, invalidComponent(
			"model group",
			name,
			"references are recursive",
		)
	case 2:
		return s.modelGroups[name], nil
	}
	group, ok := s.modelGroups[name]
	if !ok {
		return xsd.ModelGroupDefinition{}, unresolvedComponent("model group", name)
	}
	colors[name] = 1
	content, err := s.expandModelGroupContent(group.Content, colors)
	if err != nil {
		return xsd.ModelGroupDefinition{}, err
	}
	group.Content = content
	s.modelGroups[name] = group
	colors[name] = 2
	return group, nil
}

func (s *compileState) expandModelGroupContent(
	content *xsd.ModelGroup,
	colors map[xsd.QName]uint8,
) (*xsd.ModelGroup, error) {
	if err := s.contextError(); err != nil {
		return nil, err
	}

	content = cloneModelGroup(content, s)
	if content == nil {
		return nil, nil
	}
	for index := range content.Particles {
		if err := s.contextError(); err != nil {
			return nil, err
		}
		particle := &content.Particles[index]
		if particle.GroupRef.Local != "" {
			definition, err := s.expandModelGroup(particle.GroupRef, colors)
			if err != nil {
				return nil, err
			}
			particle.Group = cloneModelGroup(definition.Content, s)
			particle.GroupRef = xsd.QName{}
		} else {
			nested, err := s.expandModelGroupContent(particle.Group, colors)
			if err != nil {
				return nil, err
			}
			particle.Group = nested
		}
	}
	return content, nil
}

func (s *compileState) expandAttributeGroup(
	name xsd.QName,
	colors map[xsd.QName]uint8,
) (xsd.AttributeGroup, error) {
	if err := s.contextError(); err != nil {
		return xsd.AttributeGroup{}, err
	}

	switch colors[name] {
	case 1:
		return xsd.AttributeGroup{}, invalidComponent(
			"attribute group",
			name,
			"references are recursive",
		)
	case 2:
		return s.attributeGroups[name], nil
	}
	group, ok := s.attributeGroups[name]
	if !ok {
		return xsd.AttributeGroup{}, unresolvedComponent("attribute group", name)
	}
	colors[name] = 1
	attributes := append([]xsd.AttributeUse(nil), group.Attributes...)
	wildcardSeen := group.Wildcard != nil
	for _, reference := range group.References {
		if err := s.contextError(); err != nil {
			return xsd.AttributeGroup{}, err
		}
		referenced, err := s.expandAttributeGroup(reference, colors)
		if err != nil {
			return xsd.AttributeGroup{}, err
		}
		attributes = append(attributes, referenced.Attributes...)
		if referenced.Wildcard != nil {
			if wildcardSeen {
				group.Wildcard = intersectWildcards(group.Wildcard, referenced.Wildcard, s)
			} else {
				group.Wildcard = cloneWildcard(referenced.Wildcard, s)
				wildcardSeen = true
			}
		}
	}
	group.Attributes = attributes
	group.References = nil
	s.attributeGroups[name] = group
	colors[name] = 2
	return group, nil
}

func (s *compileState) compileComplexExtensions() error {
	if err := s.contextError(); err != nil {
		return err
	}

	colors := make(map[xsd.QName]uint8, len(s.complexTypes))
	for name := range s.complexTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := s.compileComplexType(name, colors); err != nil {
			return err
		}
	}
	return nil
}

func (s *compileState) compileAnonymousComplexTypes() error {
	if err := s.contextError(); err != nil {
		return err
	}

	for name, element := range s.elements {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := s.compileElementAnonymousType(&element, name.Namespace); err != nil {
			return err
		}
		s.elements[name] = element
	}
	for name, typeDefinition := range s.complexTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := s.compileAnonymousTypesInGroup(typeDefinition.Content, name.Namespace); err != nil {
			return err
		}
		s.complexTypes[name] = typeDefinition
	}
	return nil
}

func (s *compileState) compileElementAnonymousType(element *xsd.Element, namespace string) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if element.InlineComplexType == nil {
		return nil
	}
	typeDefinition := cloneComplexType(*element.InlineComplexType, s)
	if err := s.expandAnonymousComplexType(&typeDefinition); err != nil {
		return err
	}
	if err := s.compileAnonymousTypesInGroup(typeDefinition.Content, namespace); err != nil {
		return err
	}
	element.InlineComplexType = &typeDefinition
	return nil
}

func (s *compileState) expandAnonymousComplexType(typeDefinition *xsd.ComplexType) error {
	if err := s.contextError(); err != nil {
		return err
	}

	modelColors := make(map[xsd.QName]uint8, len(s.modelGroups))
	content, err := s.expandModelGroupContent(typeDefinition.Content, modelColors)
	if err != nil {
		return err
	}
	typeDefinition.Content = content
	attributeColors := make(map[xsd.QName]uint8, len(s.attributeGroups))
	for _, reference := range typeDefinition.AttributeGroupRefs {
		if err := s.contextError(); err != nil {
			return err
		}
		group, expandErr := s.expandAttributeGroup(reference, attributeColors)
		if expandErr != nil {
			return expandErr
		}
		typeDefinition.Attributes = append(typeDefinition.Attributes, group.Attributes...)
		if group.Wildcard != nil {
			switch {
			case typeDefinition.AttributeWildcard != nil:
				typeDefinition.AttributeWildcard = intersectWildcards(
					typeDefinition.AttributeWildcard,
					group.Wildcard,
					s)
			default:
				typeDefinition.AttributeWildcard = cloneWildcard(group.Wildcard, s)
			}
		}
	}
	typeDefinition.AttributeGroupRefs = nil

	if typeDefinition.Derivation == "" {
		return nil
	}
	if typeDefinition.Base.Local == "" {
		return fmt.Errorf("%w: anonymous complex type derivation has no base", ErrInvalidComponent)
	}
	if typeDefinition.Derivation != xsd.DerivationExtension &&
		typeDefinition.Derivation != xsd.DerivationRestriction {
		return fmt.Errorf("%w: anonymous complex type has an invalid derivation method", ErrInvalidComponent)
	}
	if typeDefinition.SimpleContent {
		if s.typeExists(typeDefinition.Base, "simple") {
			if typeDefinition.Derivation != xsd.DerivationExtension {
				return fmt.Errorf(
					"%w: anonymous complex type: simple-content derivation from a simple type must be extension",
					ErrInvalidComponent,
				)
			}
			typeDefinition.SimpleBase = typeDefinition.Base
			if err := s.compileSimpleContentValueType(typeDefinition, nil); err != nil {
				return fmt.Errorf("%w: anonymous complex type: %s", ErrInvalidComponent, err)
			}
			return nil
		}
		base, ok := s.complexTypes[typeDefinition.Base]
		if !ok {
			return unresolvedComponent("simple content base", typeDefinition.Base)
		}
		return s.applySimpleContentDerivation(typeDefinition, base)
	}
	if typeDefinition.Base == (xsd.QName{Namespace: xsd.Namespace, Local: "anyType"}) {
		return nil
	}
	base, ok := s.complexTypes[typeDefinition.Base]
	if !ok {
		return unresolvedComponent("complex type", typeDefinition.Base)
	}
	if base.Final.Contains(typeDefinition.Derivation) {
		return fmt.Errorf("%w: anonymous base type prohibits this derivation", ErrInvalidComponent)
	}
	if typeDefinition.Derivation == xsd.DerivationRestriction {
		if err := s.validateComplexRestriction(*typeDefinition, base); err != nil {
			return fmt.Errorf("%w: anonymous complex type: %s", ErrInvalidComponent, err)
		}
		typeDefinition.Attributes = restrictedAttributes(
			base.Attributes,
			typeDefinition.Attributes,
			s)
		return nil
	}
	if !typeDefinition.MixedSet {
		typeDefinition.Mixed = base.Mixed
	} else if typeDefinition.Mixed != base.Mixed {
		return fmt.Errorf("%w: anonymous extension changes the base mixed-content policy", ErrInvalidComponent)
	}
	typeDefinition.Content = extendContent(base.Content, typeDefinition.Content, s)
	typeDefinition.Attributes = append(
		append([]xsd.AttributeUse(nil), base.Attributes...),
		typeDefinition.Attributes...,
	)
	switch {
	case typeDefinition.AttributeWildcard == nil:
		typeDefinition.AttributeWildcard = cloneWildcard(base.AttributeWildcard, s)
	case base.AttributeWildcard != nil:
		typeDefinition.AttributeWildcard = unionWildcards(
			base.AttributeWildcard,
			typeDefinition.AttributeWildcard,
			s)
	}
	return nil
}

func (s *compileState) compileAnonymousTypesInGroup(group *xsd.ModelGroup, namespace string) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if group == nil {
		return nil
	}
	for index := range group.Particles {
		if err := s.contextError(); err != nil {
			return err
		}
		particle := &group.Particles[index]
		if particle.Element != nil {
			if err := s.compileElementAnonymousType(particle.Element, namespace); err != nil {
				return err
			}
		}
		if err := s.compileAnonymousTypesInGroup(particle.Group, namespace); err != nil {
			return err
		}
	}
	return nil
}

func (s *compileState) compileComplexType(
	name xsd.QName,
	colors map[xsd.QName]uint8,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	switch colors[name] {
	case 1:
		return invalidComponent("complex type", name, "derivation is recursive")
	case 2:
		return nil
	}
	colors[name] = 1
	typeDefinition := cloneComplexType(s.complexTypes[name], s)
	if typeDefinition.Derivation == "" {
		colors[name] = 2
		return nil
	}
	if typeDefinition.Base.Local == "" {
		return invalidComponent("complex type", name, "derivation has no base")
	}
	if typeDefinition.Derivation != xsd.DerivationExtension &&
		typeDefinition.Derivation != xsd.DerivationRestriction {
		return invalidComponent("complex type", name, "has an invalid derivation method")
	}
	if typeDefinition.SimpleContent {
		if s.typeExists(typeDefinition.Base, "simple") {
			if typeDefinition.Derivation != xsd.DerivationExtension {
				return invalidComponent(
					"complex type",
					name,
					"simple-content derivation from a simple type must be extension",
				)
			}
			typeDefinition.SimpleBase = typeDefinition.Base
			if err := s.compileSimpleContentValueType(&typeDefinition, nil); err != nil {
				return invalidComponent("complex type", name, err.Error())
			}
			s.complexTypes[name] = typeDefinition
			colors[name] = 2
			return nil
		}
		_, ok := s.complexTypes[typeDefinition.Base]
		if !ok {
			return unresolvedComponent("simple content base", typeDefinition.Base)
		}
		if err := s.compileComplexType(typeDefinition.Base, colors); err != nil {
			return err
		}
		base := s.complexTypes[typeDefinition.Base]
		if err := s.applySimpleContentDerivation(&typeDefinition, base); err != nil {
			return invalidComponent("complex type", name, err.Error())
		}
		s.complexTypes[name] = typeDefinition
		colors[name] = 2
		return nil
	}
	if typeDefinition.Base == (xsd.QName{Namespace: xsd.Namespace, Local: "anyType"}) {
		colors[name] = 2
		return nil
	}
	if _, ok := s.complexTypes[typeDefinition.Base]; !ok {
		return unresolvedComponent("complex type", typeDefinition.Base)
	}
	if err := s.compileComplexType(typeDefinition.Base, colors); err != nil {
		return err
	}
	base := s.complexTypes[typeDefinition.Base]
	if base.Final.Contains(typeDefinition.Derivation) {
		return invalidComponent("complex type", name, "base type prohibits this derivation")
	}
	if typeDefinition.Derivation == xsd.DerivationRestriction {
		if err := s.validateComplexRestriction(typeDefinition, base); err != nil {
			return invalidComponent("complex type", name, err.Error())
		}
		typeDefinition.Attributes = restrictedAttributes(
			base.Attributes,
			typeDefinition.Attributes,
			s)
		s.complexTypes[name] = typeDefinition
		colors[name] = 2
		return nil
	}
	if !typeDefinition.MixedSet {
		typeDefinition.Mixed = base.Mixed
	} else if typeDefinition.Mixed != base.Mixed {
		return invalidComponent(
			"complex type",
			name,
			"extension changes the base mixed-content policy",
		)
	}
	typeDefinition.Content = extendContent(base.Content, typeDefinition.Content, s)
	typeDefinition.Attributes = append(
		append([]xsd.AttributeUse(nil), base.Attributes...),
		typeDefinition.Attributes...,
	)
	if typeDefinition.AttributeWildcard == nil {
		typeDefinition.AttributeWildcard = cloneWildcard(base.AttributeWildcard, s)
	} else if base.AttributeWildcard != nil {
		typeDefinition.AttributeWildcard = unionWildcards(
			base.AttributeWildcard,
			typeDefinition.AttributeWildcard,
			s)
	}
	s.complexTypes[name] = typeDefinition
	colors[name] = 2
	return nil
}

func (s *compileState) applySimpleContentDerivation(
	derived *xsd.ComplexType,
	base xsd.ComplexType,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if base.Final.Contains(derived.Derivation) {
		return errors.New("base type prohibits this derivation")
	}
	var inherited *xsd.SimpleType
	if base.SimpleContent || base.SimpleBase.Local != "" ||
		s.typeExists(base.Base, "simple") {
		derived.SimpleBase = base.SimpleBase
		if derived.SimpleBase.Local == "" {
			derived.SimpleBase = base.Base
		}
		inherited = base.InlineSimpleType
		if derived.Derivation == xsd.DerivationRestriction &&
			derived.InlineSimpleType != nil &&
			(inherited != nil || !s.inlineSimpleTypeDerivesFrom(
				*derived.InlineSimpleType,
				derived.SimpleBase,
			)) {
			return errors.New(
				"inline simple content type is not derived from its base content type",
			)
		}
	} else {
		if derived.Derivation != xsd.DerivationRestriction {
			return errors.New(
				"simple content base must have simple content or emptiable mixed content",
			)
		}
		if !base.Mixed {
			return errors.New(
				"simple content base must have simple content or emptiable mixed content",
			)
		}
		if !modelGroupNullable(base.Content, s) {
			return errors.New(
				"simple content base must have simple content or emptiable mixed content",
			)
		}
		if derived.InlineSimpleType == nil {
			return errors.New(
				"simple content base must have simple content or emptiable mixed content",
			)
		}
	}
	if err := s.compileSimpleContentValueType(derived, inherited); err != nil {
		return err
	}
	if derived.Derivation == xsd.DerivationRestriction {
		if !s.attributesRestrictContext(
			derived.Attributes,
			base.Attributes,
			base.AttributeWildcard,
			derived.Base.Namespace,
		) {
			return errors.New("attribute uses are not a valid restriction of their base")
		}
		if derived.AttributeWildcard != nil {
			if base.AttributeWildcard == nil {
				return errors.New("attribute wildcard is not a valid restriction of its base")
			}
			if !wildcardRestricts(derived.AttributeWildcard, base.AttributeWildcard, s) {
				return errors.New("attribute wildcard is not a valid restriction of its base")
			}
		}
		derived.Attributes = restrictedAttributes(base.Attributes, derived.Attributes, s)
		return nil
	}
	derived.Attributes = append(
		append([]xsd.AttributeUse(nil), base.Attributes...),
		derived.Attributes...,
	)
	if derived.AttributeWildcard == nil {
		derived.AttributeWildcard = cloneWildcard(base.AttributeWildcard, s)
	} else if base.AttributeWildcard != nil {
		derived.AttributeWildcard = unionWildcards(base.AttributeWildcard, derived.AttributeWildcard, s)
	}
	return nil
}

func (s *compileState) compileSimpleContentValueType(
	derived *xsd.ComplexType,
	inherited *xsd.SimpleType,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if derived.Derivation == xsd.DerivationExtension {
		if derived.InlineSimpleType != nil || len(derived.SimpleFacets) != 0 {
			return errors.New("simple-content extension cannot declare facets")
		}
		if inherited != nil {
			typeDefinition := cloneSimpleType(*inherited, s)
			derived.InlineSimpleType = &typeDefinition
		}
		return nil
	}
	restriction := xsd.SimpleType{
		Variety: xsd.SimpleRestriction,
		Facets:  append([]xsd.Facet(nil), derived.SimpleFacets...),
	}
	if derived.InlineSimpleType != nil {
		base := cloneSimpleType(*derived.InlineSimpleType, s)
		restriction.InlineBase = &base
	} else if inherited != nil {
		base := cloneSimpleType(*inherited, s)
		restriction.InlineBase = &base
	} else {
		restriction.Base = derived.SimpleBase
	}
	if err := s.validateSimpleTypeDefinition(restriction); err != nil {
		return fmt.Errorf("simple-content restriction: %w", err)
	}
	derived.InlineSimpleType = &restriction
	derived.SimpleFacets = nil
	return nil
}

func (s *compileState) validateComplexRestriction(derived, base xsd.ComplexType) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if derived.Mixed && !base.Mixed {
		return errors.New("restriction enables mixed content")
	}
	if !s.modelGroupRestricts(derived.Content, base.Content) {
		return errors.New("content model is not a valid restriction of its base")
	}
	if !s.attributesRestrictContext(
		derived.Attributes,
		base.Attributes,
		base.AttributeWildcard,
		derived.Base.Namespace,
	) {
		return errors.New("attribute uses are not a valid restriction of their base")
	}
	if derived.AttributeWildcard != nil &&
		(base.AttributeWildcard == nil || !wildcardRestricts(derived.AttributeWildcard, base.AttributeWildcard, s)) {
		return errors.New("attribute wildcard is not a valid restriction of its base")
	}
	return nil
}

func modelGroupRestricts(derived, base *xsd.ModelGroup) bool {
	return (&compileState{}).modelGroupRestricts(derived, base)
}

func (s *compileState) modelGroupRestricts(derived, base *xsd.ModelGroup) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	if derived == nil {
		return base == nil || modelGroupNullable(base, s)
	}
	if base == nil || derived.Compositor != base.Compositor {
		return false
	}

	switch derived.Compositor {
	case xsd.Sequence:
		baseIndex := 0
		for _, particle := range derived.Particles {
			if err := s.contextError(); err != nil {
				return false
			}
			for baseIndex < len(base.Particles) &&
				!s.particleRestricts(particle, base.Particles[baseIndex]) {
				if !particleNullable(base.Particles[baseIndex], s) {
					return false
				}
				baseIndex++
			}
			if baseIndex == len(base.Particles) {
				return false
			}
			baseIndex++
		}
		for ; baseIndex < len(base.Particles); baseIndex++ {
			if err := s.contextError(); err != nil {
				return false
			}
			if !particleNullable(base.Particles[baseIndex], s) {
				return false
			}
		}
		return true
	case xsd.Choice, xsd.All:
		for _, particle := range derived.Particles {
			if err := s.contextError(); err != nil {
				return false
			}
			if !s.particleRestrictsAny(particle, base.Particles) {
				return false
			}
		}
		if derived.Compositor == xsd.All {
			for _, baseParticle := range base.Particles {
				if err := s.contextError(); err != nil {
					return false
				}
				if baseParticle.MinOccurs != 0 && !s.anyParticleRestrictsBase(derived.Particles, baseParticle) {
					return false
				}
			}
		}
		return true
	default:
		return false
	}
}

func (s *compileState) particleRestrictsAny(derived xsd.Particle, candidates []xsd.Particle) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	for _, candidate := range candidates {
		if err := s.contextError(); err != nil {
			return false
		}
		if s.particleRestricts(derived, candidate) {
			return true
		}
	}
	return false
}

func (s *compileState) anyParticleRestrictsBase(candidates []xsd.Particle, base xsd.Particle) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	for _, candidate := range candidates {
		if err := s.contextError(); err != nil {
			return false
		}
		if s.particleRestricts(candidate, base) {
			return true
		}
	}
	return false
}

func modelGroupNullable(group *xsd.ModelGroup, owner ...*compileState) bool {
	if compileOwnerError(owner) != nil {
		return false
	}
	if group == nil || len(group.Particles) == 0 {
		return true
	}
	if group.Compositor == xsd.Choice {
		for _, particle := range group.Particles {
			if compileOwnerError(owner) != nil {
				return false
			}
			if particleNullable(particle, owner...) {
				return true
			}
		}
		return false
	}
	for _, particle := range group.Particles {
		if compileOwnerError(owner) != nil {
			return false
		}
		if !particleNullable(particle, owner...) {
			return false
		}
	}
	return true
}

func particleRestricts(derived, base xsd.Particle) bool {
	return (&compileState{}).particleRestricts(derived, base)
}

func (s *compileState) particleRestricts(derived, base xsd.Particle) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	if derived.MinOccurs < base.MinOccurs {
		return false
	}
	if !base.Unbounded && derived.Unbounded {
		return false
	}
	if !base.Unbounded && derived.MaxOccurs > base.MaxOccurs {
		return false
	}
	if derived.Element != nil && base.Element != nil {
		return s.elementTermRestricts(*derived.Element, *base.Element)
	}
	if derived.Group != nil {
		if base.Group != nil {
			return s.modelGroupRestricts(derived.Group, base.Group)
		}
		return false
	}
	if derived.Wildcard != nil {
		if base.Wildcard != nil {
			return wildcardRestricts(derived.Wildcard, base.Wildcard, s)
		}
		return false
	}
	return false
}

func elementTermEqual(derived, base xsd.Element) bool {
	return (&compileState{}).elementTermRestricts(derived, base)
}

func (s *compileState) elementTermRestricts(derived, base xsd.Element) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	if derived.Ref.Local != "" {
		return derived.Ref == base.Ref
	}
	if base.Ref.Local != "" {
		return derived.Ref == base.Ref
	}
	if derived.Name != base.Name {
		return false
	}
	if derived.Namespace != base.Namespace {
		return false
	}
	if derived.Nillable && !base.Nillable {
		return false
	}
	if elementFixedSet(base) && !elementFixedSet(derived) {
		return false
	}
	if elementFixedSet(base) && derived.Fixed != base.Fixed {
		return false
	}
	if base.InlineSimpleType != nil || base.InlineComplexType != nil {
		return false
	}
	_, valid := s.elementTypeDerivationMethods(derived, base.Type)
	return valid
}

func (s *compileState) attributesRestrict(
	derived []xsd.AttributeUse,
	base []xsd.AttributeUse,
	baseWildcard *xsd.Wildcard,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	return s.attributesRestrictContext(derived, base, baseWildcard, "")
}

func (s *compileState) attributesRestrictContext(
	derived []xsd.AttributeUse,
	base []xsd.AttributeUse,
	baseWildcard *xsd.Wildcard,
	targetNamespace string,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	uses := make(map[xsd.QName]xsd.AttributeUse, len(derived))
	for _, attribute := range derived {
		if err := s.contextError(); err != nil {
			return false
		}
		uses[attributeUseName(attribute)] = attribute
	}
	for _, baseAttribute := range base {
		if err := s.contextError(); err != nil {
			return false
		}
		attribute, ok := uses[attributeUseName(baseAttribute)]
		if !ok {
			if baseAttribute.Use == xsd.AttributeRequired {
				return false
			}
		} else {
			delete(uses, attributeUseName(baseAttribute))
			baseFixed, baseFixedSet := s.attributeUseFixedConstraint(baseAttribute)
			fixed, fixedSet := s.attributeUseFixedConstraint(attribute)
			if baseFixedSet && !fixedSet {
				return false
			}
			if baseFixedSet && fixed != baseFixed {
				return false
			}
			if baseAttribute.Use == xsd.AttributeRequired && attribute.Use != xsd.AttributeRequired {
				return false
			}
			if attribute.Use != xsd.AttributeProhibited && !s.attributeUseTypeRestricts(attribute, baseAttribute) {
				return false
			}
		}
	}
	for name, attribute := range uses {
		if err := s.contextError(); err != nil {
			return false
		}
		if attribute.Use != xsd.AttributeProhibited &&
			!wildcardAllows(baseWildcard, name.Namespace, targetNamespace, s) {
			return false
		}
	}
	return true
}

func (s *compileState) attributeUseTypeRestricts(
	derived xsd.AttributeUse,
	base xsd.AttributeUse,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	if derived.Ref.Local != "" {
		if derived.Ref == base.Ref {
			_, ok := s.attributes[derived.Ref]
			return ok
		}
		return false
	}
	baseType, baseInline, ok := s.attributeUseType(base)
	if !ok || baseInline != nil {
		return false
	}
	derivedType, derivedInline, _ := s.attributeUseType(derived)
	if derivedInline != nil {
		return s.inlineSimpleTypeDerivesFrom(*derivedInline, baseType)
	}
	return s.simpleTypeDerivesFrom(derivedType, baseType)
}

func (s *compileState) attributeUseType(
	attribute xsd.AttributeUse,
) (xsd.QName, *xsd.SimpleType, bool) {
	if err := s.contextError(); err != nil {
		return xsd.QName{}, nil, false
	}

	if attribute.Ref.Local != "" {
		declaration, ok := s.attributes[attribute.Ref]
		if !ok {
			return xsd.QName{}, nil, false
		}
		if declaration.InlineSimpleType != nil {
			return xsd.QName{}, declaration.InlineSimpleType, true
		}
		typeName := declaration.Type
		if typeName.Local == "" {
			typeName = xsd.QName{Namespace: xsd.Namespace, Local: "anySimpleType"}
		}
		return typeName, nil, true
	}
	if attribute.InlineSimpleType != nil {
		return xsd.QName{}, attribute.InlineSimpleType, true
	}
	typeName := attribute.Type
	if typeName.Local == "" {
		typeName = xsd.QName{Namespace: xsd.Namespace, Local: "anySimpleType"}
	}
	return typeName, nil, true
}

func (s *compileState) attributeUseFixedConstraint(
	attribute xsd.AttributeUse,
) (string, bool) {
	if err := s.contextError(); err != nil {
		return "", false
	}

	if attributeFixedSet(attribute) {
		return attribute.Fixed, true
	}
	if attribute.Ref.Local == "" {
		return "", false
	}
	declaration, ok := s.attributes[attribute.Ref]
	if !ok || !attributeDeclarationFixedSet(declaration) {
		return "", false
	}
	return declaration.Fixed, true
}

func (s *compileState) simpleTypeDerivesFrom(derived, base xsd.QName) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	if base.Local == "" || base == (xsd.QName{Namespace: xsd.Namespace, Local: "anySimpleType"}) {
		return derived.Local != ""
	}
	seen := make(map[xsd.QName]struct{})
	for derived.Local != "" {
		if err := s.contextError(); err != nil {
			return false
		}
		if derived == base {
			return true
		}
		if _, duplicate := seen[derived]; duplicate {
			return false
		}
		seen[derived] = struct{}{}
		if definition, ok := s.simpleTypes[derived]; ok {
			derived = definition.Base
			continue
		}
		if derived.Namespace != xsd.Namespace {
			return false
		}
		parent, ok := datatype.BuiltInBase(derived.Local)
		if !ok {
			return false
		}
		derived = xsd.QName{Namespace: xsd.Namespace, Local: parent}
	}
	return false
}

func restrictedAttributes(base, derived []xsd.AttributeUse, owner ...*compileState) []xsd.AttributeUse {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	result := append([]xsd.AttributeUse(nil), base...)
	indexes := make(map[xsd.QName]int, len(result))
	for index, attribute := range result {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		indexes[attributeUseName(attribute)] = index
	}
	for _, attribute := range derived {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		name := attributeUseName(attribute)
		if index, ok := indexes[name]; ok {
			result[index] = attribute
		} else {
			indexes[name] = len(result)
			result = append(result, attribute)
		}
	}
	return result
}

func attributeUseName(attribute xsd.AttributeUse) xsd.QName {
	if attribute.Ref.Local != "" {
		return attribute.Ref
	}
	return xsd.QName{Namespace: attribute.Namespace, Local: attribute.Name}
}

func wildcardRestricts(derived, base *xsd.Wildcard, owner ...*compileState) bool {
	if err := compileOwnerError(owner); err != nil {
		return false
	}

	if derived.ProcessContents == xsd.ProcessSkip && base.ProcessContents != xsd.ProcessSkip ||
		derived.ProcessContents == xsd.ProcessLax && base.ProcessContents == xsd.ProcessStrict {
		return false
	}
	baseNamespaces := make(map[string]struct{}, len(base.Namespaces))
	for _, namespace := range base.Namespaces {
		if err := compileOwnerError(owner); err != nil {
			return false
		}
		baseNamespaces[namespace] = struct{}{}
	}
	for _, namespace := range derived.Namespaces {
		if err := compileOwnerError(owner); err != nil {
			return false
		}
		if _, ok := baseNamespaces[namespace]; !ok {
			return false
		}
	}
	return true
}

func intersectWildcards(left, right *xsd.Wildcard, owner ...*compileState) *xsd.Wildcard {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	if wildcardHas(left, "##any", owner...) {
		return cloneWildcard(right, owner...)
	}
	if wildcardHas(right, "##any", owner...) {
		return cloneWildcard(left, owner...)
	}
	rightNamespaces := make(map[string]struct{}, len(right.Namespaces))
	for _, namespace := range right.Namespaces {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		rightNamespaces[namespace] = struct{}{}
	}
	namespaces := make([]string, 0)
	for _, namespace := range left.Namespaces {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		if _, ok := rightNamespaces[namespace]; ok {
			namespaces = append(namespaces, namespace)
		}
	}
	if len(namespaces) == 0 {
		return nil
	}
	return &xsd.Wildcard{
		Namespaces:      namespaces,
		ProcessContents: strongerProcessContents(left.ProcessContents, right.ProcessContents),
	}
}

func unionWildcards(left, right *xsd.Wildcard, owner ...*compileState) *xsd.Wildcard {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	if wildcardHas(left, "##any", owner...) {
		return &xsd.Wildcard{
			Namespaces:      []string{"##any"},
			ProcessContents: weakerProcessContents(left.ProcessContents, right.ProcessContents),
		}
	}
	if wildcardHas(right, "##any", owner...) {
		return &xsd.Wildcard{
			Namespaces:      []string{"##any"},
			ProcessContents: weakerProcessContents(left.ProcessContents, right.ProcessContents),
		}
	}
	namespaces := append([]string(nil), left.Namespaces...)
	seen := make(map[string]struct{}, len(namespaces))
	for _, namespace := range namespaces {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		seen[namespace] = struct{}{}
	}
	for _, namespace := range right.Namespaces {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		if _, ok := seen[namespace]; !ok {
			namespaces = append(namespaces, namespace)
			seen[namespace] = struct{}{}
		}
	}
	return &xsd.Wildcard{
		Namespaces:      namespaces,
		ProcessContents: weakerProcessContents(left.ProcessContents, right.ProcessContents),
	}
}

func wildcardHas(wildcard *xsd.Wildcard, namespace string, owner ...*compileState) bool {
	if err := compileOwnerError(owner); err != nil {
		return false
	}

	for _, candidate := range wildcard.Namespaces {
		if err := compileOwnerError(owner); err != nil {
			return false
		}
		if candidate == namespace {
			return true
		}
	}
	return false
}

func strongerProcessContents(left, right xsd.ProcessContents) xsd.ProcessContents {
	switch left {
	case xsd.ProcessStrict:
		return xsd.ProcessStrict
	case xsd.ProcessLax:
		if right == xsd.ProcessStrict {
			return xsd.ProcessStrict
		}
		return xsd.ProcessLax
	default:
		return right
	}
}

func weakerProcessContents(left, right xsd.ProcessContents) xsd.ProcessContents {
	switch left {
	case xsd.ProcessSkip:
		return xsd.ProcessSkip
	case xsd.ProcessLax:
		if right == xsd.ProcessSkip {
			return xsd.ProcessSkip
		}
		return xsd.ProcessLax
	default:
		return right
	}
}

func processContentsRank(value xsd.ProcessContents) int {
	switch value {
	case xsd.ProcessStrict:
		return 3
	case xsd.ProcessLax:
		return 2
	default:
		return 1
	}
}

func extendContent(base *xsd.ModelGroup, extension *xsd.ModelGroup, owner ...*compileState) *xsd.ModelGroup {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	if base == nil {
		return cloneModelGroup(extension, owner...)
	}
	if extension == nil {
		return cloneModelGroup(base, owner...)
	}
	if !admitParticleCopies(2, owner) {
		return nil
	}
	return &xsd.ModelGroup{
		Compositor: xsd.Sequence,
		Particles: []xsd.Particle{
			{MinOccurs: 1, MaxOccurs: 1, Group: cloneModelGroup(base, owner...)},
			{MinOccurs: 1, MaxOccurs: 1, Group: cloneModelGroup(extension, owner...)},
		},
	}
}

func (s *compileState) validateComponents() error {
	if err := s.contextError(); err != nil {
		return err
	}

	if err := s.validateIdentityConstraints(); err != nil {
		return err
	}
	for name, element := range s.elements {
		if err := s.contextError(); err != nil {
			return err
		}
		if elementDefaultSet(element) && elementFixedSet(element) {
			return invalidComponent("element", name, "default and fixed are mutually exclusive")
		}
		if element.Ref.Local != "" {
			return invalidComponent("global element", name, "ref is not allowed")
		}
		if element.Type.Local != "" && !s.typeExists(element.Type, "") {
			return unresolvedComponent("type", element.Type)
		}
		if directNotationType(element.Type) {
			return invalidComponent("element", name, "NOTATION cannot be used directly")
		}
		if err := s.validateAnonymousElementType(element, name.Namespace, nil); err != nil {
			return err
		}
		if err := s.validateElementValueConstraint(element); err != nil {
			return invalidComponent("element", name, err.Error())
		}
	}
	for name, attribute := range s.attributes {
		if err := s.contextError(); err != nil {
			return err
		}
		if attributeDeclarationDefaultSet(attribute) && attributeDeclarationFixedSet(attribute) {
			return invalidComponent("attribute", name, "default and fixed are mutually exclusive")
		}
		if attribute.Type.Local != "" && !s.typeExists(attribute.Type, "simple") {
			return unresolvedComponent("simple type", attribute.Type)
		}
		if directNotationType(attribute.Type) {
			return invalidComponent("attribute", name, "NOTATION cannot be used directly")
		}
		if attribute.Type.Local != "" && attribute.InlineSimpleType != nil {
			return invalidComponent(
				"attribute",
				name,
				"has more than one type definition",
			)
		}
		if attribute.InlineSimpleType != nil {
			if err := s.validateSimpleTypeDefinition(*attribute.InlineSimpleType); err != nil {
				return err
			}
		}
		if err := s.validateAttributeDeclarationValueConstraint(attribute); err != nil {
			return invalidComponent("attribute", name, err.Error())
		}
	}
	for name, simpleType := range s.simpleTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		if datatype.ValidateBuiltInLexical("NCName", simpleType.Name) != nil {
			return invalidComponent("simple type", name, "name is not an NCName")
		}
		if err := s.validateSimpleTypeDefinition(simpleType); err != nil {
			return invalidComponent("simple type", name, err.Error())
		}
	}
	colors := make(map[xsd.QName]uint8, len(s.simpleTypes))
	for name := range s.simpleTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := s.validateSimpleTypeAcyclic(name, colors); err != nil {
			return err
		}
	}
	particles := 0
	for name, element := range s.elements {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := s.validateAnonymousElementType(element, name.Namespace, &particles); err != nil {
			return err
		}
	}
	for _, group := range s.modelGroups {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := s.validateModelGroup(group.Content, "", &particles); err != nil {
			return err
		}
	}
	for _, group := range s.attributeGroups {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := validateWildcard(group.Wildcard, s); err != nil {
			return err
		}
		if err := s.validateAttributeUseSet(group.Attributes); err != nil {
			return err
		}
	}
	for name, complexType := range s.complexTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := s.validateModelGroup(complexType.Content, name.Namespace, &particles); err != nil {
			return err
		}
		if err := validateWildcard(complexType.AttributeWildcard, s); err != nil {
			return err
		}
		for _, attribute := range complexType.Attributes {
			if err := s.contextError(); err != nil {
				return err
			}
			if err := s.validateAttributeUse(attribute, name.Namespace); err != nil {
				return err
			}
		}
		if err := s.validateAttributeUseSet(complexType.Attributes); err != nil {
			return err
		}
	}
	return nil
}

func (s *compileState) validateAnonymousElementType(
	element xsd.Element,
	namespace string,
	particles *int,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	typeCount := 0
	if element.Type.Local != "" {
		typeCount++
		if directNotationType(element.Type) {
			return fmt.Errorf("%w: NOTATION cannot be used directly", ErrInvalidComponent)
		}
	}
	if element.InlineSimpleType != nil {
		typeCount++
		if err := s.validateSimpleTypeDefinition(*element.InlineSimpleType); err != nil {
			return err
		}
	}
	if element.InlineComplexType != nil {
		typeCount++
		if err := s.validateAttributeUseSet(
			element.InlineComplexType.Attributes,
		); err != nil {
			return err
		}
		if particles != nil {
			if err := s.validateModelGroup(
				element.InlineComplexType.Content,
				namespace,
				particles,
			); err != nil {
				return err
			}
			for _, attribute := range element.InlineComplexType.Attributes {
				if err := s.contextError(); err != nil {
					return err
				}
				if err := s.validateAttributeUse(attribute, namespace); err != nil {
					return err
				}
			}
		}
	}
	if typeCount > 1 {
		return fmt.Errorf(
			"%w: element has more than one type definition",
			ErrInvalidComponent,
		)
	}
	return nil
}

func (s *compileState) validateAttributeUseSet(attributes []xsd.AttributeUse) error {
	if err := s.contextError(); err != nil {
		return err
	}

	seen := make(map[xsd.QName]struct{}, len(attributes))
	idSeen := false
	id := xsd.QName{Namespace: xsd.Namespace, Local: "ID"}
	for _, attribute := range attributes {
		if err := s.contextError(); err != nil {
			return err
		}
		name := attributeUseName(attribute)
		if _, duplicate := seen[name]; duplicate {
			return fmt.Errorf(
				"%w: duplicate attribute use {%s}%s",
				ErrInvalidComponent,
				name.Namespace,
				name.Local,
			)
		}
		seen[name] = struct{}{}
		if attribute.Use != xsd.AttributeProhibited {
			typeName, inline, ok := s.attributeUseType(attribute)
			if ok {
				isID := false
				if inline != nil {
					isID = s.inlineSimpleTypeDerivesFrom(*inline, id)
				} else {
					isID = s.simpleTypeDerivesFrom(typeName, id)
				}
				if isID {
					if idSeen {
						return fmt.Errorf(
							"%w: an attribute-use set cannot contain multiple ID types",
							ErrInvalidComponent,
						)
					}
					idSeen = true
				}
			}
		}
	}
	return nil
}

func (s *compileState) validateSimpleTypeDefinition(typeDefinition xsd.SimpleType) error {
	if err := s.contextError(); err != nil {
		return err
	}

	switch typeDefinition.Variety {
	case xsd.SimpleRestriction:
		if typeDefinition.Base.Local != "" && typeDefinition.InlineBase != nil {
			return fmt.Errorf("%w: simple restriction has more than one base", ErrInvalidComponent)
		}
		if typeDefinition.InlineBase != nil {
			if err := s.validateSimpleTypeDefinition(*typeDefinition.InlineBase); err != nil {
				return err
			}
		} else {
			if typeDefinition.Base.Local == "" || !s.typeExists(typeDefinition.Base, "simple") {
				return unresolvedComponent("simple type", typeDefinition.Base)
			}
			if base, ok := s.simpleTypes[typeDefinition.Base]; ok &&
				base.Final.Contains(xsd.DerivationRestriction) {
				return fmt.Errorf("%w: base type prohibits restriction", ErrInvalidComponent)
			}
		}
		if err := s.validateRestrictionFacets(typeDefinition); err != nil {
			return err
		}
	case xsd.SimpleList:
		if typeDefinition.ItemType.Local != "" && typeDefinition.InlineItem != nil {
			return fmt.Errorf("%w: list has more than one item type", ErrInvalidComponent)
		}
		if typeDefinition.InlineItem != nil {
			if err := s.validateSimpleTypeDefinition(*typeDefinition.InlineItem); err != nil {
				return err
			}
			if s.definitionShape(*typeDefinition.InlineItem, 0).variety == listShape {
				return fmt.Errorf("%w: list item type cannot itself be a list", ErrInvalidComponent)
			}
		} else {
			if typeDefinition.ItemType.Local == "" {
				return unresolvedComponent("simple type", typeDefinition.ItemType)
			}
			if !s.typeExists(typeDefinition.ItemType, "simple") {
				return unresolvedComponent("simple type", typeDefinition.ItemType)
			}
			if item, ok := s.simpleTypes[typeDefinition.ItemType]; ok &&
				item.Final.Contains(xsd.DerivationList) {
				return fmt.Errorf("%w: item type prohibits list derivation", ErrInvalidComponent)
			}
			if s.namedShape(typeDefinition.ItemType, 0).variety == listShape {
				return fmt.Errorf("%w: list item type cannot itself be a list", ErrInvalidComponent)
			}
			if directNotationType(typeDefinition.ItemType) {
				return fmt.Errorf("%w: NOTATION cannot be used directly as a list item", ErrInvalidComponent)
			}
		}
	case xsd.SimpleUnion:
		if len(typeDefinition.MemberTypes) == 0 && len(typeDefinition.InlineMembers) == 0 {
			return fmt.Errorf("%w: anonymous union has no member types", ErrInvalidComponent)
		}
		for _, member := range typeDefinition.MemberTypes {
			if err := s.contextError(); err != nil {
				return err
			}
			if !s.typeExists(member, "simple") {
				return unresolvedComponent("simple type", member)
			}
			if definition, ok := s.simpleTypes[member]; ok &&
				definition.Final.Contains(xsd.DerivationUnion) {
				return fmt.Errorf("%w: member type prohibits union derivation", ErrInvalidComponent)
			}
			if directNotationType(member) {
				return fmt.Errorf("%w: NOTATION cannot be used directly as a union member", ErrInvalidComponent)
			}
		}
		for _, member := range typeDefinition.InlineMembers {
			if err := s.contextError(); err != nil {
				return err
			}
			if err := s.validateSimpleTypeDefinition(member); err != nil {
				return err
			}
		}
	default:
		return fmt.Errorf("%w: anonymous simple type has no variety", ErrInvalidComponent)
	}
	return nil
}

func (s *compileState) validateIdentityConstraints() error {
	if err := s.contextError(); err != nil {
		return err
	}

	constraints := make(map[xsd.QName]xsd.IdentityConstraint)
	collect := func(namespace string, element xsd.Element) error {
		for _, constraint := range element.IdentityConstraints {
			if err := s.contextError(); err != nil {
				return err
			}
			name := xsd.QName{Namespace: namespace, Local: constraint.Name}
			if constraint.Name == "" {
				return invalidComponent(
					"identity constraint",
					name,
					"requires a name, selector, and at least one field",
				)
			}
			if constraint.Selector == "" {
				return invalidComponent(
					"identity constraint",
					name,
					"requires a name, selector, and at least one field",
				)
			}
			if len(constraint.Fields) == 0 {
				return invalidComponent(
					"identity constraint",
					name,
					"requires a name, selector, and at least one field",
				)
			}
			if _, duplicate := constraints[name]; duplicate {
				return duplicateComponent("identity constraint", name)
			}
			if !validIdentitySelector(constraint.Selector, constraint.Namespaces, s) {
				return invalidComponent(
					"identity constraint",
					name,
					"selector is outside the XML Schema XPath subset",
				)
			}
			for _, field := range constraint.Fields {
				if err := s.contextError(); err != nil {
					return err
				}
				if !validIdentityField(field, constraint.Namespaces, s) {
					return invalidComponent(
						"identity constraint",
						name,
						"field is outside the XML Schema XPath subset",
					)
				}
			}
			if constraint.Kind == xsd.IdentityKeyRef && constraint.Refer.Local == "" {
				return invalidComponent(
					"identity constraint",
					name,
					"keyref requires refer",
				)
			}
			if constraint.Kind != xsd.IdentityKeyRef && constraint.Refer.Local != "" {
				return invalidComponent(
					"identity constraint",
					name,
					"only keyref permits refer",
				)
			}
			constraints[name] = constraint
		}
		return nil
	}
	for name, element := range s.elements {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := collect(name.Namespace, element); err != nil {
			return err
		}
		if element.InlineComplexType != nil {
			if err := collectModelIdentityConstraints(
				element.InlineComplexType.Content,
				name.Namespace,
				collect,
				s); err != nil {
				return err
			}
		}
	}
	for name, typeDefinition := range s.complexTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		if err := collectModelIdentityConstraints(typeDefinition.Content, name.Namespace, collect, s); err != nil {
			return err
		}
	}
	for _, name := range sortedComponentNames(constraints, s) {
		if err := s.contextError(); err != nil {
			return err
		}
		constraint := constraints[name]
		if constraint.Kind == xsd.IdentityKeyRef {
			referenced, ok := constraints[constraint.Refer]
			if !ok || referenced.Kind == xsd.IdentityKeyRef {
				return unresolvedComponent("identity constraint", constraint.Refer)
			}
			if len(referenced.Fields) != len(constraint.Fields) {
				return invalidComponent(
					"identity constraint",
					name,
					"keyref field count differs from its referenced constraint",
				)
			}
		}
	}
	return nil
}

func validIdentitySelector(expression string, namespaces map[string]string, owner ...*compileState) bool {
	if err := compileOwnerError(owner); err != nil {
		return false
	}

	expression = xsd.NormalizeIdentityXPath(expression)
	for _, branch := range strings.Split(expression, "|") {
		if err := compileOwnerError(owner); err != nil {
			return false
		}
		branch = strings.TrimSpace(branch)
		if branch == "" {
			return false
		}
		if branch != "." {
			branch = strings.TrimPrefix(branch, ".//")
			branch = strings.TrimPrefix(branch, "./")
			for _, step := range strings.Split(branch, "/") {
				if err := compileOwnerError(owner); err != nil {
					return false
				}
				if step != "." {
					step = strings.TrimPrefix(step, "child::")
					if !validIdentityName(step, namespaces, true) {
						return false
					}
				}
			}
		}
	}
	return true
}

func validIdentityField(expression string, namespaces map[string]string, owner ...*compileState) bool {
	if err := compileOwnerError(owner); err != nil {
		return false
	}

	expression = xsd.NormalizeIdentityXPath(expression)
	for _, branch := range strings.Split(expression, "|") {
		if err := compileOwnerError(owner); err != nil {
			return false
		}
		branch = strings.TrimSpace(branch)
		if branch != "." {
			branch = strings.TrimPrefix(branch, ".//")
			branch = strings.TrimPrefix(branch, "./")
			steps := strings.Split(branch, "/")
			for index, step := range steps {
				if err := compileOwnerError(owner); err != nil {
					return false
				}
				attribute := strings.HasPrefix(step, "@")
				if !attribute {
					attribute = strings.HasPrefix(step, "attribute::")
				}
				if attribute {
					if index != len(steps)-1 {
						return false
					}
					step = strings.TrimPrefix(step, "@")
					step = strings.TrimPrefix(step, "attribute::")
				}
				if step != "." || attribute {
					step = strings.TrimPrefix(step, "child::")
					if !validIdentityName(step, namespaces, true) {
						return false
					}
				}
			}
		}
	}
	return true
}

func validIdentityName(
	expression string,
	namespaces map[string]string,
	allowWildcard bool,
) bool {
	if expression == "*" {
		return allowWildcard
	}
	parts := strings.Split(expression, ":")
	if len(parts) == 1 {
		return datatype.ValidateBuiltInLexical("NCName", parts[0]) == nil
	}
	if len(parts) != 2 {
		return false
	}
	if datatype.ValidateBuiltInLexical("NCName", parts[0]) != nil {
		return false
	}
	if parts[1] != "*" && datatype.ValidateBuiltInLexical("NCName", parts[1]) != nil {
		return false
	}
	_, declared := namespaces[parts[0]]
	return declared
}

func collectModelIdentityConstraints(
	group *xsd.ModelGroup,
	namespace string,
	collect func(string, xsd.Element) error,

	owner ...*compileState,
) error {
	if err := compileOwnerError(owner); err != nil {
		return err
	}

	if group == nil {
		return nil
	}
	for _, particle := range group.Particles {
		if err := compileOwnerError(owner); err != nil {
			return err
		}
		if particle.Element != nil {
			if err := collect(namespace, *particle.Element); err != nil {
				return err
			}
			if particle.Element.InlineComplexType != nil {
				if err := collectModelIdentityConstraints(
					particle.Element.InlineComplexType.Content,
					namespace,
					collect,
					owner...); err != nil {
					return err
				}
			}
		}
		if err := collectModelIdentityConstraints(particle.Group, namespace, collect, owner...); err != nil {
			return err
		}
	}
	return nil
}

func (s *compileState) validateSimpleTypeAcyclic(
	name xsd.QName,
	colors map[xsd.QName]uint8,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	switch colors[name] {
	case 1:
		return invalidComponent("simple type", name, "derivation is recursive")
	case 2:
		return nil
	}
	colors[name] = 1
	typeDefinition := s.simpleTypes[name]
	dependencies := []xsd.QName{}
	switch typeDefinition.Variety {
	case xsd.SimpleRestriction:
		dependencies = append(dependencies, typeDefinition.Base)
	case xsd.SimpleList:
		dependencies = append(dependencies, typeDefinition.ItemType)
	case xsd.SimpleUnion:
		dependencies = append(dependencies, typeDefinition.MemberTypes...)
	}
	for _, dependency := range dependencies {
		if err := s.contextError(); err != nil {
			return err
		}
		if _, userDefined := s.simpleTypes[dependency]; userDefined {
			if err := s.validateSimpleTypeAcyclic(dependency, colors); err != nil {
				return err
			}
		}
	}
	colors[name] = 2
	return nil
}

func (s *compileState) validateModelGroup(
	group *xsd.ModelGroup,
	typeNamespace string,
	particles *int,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if group == nil {
		return nil
	}
	if err := s.validateUniqueParticleAttribution(group, typeNamespace); err != nil {
		return err
	}
	for _, particle := range group.Particles {
		if err := s.contextError(); err != nil {
			return err
		}
		*particles++
		if *particles > s.compiler.limits.MaxParticles {
			return fmt.Errorf(
				"%w: particle count exceeds %d",
				ErrLimitExceeded,
				s.compiler.limits.MaxParticles,
			)
		}
		if group.Compositor == xsd.All {
			invalidAllParticle := particle.Group != nil
			if particle.Wildcard != nil {
				invalidAllParticle = true
			}
			if particle.Unbounded {
				invalidAllParticle = true
			}
			if particle.MaxOccurs > 1 {
				invalidAllParticle = true
			}
			if particle.MinOccurs > 1 {
				invalidAllParticle = true
			}
			if invalidAllParticle {
				return fmt.Errorf(
					"%w: all compositor particles must be elements with 0..1 or 1..1 occurrence",
					ErrInvalidComponent,
				)
			}
		}
		if particle.Element != nil {
			element := particle.Element
			if element.SubstitutionGroup.Local != "" {
				return fmt.Errorf(
					"%w: local element cannot declare a substitution group",
					ErrInvalidComponent,
				)
			}
			if (element.Name == "") == (element.Ref.Local == "") {
				return fmt.Errorf(
					"%w: local element must have exactly one of name or ref",
					ErrInvalidComponent,
				)
			}
			if elementDefaultSet(*element) {
				if elementFixedSet(*element) {
					return fmt.Errorf(
						"%w: local element default and fixed are mutually exclusive",
						ErrInvalidComponent,
					)
				}
			}
			if element.Ref.Local != "" {
				if element.Type.Local != "" || element.Form != "" || element.Abstract ||
					element.Nillable || elementDefaultSet(*element) || elementFixedSet(*element) ||
					element.Block.String() != "" || element.Final.String() != "" ||
					element.InlineSimpleType != nil || element.InlineComplexType != nil ||
					len(element.IdentityConstraints) > 0 {
					return fmt.Errorf(
						"%w: local element reference has declaration-only properties",
						ErrInvalidComponent,
					)
				}
				if _, ok := s.elements[element.Ref]; !ok {
					return unresolvedComponent("element", element.Ref)
				}
			} else if element.Type.Local != "" && !s.typeExists(element.Type, "") {
				return unresolvedComponent("type", element.Type)
			}
			if err := s.validateAnonymousElementType(*element, typeNamespace, particles); err != nil {
				return err
			}
			if element.Ref.Local == "" {
				if err := s.validateElementValueConstraint(*element); err != nil {
					return fmt.Errorf("%w: local element: %s", ErrInvalidComponent, err)
				}
			}
		}
		if err := validateWildcard(particle.Wildcard, s); err != nil {
			return err
		}
		if particle.Element == nil && particle.Group == nil && particle.Wildcard == nil {
			return fmt.Errorf("%w: particle has no term", ErrInvalidComponent)
		}
		if err := s.validateModelGroup(particle.Group, typeNamespace, particles); err != nil {
			return err
		}
	}
	return nil
}

func (s *compileState) validateElementValueConstraint(element xsd.Element) error {
	if err := s.contextError(); err != nil {
		return err
	}

	lexical := element.Default
	if elementFixedSet(element) {
		lexical = element.Fixed
	}
	if !elementDefaultSet(element) && !elementFixedSet(element) {
		return nil
	}
	if element.InlineSimpleType != nil {
		if s.inlineConstraintValidContext(
			*element.InlineSimpleType,
			lexical,
			element.ValueNamespaces,
		) {
			return nil
		}
		return errors.New("value constraint is invalid for the anonymous simple type")
	}
	if element.InlineComplexType != nil {
		if element.InlineComplexType.SimpleContent {
			if element.InlineComplexType.InlineSimpleType != nil {
				if s.inlineConstraintValidContext(
					*element.InlineComplexType.InlineSimpleType,
					lexical,
					element.ValueNamespaces,
				) {
					return nil
				}
			} else if s.simpleConstraintValidContext(
				element.InlineComplexType.SimpleBase,
				lexical,
				element.ValueNamespaces,
			) {
				return nil
			}
		}
		if element.InlineComplexType.Mixed {
			return nil
		}
		return errors.New("value constraint requires simple or mixed content")
	}
	if element.Type.Local == "" || element.Type == (xsd.QName{Namespace: xsd.Namespace, Local: "anyType"}) {
		return nil
	}
	id := xsd.QName{Namespace: xsd.Namespace, Local: "ID"}
	if s.typeExists(element.Type, "simple") && s.simpleTypeDerivesFrom(element.Type, id) {
		return errors.New("ID-typed elements cannot have value constraints")
	}
	if s.simpleConstraintValidContext(element.Type, lexical, element.ValueNamespaces) {
		return nil
	}
	if complexType, ok := s.complexTypes[element.Type]; ok {
		if complexType.Mixed {
			return nil
		}
		if complexType.SimpleContent {
			if complexType.InlineSimpleType != nil {
				if s.inlineConstraintValidContext(
					*complexType.InlineSimpleType,
					lexical,
					element.ValueNamespaces,
				) {
					return nil
				}
			} else if s.simpleConstraintValidContext(
				complexType.SimpleBase,
				lexical,
				element.ValueNamespaces,
			) {
				return nil
			}
		}
	}
	return errors.New("value constraint is invalid for the element type")
}

func (s *compileState) validateAttributeDeclarationValueConstraint(attribute xsd.Attribute) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if !attributeDeclarationDefaultSet(attribute) && !attributeDeclarationFixedSet(attribute) {
		return nil
	}
	lexical := attribute.Default
	if attributeDeclarationFixedSet(attribute) {
		lexical = attribute.Fixed
	}
	return s.validateAttributeConstraintContext(
		attribute.Type,
		attribute.InlineSimpleType,
		lexical,
		attribute.ValueNamespaces,
	)
}

func (s *compileState) validateAttributeUseValueConstraint(attribute xsd.AttributeUse) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if !attributeDefaultSet(attribute) && !attributeFixedSet(attribute) {
		return nil
	}
	lexical := attribute.Default
	if attributeFixedSet(attribute) {
		lexical = attribute.Fixed
	}
	if attribute.Ref.Local != "" {
		declaration, ok := s.attributes[attribute.Ref]
		if !ok {
			return unresolvedComponent("attribute", attribute.Ref)
		}
		return s.validateAttributeConstraintContext(
			declaration.Type,
			declaration.InlineSimpleType,
			lexical,
			attribute.ValueNamespaces,
		)
	}
	return s.validateAttributeConstraintContext(
		attribute.Type,
		attribute.InlineSimpleType,
		lexical,
		attribute.ValueNamespaces,
	)
}

func (s *compileState) validateAttributeConstraint(
	typeName xsd.QName,
	inline *xsd.SimpleType,
	lexical string,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	return s.validateAttributeConstraintContext(typeName, inline, lexical, nil)
}

func (s *compileState) validateAttributeConstraintContext(
	typeName xsd.QName,
	inline *xsd.SimpleType,
	lexical string,
	namespaces map[string]string,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	id := xsd.QName{Namespace: xsd.Namespace, Local: "ID"}
	if inline != nil {
		if s.inlineSimpleTypeDerivesFrom(*inline, id) {
			return fmt.Errorf("%w: ID-typed attributes cannot have value constraints", ErrInvalidComponent)
		}
		if s.inlineConstraintValidContext(*inline, lexical, namespaces) {
			return nil
		}
		return fmt.Errorf("%w: value constraint is invalid for the anonymous simple type", ErrInvalidComponent)
	}
	if typeName.Local == "" {
		typeName = xsd.QName{Namespace: xsd.Namespace, Local: "anySimpleType"}
	}
	if s.simpleTypeDerivesFrom(typeName, id) {
		return fmt.Errorf("%w: ID-typed attributes cannot have value constraints", ErrInvalidComponent)
	}
	if s.simpleConstraintValidContext(typeName, lexical, namespaces) {
		return nil
	}
	return fmt.Errorf("%w: value constraint is invalid for the attribute type", ErrInvalidComponent)
}

func (s *compileState) inlineSimpleTypeDerivesFrom(
	typeDefinition xsd.SimpleType,
	base xsd.QName,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	for depth := 0; !compileDepthExceeded(depth); depth = compileChildDepth(depth) {
		if err := s.contextError(); err != nil {
			return false
		}
		if typeDefinition.Variety != xsd.SimpleRestriction {
			return false
		}
		if typeDefinition.InlineBase == nil {
			return s.simpleTypeDerivesFrom(typeDefinition.Base, base)
		}
		typeDefinition = *typeDefinition.InlineBase
	}
	return false
}

func elementDefaultSet(element xsd.Element) bool {
	return element.DefaultSet || element.Default != ""
}

func elementFixedSet(element xsd.Element) bool {
	return element.FixedSet || element.Fixed != ""
}

func attributeDefaultSet(attribute xsd.AttributeUse) bool {
	return attribute.DefaultSet || attribute.Default != ""
}

func attributeFixedSet(attribute xsd.AttributeUse) bool {
	return attribute.FixedSet || attribute.Fixed != ""
}

func attributeDeclarationDefaultSet(attribute xsd.Attribute) bool {
	return attribute.DefaultSet || attribute.Default != ""
}

func attributeDeclarationFixedSet(attribute xsd.Attribute) bool {
	return attribute.FixedSet || attribute.Fixed != ""
}

func (s *compileState) simpleConstraintValid(typeName xsd.QName, lexical string) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	return s.simpleConstraintValidContext(typeName, lexical, nil)
}

func (s *compileState) simpleConstraintValidContext(
	typeName xsd.QName,
	lexical string,
	namespaces map[string]string,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	return s.simpleConstraintValidDepthContext(typeName, lexical, namespaces, 0)
}

func (s *compileState) simpleConstraintValidDepth(
	typeName xsd.QName,
	lexical string,
	depth int,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	return s.simpleConstraintValidDepthContext(typeName, lexical, nil, depth)
}

func (s *compileState) simpleConstraintValidDepthContext(
	typeName xsd.QName,
	lexical string,
	namespaces map[string]string,
	depth int,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	if compileDepthExceeded(depth) {
		return false
	}
	if typeName.Namespace == xsd.Namespace {
		whitespace, _ := s.namedWhitespace(typeName, depth)
		lexical = normalizeConstraintWhitespace(lexical, whitespace)
		switch typeName.Local {
		case "anySimpleType":
			return true
		case "boolean":
			return lexical == "true" || lexical == "false" || lexical == "1" || lexical == "0"
		case "decimal":
			_, err := datatype.ParseDecimal(lexical)
			return err == nil
		case "integer", "nonPositiveInteger", "negativeInteger", "long", "int", "short", "byte",
			"nonNegativeInteger", "unsignedLong", "unsignedInt", "unsignedShort", "unsignedByte",
			"positiveInteger":
			value, err := datatype.ParseInteger(lexical)
			return err == nil && datatype.ValidateBuiltInInteger(typeName.Local, value) == nil
		default:
			valid := datatype.ValidateBuiltInLexical(typeName.Local, lexical) == nil
			if valid && (typeName.Local == "QName" || typeName.Local == "NOTATION") {
				_, valid = resolveConstraintQName(lexical, namespaces)
			}
			return valid
		}
	}
	typeDefinition, ok := s.simpleTypes[typeName]
	return ok && s.inlineConstraintValidDepthContext(
		typeDefinition,
		lexical,
		namespaces,
		compileChildDepth(depth),
	)
}

func (s *compileState) inlineConstraintValid(typeDefinition xsd.SimpleType, lexical string) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	return s.inlineConstraintValidContext(typeDefinition, lexical, nil)
}

func (s *compileState) inlineConstraintValidContext(
	typeDefinition xsd.SimpleType,
	lexical string,
	namespaces map[string]string,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	return s.inlineConstraintValidDepthContext(typeDefinition, lexical, namespaces, 0)
}

func (s *compileState) inlineConstraintValidDepth(
	typeDefinition xsd.SimpleType,
	lexical string,
	depth int,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	return s.inlineConstraintValidDepthContext(typeDefinition, lexical, nil, depth)
}

func (s *compileState) inlineConstraintValidDepthContext(
	typeDefinition xsd.SimpleType,
	lexical string,
	namespaces map[string]string,
	depth int,
) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	if compileDepthExceeded(depth) {
		return false
	}
	switch typeDefinition.Variety {
	case xsd.SimpleRestriction:
		normalized := s.normalizeConstraintLexical(typeDefinition, lexical)
		valid := s.simpleConstraintValidDepthContext(
			typeDefinition.Base,
			normalized,
			namespaces,
			compileChildDepth(depth),
		)
		if typeDefinition.InlineBase != nil {
			valid = s.inlineConstraintValidDepthContext(
				*typeDefinition.InlineBase,
				normalized,
				namespaces,
				compileChildDepth(depth),
			)
		}
		if !valid {
			return false
		}
		hasPattern := false
		patternMatched := false
		for _, facet := range typeDefinition.Facets {
			if err := s.contextError(); err != nil {
				return false
			}
			if facet.Kind == xsd.FacetPattern {
				hasPattern = true
				pattern, err := datatype.CompilePatternContext(s.ctx, facet.Value)
				if err != nil {
					return false
				}
				if pattern.MatchString(normalized) {
					patternMatched = true
				}
			}
		}
		return (!hasPattern || patternMatched) &&
			s.restrictionConstraintFacetsValidContext(
				typeDefinition,
				normalized,
				namespaces,
				compileChildDepth(depth),
			)
	case xsd.SimpleList:
		items := strings.Fields(lexical)
		if len(items) == 0 {
			return false
		}
		for _, item := range items {
			if err := s.contextError(); err != nil {
				return false
			}
			valid := s.simpleConstraintValidDepthContext(
				typeDefinition.ItemType,
				item,
				namespaces,
				compileChildDepth(depth),
			)
			if typeDefinition.InlineItem != nil {
				valid = s.inlineConstraintValidDepthContext(
					*typeDefinition.InlineItem,
					item,
					namespaces,
					compileChildDepth(depth),
				)
			}
			if !valid {
				return false
			}
		}
		return true
	case xsd.SimpleUnion:
		for _, member := range typeDefinition.MemberTypes {
			if err := s.contextError(); err != nil {
				return false
			}
			if s.simpleConstraintValidDepthContext(member, lexical, namespaces, compileChildDepth(depth)) {
				return true
			}
		}
		for _, member := range typeDefinition.InlineMembers {
			if err := s.contextError(); err != nil {
				return false
			}
			if s.inlineConstraintValidDepthContext(member, lexical, namespaces, compileChildDepth(depth)) {
				return true
			}
		}
	}
	return false
}

type nameClass struct {
	name         *xsd.QName
	alternatives []xsd.QName
	wildcard     *xsd.Wildcard
}

func (s *compileState) validateUniqueParticleAttribution(
	group *xsd.ModelGroup,
	targetNamespace string,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	state := upaState{compile: s, follow: make(map[int]upaPositions)}
	info := state.group(group, targetNamespace)
	states := make([]upaPositions, 0, len(state.follow)+1)
	states = append(states, info.first)
	for _, positions := range state.follow {
		if err := s.contextError(); err != nil {
			return err
		}
		states = append(states, positions)
	}
	for _, positions := range states {
		if err := s.contextError(); err != nil {
			return err
		}
		values := make([]nameClass, 0, len(positions))
		for _, class := range positions {
			if err := s.contextError(); err != nil {
				return err
			}
			values = append(values, class)
		}
		for left := range values {
			if err := s.contextError(); err != nil {
				return err
			}
			for offset := range values[left+1:] {
				if err := s.contextError(); err != nil {
					return err
				}
				right := left + offset + 1
				if nameClassesOverlap(values[left], values[right], targetNamespace, s) {
					return fmt.Errorf(
						"%w: content model violates unique particle attribution",
						ErrInvalidComponent,
					)
				}
			}
		}
	}
	return nil
}

type upaPositions map[int]nameClass

type upaInfo struct {
	nullable bool
	first    upaPositions
	last     upaPositions
}

type upaState struct {
	compile *compileState
	next    int
	follow  map[int]upaPositions
}

func (s *upaState) group(group *xsd.ModelGroup, targetNamespace string) upaInfo {
	if err := s.compile.contextError(); err != nil {
		return upaInfo{}
	}

	children := make([]upaInfo, len(group.Particles))
	for index, particle := range group.Particles {
		if err := s.compile.contextError(); err != nil {
			return upaInfo{}
		}
		children[index] = s.particle(particle, targetNamespace)
	}
	var result upaInfo
	switch group.Compositor {
	case xsd.Choice:
		result = upaInfo{first: upaPositions{}, last: upaPositions{}}
		for _, child := range children {
			if err := s.compile.contextError(); err != nil {
				return upaInfo{}
			}
			if child.nullable {
				result.nullable = true
			}
			mergeUPAPositions(result.first, child.first, s.compile)
			mergeUPAPositions(result.last, child.last, s.compile)
		}
	case xsd.All:
		result = upaInfo{nullable: true, first: upaPositions{}, last: upaPositions{}}
		for _, child := range children {
			if err := s.compile.contextError(); err != nil {
				return upaInfo{}
			}
			if !child.nullable {
				result.nullable = false
			}
			mergeUPAPositions(result.first, child.first, s.compile)
			mergeUPAPositions(result.last, child.last, s.compile)
		}
		for left, child := range children {
			if err := s.compile.contextError(); err != nil {
				return upaInfo{}
			}
			for _, next := range children[left+1:] {
				if err := s.compile.contextError(); err != nil {
					return upaInfo{}
				}
				s.addFollow(child.last, next.first)
				s.addFollow(next.last, child.first)
			}
		}
	default:
		result = upaInfo{nullable: true, first: upaPositions{}, last: upaPositions{}}
		for _, child := range children {
			if err := s.compile.contextError(); err != nil {
				return upaInfo{}
			}
			if result.nullable {
				mergeUPAPositions(result.first, child.first, s.compile)
			}
			s.addFollow(result.last, child.first)
			if child.nullable {
				mergeUPAPositions(result.last, child.last, s.compile)
			} else {
				result.last = cloneUPAPositions(child.last, s.compile)
			}
			if !child.nullable {
				result.nullable = false
			}
		}
	}
	if group.OccursSet {
		result = s.occurs(result, group.MinOccurs, group.MaxOccurs, group.Unbounded)
	}
	return result
}

func (s *upaState) particle(particle xsd.Particle, targetNamespace string) upaInfo {
	if err := s.compile.contextError(); err != nil {
		return upaInfo{}
	}

	var result upaInfo
	if particle.Element != nil {
		name := particle.Element.Ref
		if name.Local == "" {
			name = xsd.QName{Namespace: particle.Element.Namespace, Local: particle.Element.Name}
		}
		class := nameClass{name: &name}
		if particle.Element.Ref.Local != "" {
			set := &Set{
				elements:          s.compile.elements,
				simpleTypes:       s.compile.simpleTypes,
				complexTypes:      s.compile.complexTypes,
				substitutionHeads: s.compile.substitutionHeads,
			}
			for member := range s.compile.substitutionHeads {
				if err := s.compile.contextError(); err != nil {
					return upaInfo{}
				}
				if _, ok := substitutionMember(set, particle.Element.Ref, member, s.compile); ok {
					class.alternatives = append(class.alternatives, member)
				}
			}
		}
		result = s.leaf(class)
	} else if particle.Wildcard != nil {
		result = s.leaf(nameClass{wildcard: particle.Wildcard})
	} else if particle.Group != nil {
		result = s.group(particle.Group, targetNamespace)
	} else {
		result = upaInfo{nullable: true, first: upaPositions{}, last: upaPositions{}}
	}
	return s.occurs(result, particle.MinOccurs, particle.MaxOccurs, particle.Unbounded)
}

func (s *upaState) leaf(class nameClass) upaInfo {
	if err := s.compile.contextError(); err != nil {
		return upaInfo{}
	}

	position := s.next
	s.next++
	positions := upaPositions{position: class}
	return upaInfo{first: positions, last: cloneUPAPositions(positions, s.compile)}
}

func (s *upaState) occurs(info upaInfo, minimum uint64, maximum uint64, unbounded bool) upaInfo {
	if err := s.compile.contextError(); err != nil {
		return upaInfo{}
	}

	if unbounded || maximum > 1 {
		s.addFollow(info.last, info.first)
	}
	if minimum == 0 {
		info.nullable = true
	}
	return info
}

func (s *upaState) addFollow(from upaPositions, to upaPositions) {
	if err := s.compile.contextError(); err != nil {
		return
	}

	for position := range from {
		if err := s.compile.contextError(); err != nil {
			return
		}
		if s.follow[position] == nil {
			s.follow[position] = upaPositions{}
		}
		mergeUPAPositions(s.follow[position], to, s.compile)
	}
}

func cloneUPAPositions(source upaPositions, owner ...*compileState) upaPositions {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	clone := make(upaPositions, len(source))
	mergeUPAPositions(clone, source, owner...)
	return clone
}

func mergeUPAPositions(target upaPositions, source upaPositions, owner ...*compileState) {
	if err := compileOwnerError(owner); err != nil {
		return
	}

	for position, class := range source {
		if err := compileOwnerError(owner); err != nil {
			return
		}
		target[position] = class
	}
}

func particleNullable(particle xsd.Particle, owner ...*compileState) bool {
	if err := compileOwnerError(owner); err != nil {
		return false
	}

	if particle.MinOccurs == 0 {
		return true
	}
	if particle.Group == nil {
		return false
	}
	switch particle.Group.Compositor {
	case xsd.Choice:
		for _, child := range particle.Group.Particles {
			if err := compileOwnerError(owner); err != nil {
				return false
			}
			if particleNullable(child, owner...) {
				return true
			}
		}
		return false
	case xsd.Sequence, xsd.All:
		for _, child := range particle.Group.Particles {
			if err := compileOwnerError(owner); err != nil {
				return false
			}
			if !particleNullable(child, owner...) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

func nameClassesOverlap(left nameClass, right nameClass, targetNamespace string, owner ...*compileState) bool {
	if err := compileOwnerError(owner); err != nil {
		return false
	}

	if left.name != nil && right.name != nil {
		for _, leftName := range nameClassNames(left, owner...) {
			if err := compileOwnerError(owner); err != nil {
				return false
			}
			for _, rightName := range nameClassNames(right, owner...) {
				if err := compileOwnerError(owner); err != nil {
					return false
				}
				if leftName == rightName {
					return true
				}
			}
		}
		return false
	}
	if left.name != nil {
		for _, name := range nameClassNames(left, owner...) {
			if err := compileOwnerError(owner); err != nil {
				return false
			}
			if wildcardAllows(right.wildcard, name.Namespace, targetNamespace, owner...) {
				return true
			}
		}
		return false
	}
	if right.name != nil {
		for _, name := range nameClassNames(right, owner...) {
			if err := compileOwnerError(owner); err != nil {
				return false
			}
			if wildcardAllows(left.wildcard, name.Namespace, targetNamespace, owner...) {
				return true
			}
		}
		return false
	}
	candidates := []string{"", targetNamespace, "urn:xsd-wildcard-probe"}
	for _, wildcard := range []*xsd.Wildcard{left.wildcard, right.wildcard} {
		if err := compileOwnerError(owner); err != nil {
			return false
		}
		for _, namespace := range wildcard.Namespaces {
			if err := compileOwnerError(owner); err != nil {
				return false
			}
			if !strings.HasPrefix(namespace, "##") {
				candidates = append(candidates, namespace)
			}
		}
	}
	for _, namespace := range candidates {
		if err := compileOwnerError(owner); err != nil {
			return false
		}
		if wildcardAllows(left.wildcard, namespace, targetNamespace, owner...) &&
			wildcardAllows(right.wildcard, namespace, targetNamespace, owner...) {
			return true
		}
	}
	return false
}

func nameClassNames(class nameClass, owner ...*compileState) []xsd.QName {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	return append([]xsd.QName{*class.name}, class.alternatives...)
}

func wildcardAllows(wildcard *xsd.Wildcard, namespace string, targetNamespace string, owner ...*compileState) bool {
	if err := compileOwnerError(owner); err != nil {
		return false
	}

	if wildcard == nil {
		return false
	}
	for _, constraint := range wildcard.Namespaces {
		if err := compileOwnerError(owner); err != nil {
			return false
		}
		switch constraint {
		case "##any":
			return true
		case "##other":
			if namespace != "" && namespace != targetNamespace {
				return true
			}
		case "##local":
			if namespace == "" {
				return true
			}
		case "##targetNamespace":
			if namespace == targetNamespace {
				return true
			}
		default:
			if namespace == constraint {
				return true
			}
		}
	}
	return false
}

func validateWildcard(wildcard *xsd.Wildcard, owner ...*compileState) error {
	if err := compileOwnerError(owner); err != nil {
		return err
	}

	if wildcard == nil {
		return nil
	}
	if wildcard.ProcessContents != xsd.ProcessStrict &&
		wildcard.ProcessContents != xsd.ProcessLax &&
		wildcard.ProcessContents != xsd.ProcessSkip {
		return fmt.Errorf(
			"%w: wildcard has invalid processContents %q",
			ErrInvalidComponent,
			wildcard.ProcessContents,
		)
	}
	if len(wildcard.Namespaces) == 0 {
		return fmt.Errorf("%w: wildcard has no namespace constraint", ErrInvalidComponent)
	}
	seen := make(map[string]struct{}, len(wildcard.Namespaces))
	for _, namespace := range wildcard.Namespaces {
		if err := compileOwnerError(owner); err != nil {
			return err
		}
		if strings.HasPrefix(namespace, "##") && namespace != "##any" &&
			namespace != "##other" && namespace != "##local" &&
			namespace != "##targetNamespace" {
			return fmt.Errorf(
				"%w: wildcard has invalid namespace token %q",
				ErrInvalidComponent,
				namespace,
			)
		}
		if _, duplicate := seen[namespace]; duplicate {
			return fmt.Errorf("%w: wildcard namespace %q is duplicated", ErrInvalidComponent, namespace)
		}
		seen[namespace] = struct{}{}
	}
	if len(wildcard.Namespaces) > 1 {
		if _, any := seen["##any"]; any {
			return fmt.Errorf("%w: ##any must be the only wildcard namespace", ErrInvalidComponent)
		}
		if _, other := seen["##other"]; other {
			return fmt.Errorf("%w: ##other must be the only wildcard namespace", ErrInvalidComponent)
		}
	}
	return nil
}

func (s *compileState) validateAttributeUse(
	attribute xsd.AttributeUse,
	typeNamespace string,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	_ = typeNamespace
	if (attribute.Name == "") == (attribute.Ref.Local == "") {
		return fmt.Errorf(
			"%w: attribute use must have exactly one of name or ref",
			ErrInvalidComponent,
		)
	}
	if attribute.Ref.Local != "" &&
		(attribute.Type.Local != "" || attribute.InlineSimpleType != nil) {
		return fmt.Errorf(
			"%w: referenced attribute use cannot define a type",
			ErrInvalidComponent,
		)
	}
	if attributeDefaultSet(attribute) && attributeFixedSet(attribute) {
		return fmt.Errorf(
			"%w: attribute use default and fixed are mutually exclusive",
			ErrInvalidComponent,
		)
	}
	if attributeDefaultSet(attribute) && attribute.Use != xsd.AttributeOptional {
		return fmt.Errorf(
			"%w: attribute default requires optional use",
			ErrInvalidComponent,
		)
	}
	if attribute.Ref.Local != "" {
		if _, ok := s.attributes[attribute.Ref]; !ok {
			return unresolvedComponent("attribute", attribute.Ref)
		}
		return s.validateAttributeUseValueConstraint(attribute)
	}
	if attribute.Type.Local != "" && attribute.InlineSimpleType != nil {
		return fmt.Errorf(
			"%w: attribute use has more than one type definition",
			ErrInvalidComponent,
		)
	}
	if attribute.InlineSimpleType != nil {
		if err := s.validateSimpleTypeDefinition(*attribute.InlineSimpleType); err != nil {
			return err
		}
	}
	if attribute.Type.Local != "" && !s.typeExists(attribute.Type, "simple") {
		return unresolvedComponent("simple type", attribute.Type)
	}
	if directNotationType(attribute.Type) {
		return fmt.Errorf("%w: NOTATION cannot be used directly", ErrInvalidComponent)
	}
	return s.validateAttributeUseValueConstraint(attribute)
}

func directNotationType(name xsd.QName) bool {
	return name.Namespace == xsd.Namespace && name.Local == "NOTATION"
}

func (s *compileState) typeExists(name xsd.QName, requiredKind string) bool {
	if err := s.contextError(); err != nil {
		return false
	}

	kind := ""
	if name.Namespace == xsd.Namespace {
		kind = builtInTypes[name.Local]
	} else if stored, ok := s.typeKinds[name]; ok {
		if strings.HasPrefix(stored, "simple") {
			kind = "simple"
		} else {
			kind = "complex"
		}
	}
	return kind != "" && (requiredKind == "" || kind == requiredKind)
}

func invalidComponent(kind string, name xsd.QName, reason string) error {
	return fmt.Errorf(
		"%w: %s {%s}%s: %s",
		ErrInvalidComponent,
		kind,
		name.Namespace,
		name.Local,
		reason,
	)
}

func unresolvedComponent(kind string, name xsd.QName) error {
	return fmt.Errorf(
		"%w: %s {%s}%s",
		ErrUnresolvedComponent,
		kind,
		name.Namespace,
		name.Local,
	)
}

type resourceDocument struct {
	document *xsd.Document
}

type instanceKey struct {
	uri       string
	namespace string
}

type compileState struct {
	particleCopies    int
	copyError         error
	ctx               context.Context
	compiler          *Compiler
	resources         map[string]resourceDocument
	instances         map[instanceKey]*Document
	bytes             int64
	references        int
	components        int
	elements          map[xsd.QName]xsd.Element
	attributes        map[xsd.QName]xsd.Attribute
	simpleTypes       map[xsd.QName]xsd.SimpleType
	complexTypes      map[xsd.QName]xsd.ComplexType
	modelGroups       map[xsd.QName]xsd.ModelGroupDefinition
	attributeGroups   map[xsd.QName]xsd.AttributeGroup
	notations         map[xsd.QName]xsd.Notation
	substitutionHeads map[xsd.QName]xsd.QName
	typeKinds         map[xsd.QName]string
}

// Value-space and copy helpers stop work without publishing partial values;
// Compile gives the caller's context error precedence at the publication owner.
func (s *compileState) contextError() error {
	if s == nil {
		return nil
	}
	if s.ctx != nil {
		if err := s.ctx.Err(); err != nil {
			return err
		}
	}
	return s.copyError
}

// Admission precedes each compiler-owned particle slice allocation. Subtraction
// keeps cumulative accounting overflow-safe without inspecting expanded trees.
// Set accessors deliberately have no invocation owner or copy-work policy.
func admitParticleCopies(count int, owner []*compileState) bool {
	if len(owner) == 0 || owner[0] == nil {
		return true
	}
	s := owner[0]
	if s.contextError() != nil {
		return false
	}
	limit := defaultMaxParticleCopies
	if s.compiler != nil && s.compiler.limits.MaxParticleCopies > 0 {
		limit = s.compiler.limits.MaxParticleCopies
	}
	if count > limit-s.particleCopies {
		s.copyError = fmt.Errorf("%w: particle copies exceed %d", ErrLimitExceeded, limit)
		return false
	}
	s.particleCopies += count
	return true
}

// Shared copy helpers also serve context-free immutable Set accessors.
func compileOwnerError(owner []*compileState) error {
	if len(owner) == 0 {
		return nil
	}
	return owner[0].contextError()
}

func (s *compileState) compileDocument(
	ctx context.Context,
	identity string,
	effectiveNamespace string,
	depth int,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	if err := ctx.Err(); err != nil {
		return err
	}
	if s.ctx == nil {
		s.ctx = ctx
	}
	if depth > s.compiler.limits.MaxDepth {
		return fmt.Errorf("%w: schema depth exceeds %d", ErrLimitExceeded, s.compiler.limits.MaxDepth)
	}
	resource := s.resources[identity]
	namespace := resource.document.TargetNamespace
	chameleon := false
	if namespace == "" {
		chameleon = effectiveNamespace != ""
	}
	if chameleon {
		namespace = effectiveNamespace
	}
	key := instanceKey{uri: identity, namespace: namespace}
	if _, ok := s.instances[key]; ok {
		return nil
	}
	if len(s.instances) >= s.compiler.limits.MaxSchemas {
		return fmt.Errorf("%w: schema count exceeds %d", ErrLimitExceeded, s.compiler.limits.MaxSchemas)
	}
	compiled := &Document{URI: identity, Namespace: namespace, Chameleon: chameleon}
	s.instances[key] = compiled
	if err := s.indexComponents(resource.document, namespace, chameleon); err != nil {
		return err
	}

	for _, reference := range resource.document.References {
		if err := s.contextError(); err != nil {
			return err
		}
		if compilableReference(reference) {
			if err := s.compileReference(ctx, compiled, resource.document, reference, namespace, chameleon, depth); err != nil {
				return err
			}
		}
	}
	return nil
}

func compilableReference(reference xsd.SchemaReference) bool {
	if reference.URI != "" {
		return true
	}
	return reference.Kind == xsd.ReferenceImport
}

func (s *compileState) compileReference(
	ctx context.Context,
	compiled *Document,
	document *xsd.Document,
	reference xsd.SchemaReference,
	namespace string,
	chameleon bool,
	depth int,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	s.references++
	if s.references > s.compiler.limits.MaxReferences {
		return fmt.Errorf("%w: reference count exceeds %d", ErrLimitExceeded, s.compiler.limits.MaxReferences)
	}
	referenced, resolvedIdentity, err := s.load(ctx, reference)
	if err != nil {
		if reference.URI == "" {
			if errors.Is(err, resolve.ErrNotFound) || errors.Is(err, resolve.ErrAccessDenied) {
				return nil
			}
		}
		return err
	}
	resolvedReference := reference
	resolvedReference.URI = resolvedIdentity
	compiled.Dependencies = append(compiled.Dependencies, resolvedIdentity)
	childNamespace, err := requiredNamespace(resolvedReference, namespace, referenced.TargetNamespace)
	if err != nil {
		return err
	}
	if err := s.compileDocument(ctx, resolvedIdentity, childNamespace, compileChildDepth(depth)); err != nil {
		return err
	}
	if reference.Kind == xsd.ReferenceRedefine {
		for _, redefinition := range document.Redefinitions {
			if err := s.contextError(); err != nil {
				return err
			}
			if redefinition.Reference.URI == reference.URI {
				if err := s.applyRedefinition(redefinition, document, namespace, chameleon); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (s *compileState) applyRedefinition(
	redefinition xsd.Redefinition,
	document *xsd.Document,
	namespace string,
	chameleon bool,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	for _, definition := range redefinition.SimpleTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		name := xsd.QName{Namespace: namespace, Local: definition.Name}
		original, ok := s.simpleTypes[name]
		if !ok {
			return unresolvedComponent("redefined simple type", name)
		}
		definition = normalizeInlineSimpleType(definition, namespace, chameleon, s)
		if definition.Variety != xsd.SimpleRestriction {
			return invalidComponent("redefined simple type", name, "must restrict itself")
		}
		if definition.Base != name {
			return invalidComponent("redefined simple type", name, "must restrict itself")
		}
		definition.Base = original.Base
		definition.InlineBase = original.InlineBase
		definition.Facets = append(append([]xsd.Facet(nil), original.Facets...), definition.Facets...)
		s.simpleTypes[name] = definition
	}
	for _, definition := range redefinition.ModelGroups {
		if err := s.contextError(); err != nil {
			return err
		}
		name := xsd.QName{Namespace: namespace, Local: definition.Name}
		original, ok := s.modelGroups[name]
		if !ok {
			return unresolvedComponent("redefined model group", name)
		}
		definition.Content = cloneModelGroup(definition.Content, s)
		normalizeModelGroup(
			definition.Content,
			document.ElementFormDefault,
			document.AttributeFormDefault,
			namespace,
			chameleon,
			s)
		replaceRedefinedGroupRefs(definition.Content, name, original.Content, s)
		s.modelGroups[name] = definition
	}
	for _, definition := range redefinition.AttributeGroups {
		if err := s.contextError(); err != nil {
			return err
		}
		name := xsd.QName{Namespace: namespace, Local: definition.Name}
		original, ok := s.attributeGroups[name]
		if !ok {
			return unresolvedComponent("redefined attribute group", name)
		}
		definition = normalizeAttributeGroup(
			definition,
			document.AttributeFormDefault,
			namespace,
			chameleon,
			s)
		foundSelf := false
		refs := make([]xsd.QName, 0, len(definition.References))
		for _, reference := range definition.References {
			if err := s.contextError(); err != nil {
				return err
			}
			if reference == name {
				foundSelf = true
				definition.Attributes = append(
					append([]xsd.AttributeUse(nil), original.Attributes...),
					definition.Attributes...,
				)
				if definition.Wildcard == nil {
					definition.Wildcard = cloneWildcard(original.Wildcard, s)
				}
			} else {
				refs = append(refs, reference)
			}
		}
		if !foundSelf {
			return invalidComponent("redefined attribute group", name, "must reference itself")
		}
		definition.References = refs
		s.attributeGroups[name] = definition
	}
	for _, definition := range redefinition.ComplexTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		name := xsd.QName{Namespace: namespace, Local: definition.Name}
		original, ok := s.complexTypes[name]
		if !ok {
			return unresolvedComponent("redefined complex type", name)
		}
		definition = normalizeComplexType(
			definition,
			document.ElementFormDefault,
			document.AttributeFormDefault,
			namespace,
			chameleon,
			s)
		if definition.Base != name {
			return invalidComponent("redefined complex type", name, "must derive from itself")
		}
		if definition.Derivation != xsd.DerivationExtension &&
			definition.Derivation != xsd.DerivationRestriction {
			return invalidComponent("redefined complex type", name, "must derive from itself")
		}
		if definition.Derivation == xsd.DerivationExtension {
			definition.Content = extendContent(original.Content, definition.Content, s)
			definition.Attributes = append(
				append([]xsd.AttributeUse(nil), original.Attributes...),
				definition.Attributes...,
			)
			definition.AttributeGroupRefs = append(
				append([]xsd.QName(nil), original.AttributeGroupRefs...),
				definition.AttributeGroupRefs...,
			)
			if definition.AttributeWildcard == nil {
				definition.AttributeWildcard = cloneWildcard(original.AttributeWildcard, s)
			}
		} else if err := s.validateComplexRestriction(definition, original); err != nil {
			return invalidComponent("redefined complex type", name, err.Error())
		}
		definition.Base = original.Base
		definition.Derivation = original.Derivation
		s.complexTypes[name] = definition
	}
	return nil
}

func replaceRedefinedGroupRefs(
	group *xsd.ModelGroup,
	name xsd.QName,
	original *xsd.ModelGroup,

	owner ...*compileState,
) {
	if err := compileOwnerError(owner); err != nil {
		return
	}

	if group == nil {
		return
	}
	for index := range group.Particles {
		if err := compileOwnerError(owner); err != nil {
			return
		}
		particle := &group.Particles[index]
		if particle.GroupRef == name {
			particle.GroupRef = xsd.QName{}
			particle.Group = cloneModelGroup(original, owner...)
		} else {
			replaceRedefinedGroupRefs(particle.Group, name, original, owner...)
		}
	}
}

func (s *compileState) indexComponents(
	document *xsd.Document,
	namespace string,
	chameleon bool,
) error {
	if err := s.contextError(); err != nil {
		return err
	}

	for _, notation := range document.Notations {
		if err := s.contextError(); err != nil {
			return err
		}
		name, err := s.componentName(namespace, notation.Name, "notation")
		if err != nil {
			return err
		}
		if _, exists := s.notations[name]; exists {
			return duplicateComponent("notation", name)
		}
		notation.Annotation = cloneAnnotation(notation.Annotation, s)
		s.notations[name] = notation
	}
	for _, element := range document.Elements {
		if err := s.contextError(); err != nil {
			return err
		}
		element = cloneElement(element, s)
		if element.Block.String() == "" {
			element.Block = document.BlockDefault
		}
		if element.Final.String() == "" {
			element.Final = document.FinalDefault
		}
		element = normalizeElementTypes(
			element,
			document.ElementFormDefault,
			document.AttributeFormDefault,
			namespace,
			chameleon,
			s)
		name, err := s.componentName(namespace, element.Name, "element")
		if err != nil {
			return err
		}
		if _, exists := s.elements[name]; exists {
			return duplicateComponent("element", name)
		}
		if chameleon {
			element.Type = adoptNamespace(element.Type, namespace)
			element.Ref = adoptNamespace(element.Ref, namespace)
			element.SubstitutionGroup = adoptNamespace(
				element.SubstitutionGroup,
				namespace,
			)
			for index := range element.IdentityConstraints {
				if err := s.contextError(); err != nil {
					return err
				}
				element.IdentityConstraints[index].Refer = adoptNamespace(
					element.IdentityConstraints[index].Refer,
					namespace,
				)
			}
		}
		s.elements[name] = element
	}
	for _, attribute := range document.Attributes {
		if err := s.contextError(); err != nil {
			return err
		}
		attribute = cloneAttribute(attribute, s)
		name, err := s.componentName(namespace, attribute.Name, "attribute")
		if err != nil {
			return err
		}
		if _, exists := s.attributes[name]; exists {
			return duplicateComponent("attribute", name)
		}
		if chameleon {
			attribute.Type = adoptNamespace(attribute.Type, namespace)
		}
		if attribute.InlineSimpleType != nil {
			typeDefinition := normalizeInlineSimpleType(
				*attribute.InlineSimpleType,
				namespace,
				chameleon,
				s)
			attribute.InlineSimpleType = &typeDefinition
		}
		s.attributes[name] = attribute
	}
	for _, simpleType := range document.SimpleTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		simpleType = cloneSimpleType(simpleType, s)
		name, err := s.componentName(namespace, simpleType.Name, "type")
		if err != nil {
			return err
		}
		if kind, exists := s.typeKinds[name]; exists {
			return duplicateComponent(kind+" and simple type", name)
		}
		if chameleon {
			simpleType.Base = adoptNamespace(simpleType.Base, namespace)
			simpleType.ItemType = adoptNamespace(simpleType.ItemType, namespace)
			for index := range simpleType.MemberTypes {
				if err := s.contextError(); err != nil {
					return err
				}
				simpleType.MemberTypes[index] = adoptNamespace(
					simpleType.MemberTypes[index],
					namespace,
				)
			}
		}
		s.typeKinds[name] = "simple type"
		s.simpleTypes[name] = cloneSimpleType(simpleType, s)
	}
	for _, complexType := range document.ComplexTypes {
		if err := s.contextError(); err != nil {
			return err
		}
		name, err := s.componentName(namespace, complexType.Name, "type")
		if err != nil {
			return err
		}
		if kind, exists := s.typeKinds[name]; exists {
			return duplicateComponent(kind+" and complex type", name)
		}
		if complexType.Block.String() == "" {
			complexType.Block = document.BlockDefault
		}
		if complexType.Final.String() == "" {
			complexType.Final = document.FinalDefault
		}
		complexType = normalizeComplexType(
			complexType,
			document.ElementFormDefault,
			document.AttributeFormDefault,
			namespace,
			chameleon,
			s)
		s.typeKinds[name] = "complex type"
		s.complexTypes[name] = complexType
	}
	for _, group := range document.ModelGroups {
		if err := s.contextError(); err != nil {
			return err
		}
		name, err := s.componentName(namespace, group.Name, "model group")
		if err != nil {
			return err
		}
		if _, exists := s.modelGroups[name]; exists {
			return duplicateComponent("model group", name)
		}
		group.Content = cloneModelGroup(group.Content, s)
		normalizeModelGroup(
			group.Content,
			document.ElementFormDefault,
			document.AttributeFormDefault,
			namespace,
			chameleon,
			s)
		s.modelGroups[name] = group
	}
	for _, group := range document.AttributeGroups {
		if err := s.contextError(); err != nil {
			return err
		}
		name, err := s.componentName(namespace, group.Name, "attribute group")
		if err != nil {
			return err
		}
		if _, exists := s.attributeGroups[name]; exists {
			return duplicateComponent("attribute group", name)
		}
		group = normalizeAttributeGroup(
			group,
			document.AttributeFormDefault,
			namespace,
			chameleon,
			s)
		s.attributeGroups[name] = group
	}
	return nil
}

func normalizeComplexType(
	complexType xsd.ComplexType,
	elementDefault xsd.Form,
	attributeDefault xsd.Form,
	namespace string,
	chameleon bool,

	owner ...*compileState,
) xsd.ComplexType {
	if err := compileOwnerError(owner); err != nil {
		return xsd.ComplexType{}
	}

	complexType = cloneComplexType(complexType, owner...)
	if complexType.InlineSimpleType != nil {
		typeDefinition := normalizeInlineSimpleType(
			*complexType.InlineSimpleType,
			namespace,
			chameleon,
			owner...)
		complexType.InlineSimpleType = &typeDefinition
	}
	normalizeModelGroup(
		complexType.Content,
		elementDefault,
		attributeDefault,
		namespace,
		chameleon,
		owner...)
	complexType.Attributes = normalizeAttributeUses(
		complexType.Attributes,
		attributeDefault,
		namespace,
		chameleon,
		owner...)
	if chameleon {
		for index := range complexType.AttributeGroupRefs {
			if err := compileOwnerError(owner); err != nil {
				return xsd.ComplexType{}
			}
			complexType.AttributeGroupRefs[index] = adoptNamespace(
				complexType.AttributeGroupRefs[index],
				namespace,
			)
		}
		complexType.Base = adoptNamespace(complexType.Base, namespace)
	}
	return complexType
}

func normalizeElementTypes(
	element xsd.Element,
	elementDefault xsd.Form,
	attributeDefault xsd.Form,
	namespace string,
	chameleon bool,

	owner ...*compileState,
) xsd.Element {
	if err := compileOwnerError(owner); err != nil {
		return xsd.Element{}
	}

	element.TargetNamespace = namespace
	for index := range element.IdentityConstraints {
		if err := compileOwnerError(owner); err != nil {
			return xsd.Element{}
		}
		element.IdentityConstraints[index].TargetNamespace = namespace
		if chameleon {
			element.IdentityConstraints[index].Refer = adoptNamespace(
				element.IdentityConstraints[index].Refer,
				namespace,
			)
		}
	}
	if element.InlineSimpleType != nil {
		typeDefinition := normalizeInlineSimpleType(
			*element.InlineSimpleType,
			namespace,
			chameleon,
			owner...)
		element.InlineSimpleType = &typeDefinition
	}
	if element.InlineComplexType != nil {
		typeDefinition := normalizeComplexType(
			*element.InlineComplexType,
			elementDefault,
			attributeDefault,
			namespace,
			chameleon,
			owner...)
		element.InlineComplexType = &typeDefinition
	}
	return element
}

func normalizeInlineSimpleType(
	typeDefinition xsd.SimpleType,
	namespace string,
	chameleon bool,

	owner ...*compileState,
) xsd.SimpleType {
	if err := compileOwnerError(owner); err != nil {
		return xsd.SimpleType{}
	}

	typeDefinition = cloneSimpleType(typeDefinition, owner...)
	if typeDefinition.InlineBase != nil {
		base := normalizeInlineSimpleType(*typeDefinition.InlineBase, namespace, chameleon, owner...)
		typeDefinition.InlineBase = &base
	}
	if typeDefinition.InlineItem != nil {
		item := normalizeInlineSimpleType(*typeDefinition.InlineItem, namespace, chameleon, owner...)
		typeDefinition.InlineItem = &item
	}
	for index := range typeDefinition.InlineMembers {
		if err := compileOwnerError(owner); err != nil {
			return xsd.SimpleType{}
		}
		typeDefinition.InlineMembers[index] = normalizeInlineSimpleType(
			typeDefinition.InlineMembers[index], namespace, chameleon,
			owner...)
	}
	if chameleon {
		typeDefinition.Base = adoptNamespace(typeDefinition.Base, namespace)
		typeDefinition.ItemType = adoptNamespace(typeDefinition.ItemType, namespace)
		for index := range typeDefinition.MemberTypes {
			if err := compileOwnerError(owner); err != nil {
				return xsd.SimpleType{}
			}
			typeDefinition.MemberTypes[index] = adoptNamespace(
				typeDefinition.MemberTypes[index],
				namespace,
			)
		}
	}
	return typeDefinition
}

func normalizeAttributeUses(
	attributes []xsd.AttributeUse,
	attributeDefault xsd.Form,
	namespace string,
	chameleon bool,

	owner ...*compileState,
) []xsd.AttributeUse {
	if err := compileOwnerError(owner); err != nil {
		return nil
	}

	attributes = cloneAttributeUses(attributes, owner...)
	for index := range attributes {
		if err := compileOwnerError(owner); err != nil {
			return nil
		}
		attribute := &attributes[index]
		if attribute.Form == "" {
			attribute.Form = attributeDefault
		}
		if attribute.Form == xsd.FormQualified {
			attribute.Namespace = namespace
		} else {
			attribute.Namespace = ""
		}
		if chameleon {
			attribute.Ref = adoptNamespace(attribute.Ref, namespace)
			attribute.Type = adoptNamespace(attribute.Type, namespace)
		}
		if attribute.InlineSimpleType != nil {
			typeDefinition := normalizeInlineSimpleType(
				*attribute.InlineSimpleType,
				namespace,
				chameleon,
				owner...)
			attribute.InlineSimpleType = &typeDefinition
		}
	}
	return attributes
}

func normalizeAttributeGroup(
	group xsd.AttributeGroup,
	attributeDefault xsd.Form,
	namespace string,
	chameleon bool,

	owner ...*compileState,
) xsd.AttributeGroup {
	if err := compileOwnerError(owner); err != nil {
		return xsd.AttributeGroup{}
	}

	group = cloneAttributeGroup(group, owner...)
	group.Attributes = normalizeAttributeUses(
		group.Attributes,
		attributeDefault,
		namespace,
		chameleon,
		owner...)
	if chameleon {
		for index := range group.References {
			if err := compileOwnerError(owner); err != nil {
				return xsd.AttributeGroup{}
			}
			group.References[index] = adoptNamespace(group.References[index], namespace)
		}
	}
	return group
}

func normalizeModelGroup(
	group *xsd.ModelGroup,
	elementDefault xsd.Form,
	attributeDefault xsd.Form,
	namespace string,
	chameleon bool,

	owner ...*compileState,
) {
	if err := compileOwnerError(owner); err != nil {
		return
	}

	if group == nil {
		return
	}
	for index := range group.Particles {
		if err := compileOwnerError(owner); err != nil {
			return
		}
		particle := &group.Particles[index]
		if particle.Element != nil {
			if particle.Element.Ref.Local != "" {
				particle.Element.Namespace = ""
			} else {
				if particle.Element.Form == "" {
					particle.Element.Form = elementDefault
				}
				if particle.Element.Form == xsd.FormQualified {
					particle.Element.Namespace = namespace
				} else {
					particle.Element.Namespace = ""
				}
			}
			if chameleon {
				particle.Element.Ref = adoptNamespace(particle.Element.Ref, namespace)
				particle.Element.Type = adoptNamespace(particle.Element.Type, namespace)
				for index := range particle.Element.IdentityConstraints {
					if err := compileOwnerError(owner); err != nil {
						return
					}
					constraint := &particle.Element.IdentityConstraints[index]
					constraint.Refer = adoptNamespace(constraint.Refer, namespace)
				}
			}
			*particle.Element = normalizeElementTypes(
				*particle.Element,
				elementDefault,
				attributeDefault,
				namespace,
				chameleon,
				owner...)
		}
		if chameleon {
			particle.GroupRef = adoptNamespace(particle.GroupRef, namespace)
		}
		normalizeModelGroup(
			particle.Group,
			elementDefault,
			attributeDefault,
			namespace,
			chameleon,
			owner...)
	}
}

func (s *compileState) componentName(
	namespace string,
	local string,
	kind string,
) (xsd.QName, error) {
	if err := s.contextError(); err != nil {
		return xsd.QName{}, err
	}

	if local == "" {
		return xsd.QName{}, fmt.Errorf("%w: global %s has no name", ErrInvalidComponent, kind)
	}
	s.components++
	if s.components > s.compiler.limits.MaxComponents {
		return xsd.QName{}, fmt.Errorf(
			"%w: component count exceeds %d",
			ErrLimitExceeded,
			s.compiler.limits.MaxComponents,
		)
	}
	return xsd.QName{Namespace: namespace, Local: local}, nil
}

func duplicateComponent(kind string, name xsd.QName) error {
	return fmt.Errorf(
		"%w: %s {%s}%s",
		ErrDuplicateComponent,
		kind,
		name.Namespace,
		name.Local,
	)
}

func adoptNamespace(name xsd.QName, namespace string) xsd.QName {
	if name.Local != "" && name.Namespace == "" {
		name.Namespace = namespace
	}
	return name
}

func (s *compileState) load(
	ctx context.Context,
	reference xsd.SchemaReference,
) (*xsd.Document, string, error) {
	if err := s.contextError(); err != nil {
		return nil, "", err
	}

	if err := s.compiler.admitURI(reference.URI); err != nil {
		return nil, "", err
	}
	if cached, ok := s.resources[reference.URI]; ok {
		return cached.document, reference.URI, nil
	}
	resource, err := s.compiler.resolver.Resolve(ctx, resolve.Request{
		URI:       reference.URI,
		Namespace: reference.Namespace,
		Kind:      resolveKind(reference.Kind),
	})
	if err != nil {
		return nil, "", err
	}
	if err := s.contextError(); err != nil {
		return nil, "", err
	}
	if err := s.compiler.admitURI(resource.URI); err != nil {
		return nil, "", err
	}
	if reference.URI != "" && resource.URI != reference.URI {
		return nil, "", fmt.Errorf(
			"%w: requested %q, received %q",
			ErrResourceIdentity,
			reference.URI,
			resource.URI,
		)
	}
	if reference.URI == "" {
		if err := validateIdentity(resource.URI); err != nil {
			return nil, "", fmt.Errorf("%w: %v", ErrResourceIdentity, err)
		}
		if cached, ok := s.resources[resource.URI]; ok {
			return cached.document, resource.URI, nil
		}
	}
	if s.bytes+int64(len(resource.Content)) > s.compiler.limits.MaxBytes {
		return nil, "", fmt.Errorf("%w: schema bytes exceed %d", ErrLimitExceeded, s.compiler.limits.MaxBytes)
	}
	s.bytes += int64(len(resource.Content))
	document, err := xsd.Parse(ctx, resource.Content, xsd.ParseOptions{
		SystemID:            resource.URI,
		MaxDocumentBytes:    s.compiler.limits.MaxBytes,
		MaxNamespaceEntries: s.compiler.limits.MaxParseNamespaceEntries,
		MaxModelBytes:       s.compiler.limits.MaxParseModelBytes,
	})
	if err != nil {
		return nil, "", err
	}
	s.resources[resource.URI] = resourceDocument{document: document}
	return document, resource.URI, nil
}

func requiredNamespace(
	reference xsd.SchemaReference,
	parentNamespace string,
	actualNamespace string,
) (string, error) {
	switch reference.Kind {
	case xsd.ReferenceInclude, xsd.ReferenceRedefine:
		if actualNamespace != "" && actualNamespace != parentNamespace {
			return "", fmt.Errorf(
				"%w: %s has %q, expected %q",
				ErrNamespace,
				reference.URI,
				actualNamespace,
				parentNamespace,
			)
		}
		return parentNamespace, nil
	case xsd.ReferenceImport:
		if actualNamespace != reference.Namespace {
			return "", fmt.Errorf(
				"%w: %s has %q, import requested %q",
				ErrNamespace,
				reference.URI,
				actualNamespace,
				reference.Namespace,
			)
		}
		return actualNamespace, nil
	default:
		return "", fmt.Errorf("xsd compile: unknown reference kind %q", reference.Kind)
	}
}

func resolveKind(kind xsd.ReferenceKind) resolve.Kind {
	switch kind {
	case xsd.ReferenceInclude:
		return resolve.KindInclude
	case xsd.ReferenceImport:
		return resolve.KindImport
	case xsd.ReferenceRedefine:
		return resolve.KindRedefine
	default:
		return ""
	}
}

func (c *Compiler) admitURI(identity string) error {
	if int64(len(identity)) > c.limits.MaxURIBytes {
		return fmt.Errorf("%w: resource URI exceeds %d bytes", ErrLimitExceeded, c.limits.MaxURIBytes)
	}
	return nil
}

func validateIdentity(identity string) error {
	uri, err := url.Parse(identity)
	if err != nil || !uri.IsAbs() || uri.Fragment != "" {
		return fmt.Errorf("xsd compile: invalid resource URI %q", identity)
	}
	return nil
}
