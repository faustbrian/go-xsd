package xsd

import (
	"sort"
	"strings"
)

// Namespace is the XML Schema 1.0 schema vocabulary namespace.
const Namespace = "http://www.w3.org/2001/XMLSchema"

// Form controls whether locally declared element or attribute names are
// namespace-qualified in instance documents.
type Form string

const (
	// FormUnqualified leaves local names outside the target namespace.
	FormUnqualified Form = "unqualified"
	// FormQualified places local names in the target namespace.
	FormQualified Form = "qualified"
)

// Derivation is an XML Schema derivation or substitution control token.
type Derivation string

const (
	// DerivationExtension identifies extension from a base type.
	DerivationExtension Derivation = "extension"
	// DerivationRestriction identifies restriction of a base type.
	DerivationRestriction Derivation = "restriction"
	// DerivationSubstitution controls element substitution.
	DerivationSubstitution Derivation = "substitution"
	// DerivationList identifies list construction from an item type.
	DerivationList Derivation = "list"
	// DerivationUnion identifies union construction from member types.
	DerivationUnion Derivation = "union"
)

// DerivationSet represents blockDefault and finalDefault values.
// The zero value is an empty set.
type DerivationSet struct {
	all    bool
	values map[Derivation]struct{}
}

// All reports whether the lexical value was #all.
func (s DerivationSet) All() bool { return s.all }

// Contains reports whether a derivation is in the set. #all contains every
// derivation.
func (s DerivationSet) Contains(value Derivation) bool {
	if s.all {
		return true
	}
	_, ok := s.values[value]
	return ok
}

// String returns the deterministic XML Schema lexical representation.
func (s DerivationSet) String() string {
	if s.all {
		return "#all"
	}
	values := make([]string, 0, len(s.values))
	for value := range s.values {
		values = append(values, string(value))
	}
	sort.Strings(values)
	return strings.Join(values, " ")
}

// Annotation represents xs:annotation content retained by the parser.
type Annotation struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Documentation retains documentation entries in source order.
	Documentation []Documentation
	// AppInformation retains appinfo entries in source order.
	AppInformation []AppInfo
}

// AppInfo preserves machine-readable xs:appinfo content.
type AppInfo struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Source preserves the optional source URI text without fetching it.
	Source string
	// Content is retained appinfo markup.
	Content string
}

// Documentation represents human-readable xs:documentation content.
type Documentation struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Source preserves the optional source URI text without fetching it.
	Source string
	// Language preserves the optional xml:lang value.
	Language string
	// Content is extracted and trimmed documentation text, not the retained markup.
	Content string
	// Markup retains the documentation inner XML independently of extracted text.
	Markup string
}

// Notation is a global xs:notation declaration.
type Notation struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Name is the declared local name.
	Name string
	// Public is the notation public identifier.
	Public string
	// System is the optional notation system identifier, not a fetched resource.
	System string
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// ReferenceKind identifies a schema composition directive.
type ReferenceKind string

const (
	// ReferenceInclude composes a same-namespace or chameleon schema.
	ReferenceInclude ReferenceKind = "include"
	// ReferenceImport composes declarations from the specified namespace.
	ReferenceImport ReferenceKind = "import"
	// ReferenceRedefine composes a schema with overriding declarations.
	ReferenceRedefine ReferenceKind = "redefine"
)

// SchemaReference preserves an include, import, or redefine in source order.
// Location is the lexical schemaLocation; URI is resolved against xml:base and
// the document SystemID without loading it.
type SchemaReference struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Kind selects the declaration or composition operation.
	Kind ReferenceKind
	// Namespace is the import namespace, including empty for a no-namespace import.
	Namespace string
	// Location is the lexical schemaLocation, not a loaded resource.
	Location string
	// URI is the resolved composition identity without external loading.
	URI string
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// QName is an expanded XML qualified name.
type QName struct {
	// Namespace is the namespace URI; prefixes are not part of identity.
	Namespace string
	// Local is the unprefixed local part of the expanded name.
	Local string
}

// IdentityKind identifies an xs:unique, xs:key, or xs:keyref constraint.
type IdentityKind string

const (
	// IdentityUnique requires distinct complete field tuples where fields are
	// present.
	IdentityUnique IdentityKind = "unique"
	// IdentityKey requires present, distinct field tuples.
	IdentityKey IdentityKind = "key"
	// IdentityKeyRef requires field tuples to match the referenced key or unique
	// constraint.
	IdentityKeyRef IdentityKind = "keyref"
)

// IdentityConstraint preserves the XML Schema identity XPath subset.
type IdentityConstraint struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// TargetNamespace is the effective target namespace used during compilation.
	TargetNamespace string
	// Name is the declared local name.
	Name string
	// Kind selects unique, key or keyref tuple semantics.
	Kind IdentityKind
	// Refer is the expanded key/unique name referenced by a keyref.
	Refer QName
	// Selector is the identity-constraint XPath subset expression.
	Selector string
	// SelectorID preserves the selector element id.
	SelectorID string
	// Fields retains field XPath expressions in source order.
	Fields []string
	// FieldIDs retains ids corresponding to Fields.
	FieldIDs []string
	// Namespaces retains prefix bindings needed to interpret lexical names.
	Namespaces map[string]string
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
	// SelectorAnnotation is the optional selector annotation.
	SelectorAnnotation *Annotation
	// FieldAnnotations retains optional annotations corresponding to Fields.
	FieldAnnotations []*Annotation
}

// Element is a global declaration or a local element term in a particle.
type Element struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Name is the declared local name.
	Name string
	// Namespace is the effective declaration namespace populated by compilation.
	Namespace string
	// Type is the expanded named type reference; zero when unspecified.
	Type QName
	// Ref is the expanded declaration reference; zero when absent.
	Ref QName
	// Form overrides local-name qualification; empty uses the document default.
	Form Form
	// Abstract prevents direct instance use of the declaration.
	Abstract bool
	// Nillable permits xsi:nil under the element constraints.
	Nillable bool
	// Default is the lexical default value, interpreted in ValueNamespaces.
	Default string
	// Fixed is the lexical fixed value, interpreted in ValueNamespaces.
	Fixed string
	// DefaultSet distinguishes an explicit empty default from no default attribute.
	DefaultSet bool
	// FixedSet distinguishes an explicit empty fixed value from no fixed attribute.
	FixedSet bool
	// ValueNamespaces retains the namespace context of QName-valued constraints.
	ValueNamespaces map[string]string
	// Block controls prohibited derivation or substitution methods.
	Block DerivationSet
	// Final controls prohibited derivation methods.
	Final DerivationSet
	// IdentityConstraints retains constraints attached to this element.
	IdentityConstraints []IdentityConstraint
	// InlineSimpleType is the optional anonymous simple type.
	InlineSimpleType *SimpleType
	// InlineComplexType is the optional anonymous complex type.
	InlineComplexType *ComplexType
	// TargetNamespace is the effective target namespace used during compilation.
	TargetNamespace string
	// SubstitutionGroup is the expanded global element head affiliation.
	SubstitutionGroup QName
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// Attribute is a global attribute declaration.
type Attribute struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Name is the declared local name.
	Name string
	// Type is the expanded named type reference; zero when unspecified.
	Type QName
	// Default is the lexical default value, interpreted in ValueNamespaces.
	Default string
	// Fixed is the lexical fixed value, interpreted in ValueNamespaces.
	Fixed string
	// DefaultSet distinguishes an explicit empty default from no default attribute.
	DefaultSet bool
	// FixedSet distinguishes an explicit empty fixed value from no fixed attribute.
	FixedSet bool
	// ValueNamespaces retains the namespace context of QName-valued constraints.
	ValueNamespaces map[string]string
	// InlineSimpleType is the optional anonymous simple type.
	InlineSimpleType *SimpleType
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// AttributeUseKind controls whether an attribute use is optional, required,
// or prohibited.
type AttributeUseKind string

const (
	// AttributeOptional permits omission of the attribute.
	AttributeOptional AttributeUseKind = "optional"
	// AttributeRequired requires the attribute to be present.
	AttributeRequired AttributeUseKind = "required"
	// AttributeProhibited forbids the attribute.
	AttributeProhibited AttributeUseKind = "prohibited"
)

// AttributeUse is an attribute use within a complex type.
type AttributeUse struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Name is the declared local name.
	Name string
	// Namespace is the effective attribute-use namespace populated by compilation.
	Namespace string
	// Ref is the expanded declaration reference; zero when absent.
	Ref QName
	// Type is the expanded named type reference; zero when unspecified.
	Type QName
	// Form overrides local-name qualification; empty uses the document default.
	Form Form
	// Use controls optional, required or prohibited attribute presence.
	Use AttributeUseKind
	// Default is the lexical default value, interpreted in ValueNamespaces.
	Default string
	// Fixed is the lexical fixed value, interpreted in ValueNamespaces.
	Fixed string
	// DefaultSet distinguishes an explicit empty default from no default attribute.
	DefaultSet bool
	// FixedSet distinguishes an explicit empty fixed value from no fixed attribute.
	FixedSet bool
	// ValueNamespaces retains the namespace context of QName-valued constraints.
	ValueNamespaces map[string]string
	// InlineSimpleType is the optional anonymous simple type.
	InlineSimpleType *SimpleType
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// Compositor identifies an XML Schema model group compositor.
type Compositor string

const (
	// Sequence orders child particles in declaration order.
	Sequence Compositor = "sequence"
	// Choice selects a child particle alternative.
	Choice Compositor = "choice"
	// All permits the supported all-group children in any order.
	All Compositor = "all"
)

// ModelGroup is a sequence, choice, or all group.
type ModelGroup struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Compositor selects sequence, choice or all semantics.
	Compositor Compositor
	// Particles retains child particles in declaration order.
	Particles []Particle
	// MinOccurs is the minimum occurrence count.
	MinOccurs uint64
	// MaxOccurs is the maximum count when Unbounded is false.
	MaxOccurs uint64
	// Unbounded ignores MaxOccurs in favor of an unbounded occurrence maximum.
	Unbounded bool
	// OccursSet distinguishes explicit group occurrence constraints from the
	// implicit one occurrence.
	OccursSet bool
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// Particle applies occurrence constraints to an element or nested group.
type Particle struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// MinOccurs is the minimum occurrence count.
	MinOccurs uint64
	// MaxOccurs is the maximum count when Unbounded is false.
	MaxOccurs uint64
	// Unbounded ignores MaxOccurs in favor of an unbounded occurrence maximum.
	Unbounded bool
	// Element is the optional element term; use one particle term only.
	Element *Element
	// Group is the optional inline model-group term.
	Group *ModelGroup
	// GroupRef is the expanded named model-group term.
	GroupRef QName
	// Wildcard is the optional wildcard term.
	Wildcard *Wildcard
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// ProcessContents controls validation for declarations matched by a wildcard.
type ProcessContents string

const (
	// ProcessStrict requires a declaration and validation for matched content.
	ProcessStrict ProcessContents = "strict"
	// ProcessLax validates matched content when a declaration is available.
	ProcessLax ProcessContents = "lax"
	// ProcessSkip does not validate matched wildcard content.
	ProcessSkip ProcessContents = "skip"
)

// Wildcard constrains namespaces and validation for xs:any or xs:anyAttribute.
type Wildcard struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Namespaces retains wildcard namespace tokens such as ##any, not prefix
	// bindings.
	Namespaces []string
	// ProcessContents selects strict, lax or skip processing for matched content.
	ProcessContents ProcessContents
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// ModelGroupDefinition is a named global model group.
type ModelGroupDefinition struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Name is the declared local name.
	Name string
	// Content is the named model-group body.
	Content *ModelGroup
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// AttributeGroup is a named global collection of attribute uses.
type AttributeGroup struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Name is the declared local name.
	Name string
	// Attributes retains attribute declarations or uses in source order.
	Attributes []AttributeUse
	// References retains expanded attribute-group references.
	References []QName
	// Wildcard is the optional anyAttribute constraint.
	Wildcard *Wildcard
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
	// AttributeGroupReferences retains annotated references alongside the
	// expanded-name projection.
	AttributeGroupReferences []AttributeGroupReference
}

// AttributeGroupReference preserves an attribute-group reference annotation.
type AttributeGroupReference struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Ref is the expanded declaration reference; zero when absent.
	Ref QName
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// SimpleVariety identifies restriction, list, or union construction.
type SimpleVariety string

const (
	// SimpleRestriction derives a type by constraining its base.
	SimpleRestriction SimpleVariety = "restriction"
	// SimpleList constructs whitespace-separated values of an item type.
	SimpleList SimpleVariety = "list"
	// SimpleUnion accepts values from member simple types.
	SimpleUnion SimpleVariety = "union"
)

// FacetKind identifies an XML Schema constraining facet.
type FacetKind string

const (
	// FacetLength requires the type-specific value length.
	FacetLength FacetKind = "length"
	// FacetMinLength bounds the minimum type-specific value length.
	FacetMinLength FacetKind = "minLength"
	// FacetMaxLength bounds the maximum type-specific value length.
	FacetMaxLength FacetKind = "maxLength"
	// FacetPattern constrains lexical content with the XSD regular-expression
	// dialect.
	FacetPattern FacetKind = "pattern"
	// FacetEnumeration admits values equal in the type value space.
	FacetEnumeration FacetKind = "enumeration"
	// FacetWhiteSpace controls preserve, replace or collapse normalization.
	FacetWhiteSpace FacetKind = "whiteSpace"
	// FacetMaxInclusive sets an inclusive upper value-space bound.
	FacetMaxInclusive FacetKind = "maxInclusive"
	// FacetMaxExclusive sets an exclusive upper value-space bound.
	FacetMaxExclusive FacetKind = "maxExclusive"
	// FacetMinInclusive sets an inclusive lower value-space bound.
	FacetMinInclusive FacetKind = "minInclusive"
	// FacetMinExclusive sets an exclusive lower value-space bound.
	FacetMinExclusive FacetKind = "minExclusive"
	// FacetTotalDigits bounds the decimal value-space digit count.
	FacetTotalDigits FacetKind = "totalDigits"
	// FacetFractionDigits bounds the decimal value-space fractional digit count.
	FacetFractionDigits FacetKind = "fractionDigits"
)

// Facet preserves one restriction facet in source order.
type Facet struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Kind selects the constraining facet applied to the base type.
	Kind FacetKind
	// Value is the facet lexical value, not yet normalized to the base value space.
	Value string
	// Fixed prevents further derivations from changing this facet.
	Fixed bool
	// Namespaces retains the namespace context for QName-valued facet values.
	Namespaces map[string]string
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
}

// SimpleType is a named or anonymous simple type definition.
type SimpleType struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Name is the declared local name.
	Name string
	// Variety selects restriction, list or union construction.
	Variety SimpleVariety
	// Final controls prohibited derivation methods.
	Final DerivationSet
	// Base is the expanded base type reference.
	Base QName
	// InlineBase is an optional anonymous restriction base.
	InlineBase *SimpleType
	// Facets retains constraining facets in source order.
	Facets []Facet
	// ItemType is the expanded list item-type reference.
	ItemType QName
	// InlineItem is an optional anonymous list item type.
	InlineItem *SimpleType
	// MemberTypes retains expanded union member-type references.
	MemberTypes []QName
	// InlineMembers retains anonymous union member types.
	InlineMembers []SimpleType
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
	// VarietyAnnotation is the optional annotation on restriction, list or union.
	VarietyAnnotation *Annotation
	// VarietyID preserves the id on restriction, list or union.
	VarietyID string
}

// ComplexType is a named or anonymous complex type definition.
type ComplexType struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// Name is the declared local name.
	Name string
	// Base is the expanded base type reference.
	Base QName
	// SimpleBase is the expanded simple-content base reference.
	SimpleBase QName
	// InlineSimpleType is the optional anonymous simple type.
	InlineSimpleType *SimpleType
	// SimpleFacets retains facets constraining simple content.
	SimpleFacets []Facet
	// Derivation selects extension or restriction of the base.
	Derivation Derivation
	// SimpleContent distinguishes simple content from a complex-content model.
	SimpleContent bool
	// Abstract prevents direct instance use of the declaration.
	Abstract bool
	// Mixed permits text alongside the model-group content.
	Mixed bool
	// MixedSet distinguishes an explicit mixed attribute from its default.
	MixedSet bool
	// Block controls prohibited extension/restriction when substituting a type.
	Block DerivationSet
	// Final controls prohibited derivation methods.
	Final DerivationSet
	// Content is the optional model-group content.
	Content *ModelGroup
	// Attributes retains attribute declarations or uses in source order.
	Attributes []AttributeUse
	// AttributeGroupRefs retains expanded attribute-group references.
	AttributeGroupRefs []QName
	// AttributeGroupReferences retains annotated references alongside the
	// expanded-name projection.
	AttributeGroupReferences []AttributeGroupReference
	// AttributeWildcard is the optional anyAttribute constraint.
	AttributeWildcard *Wildcard
	// Annotation is optional annotation content; nil means absent.
	Annotation *Annotation
	// ContentAnnotation is the optional simpleContent or complexContent annotation.
	ContentAnnotation *Annotation
	// DerivationAnnotation is the optional extension or restriction annotation.
	DerivationAnnotation *Annotation
	// ContentID preserves the simpleContent or complexContent id.
	ContentID string
	// DerivationID preserves the extension or restriction id.
	DerivationID string
}

// Document is a mutable XML Schema source model. Declarations retain source
// order for deterministic serialization; slices, maps and nested pointers are
// caller-owned, not immutable compiled components. Synchronize shared mutation.
type Document struct {
	// ID preserves the optional lexical schema id attribute.
	ID string
	// SystemID is the caller-supplied document identity used for base resolution.
	SystemID string
	// BaseURI is the document base after xml:base resolution.
	BaseURI string
	// TargetNamespace is the lexical targetNamespace; empty permits chameleon
	// composition.
	TargetNamespace string
	// Namespaces retains prefix bindings needed to interpret lexical names.
	Namespaces map[string]string
	// ElementFormDefault controls local element qualification when Form is empty.
	ElementFormDefault Form
	// AttributeFormDefault controls local attribute qualification when Form is
	// empty.
	AttributeFormDefault Form
	// BlockDefault supplies the default derivation/substitution blocking set.
	BlockDefault DerivationSet
	// FinalDefault supplies the default final derivation set.
	FinalDefault DerivationSet
	// Version preserves the schema version text without selecting an XSD edition.
	Version string
	// Language preserves the optional xml:lang value.
	Language string
	// Annotations retains top-level annotations in source order.
	Annotations []Annotation
	// References retains include, import and redefine directives in source order.
	References []SchemaReference
	// Redefinitions retains overriding declarations from redefine directives.
	Redefinitions []Redefinition
	// Elements retains global element declarations in source order.
	Elements []Element
	// Attributes retains global attribute declarations in source order.
	Attributes []Attribute
	// SimpleTypes retains named simple types in source order.
	SimpleTypes []SimpleType
	// ComplexTypes retains named complex types in source order.
	ComplexTypes []ComplexType
	// ModelGroups retains named model-group definitions in source order.
	ModelGroups []ModelGroupDefinition
	// AttributeGroups retains named attribute-group definitions in source order.
	AttributeGroups []AttributeGroup
	// Notations retains global notation declarations in source order.
	Notations []Notation
}

// Redefinition contains the declarations nested in xs:redefine.
type Redefinition struct {
	// Reference identifies the schema redefined by these declarations.
	Reference SchemaReference
	// SimpleTypes retains named simple types in source order.
	SimpleTypes []SimpleType
	// ComplexTypes retains named complex types in source order.
	ComplexTypes []ComplexType
	// ModelGroups retains named model-group definitions in source order.
	ModelGroups []ModelGroupDefinition
	// AttributeGroups retains named attribute-group definitions in source order.
	AttributeGroups []AttributeGroup
}
