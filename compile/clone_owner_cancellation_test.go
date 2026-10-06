package compile

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/resolve"
)

func TestCloneOwnersHonorCancellationWithoutChangingContextFreeCopies(t *testing.T) {
	const rootURI = "https://example.test/root.xsd"
	const schema = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:ordinary">
 <xs:include schemaLocation="child.xsd"/>
 <xs:element name="value" type="xs:string" default="ordinary"><xs:annotation><xs:documentation>original</xs:documentation><xs:appinfo>original</xs:appinfo></xs:annotation></xs:element>
 <xs:attribute name="label" type="xs:string" default="ordinary"><xs:annotation><xs:documentation>original</xs:documentation><xs:appinfo>original</xs:appinfo></xs:annotation></xs:attribute>
</xs:schema>`
	resolver, err := resolve.NewMemory(map[string][]byte{
		"https://example.test/child.xsd": []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`),
	})
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := New(Options{Resolver: resolver})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	parsed, err := xsd.Parse(ctx, []byte(schema), xsd.ParseOptions{SystemID: rootURI})
	if err != nil {
		t.Fatal(err)
	}
	state := emptyValidationState()
	state.compiler = compiler
	state.resources = map[string]resourceDocument{rootURI: {document: parsed}}
	state.instances = map[instanceKey]*Document{}
	if err := state.compileDocument(ctx, rootURI, parsed.TargetNamespace, 1); err != nil {
		t.Fatal(err)
	}
	set, err := compiler.Compile(ctx, Source{URI: rootURI, Content: []byte(schema)})
	if err != nil {
		t.Fatal(err)
	}
	elementName := xsd.QName{Namespace: "urn:ordinary", Local: "value"}
	attributeName := xsd.QName{Namespace: "urn:ordinary", Local: "label"}
	element := state.elements[elementName]
	attribute := state.attributes[attributeName]
	document := *state.instances[instanceKey{uri: rootURI, namespace: "urn:ordinary"}]
	namespaces := element.ValueNamespaces
	annotation := element.Annotation
	if annotation == nil || len(annotation.Documentation) != 1 || len(annotation.AppInformation) != 1 ||
		attribute.Annotation == nil || len(attribute.Annotation.Documentation) != 1 || len(attribute.Annotation.AppInformation) != 1 ||
		len(namespaces) == 0 || len(attribute.ValueNamespaces) == 0 || len(document.Dependencies) != 1 {
		t.Fatal("ordinary fixture lacks owned annotations, namespaces, or dependency")
	}
	inputs := []any{element, attribute, document, namespaces, annotation}
	before, err := json.Marshal(inputs)
	if err != nil {
		t.Fatal(err)
	}
	assertInputsUnchanged := func() {
		t.Helper()
		after, err := json.Marshal(inputs)
		if err != nil || string(after) != string(before) {
			t.Fatalf("clone changed original inputs: %v", err)
		}
	}
	assertCopies := func(owner ...*compileState) {
		t.Helper()
		copiedElement := cloneElement(element, owner...)
		copiedAttribute := cloneAttribute(attribute, owner...)
		copiedDocument := cloneDocument(document, owner...)
		copiedNamespaces := cloneNamespaceMap(namespaces, owner...)
		copiedAnnotation := cloneAnnotation(annotation, owner...)
		if !reflect.DeepEqual(copiedElement, element) || !reflect.DeepEqual(copiedAttribute, attribute) ||
			!reflect.DeepEqual(copiedDocument, document) || !reflect.DeepEqual(copiedNamespaces, namespaces) ||
			!reflect.DeepEqual(copiedAnnotation, annotation) {
			t.Fatal("live clone did not preserve equivalent values")
		}
		copiedElement.ValueNamespaces["xs"] = "urn:changed"
		copiedElement.Annotation.Documentation[0].Content = "changed"
		copiedElement.Annotation.AppInformation[0].Content = "changed"
		copiedAttribute.ValueNamespaces["xs"] = "urn:changed"
		copiedAttribute.Annotation.Documentation[0].Content = "changed"
		copiedAttribute.Annotation.AppInformation[0].Content = "changed"
		copiedDocument.Dependencies[0] = "urn:changed"
		copiedNamespaces["xs"] = "urn:changed"
		copiedAnnotation.Documentation[0].Content = "changed"
		copiedAnnotation.AppInformation[0].Content = "changed"
		assertInputsUnchanged()
	}
	assertCopies(&state)
	if err := state.contextError(); err != nil {
		t.Fatalf("live owner error = %v", err)
	}
	cancel()
	if !errors.Is(state.contextError(), context.Canceled) {
		t.Fatal("owner lost real cancellation cause")
	}
	if !reflect.DeepEqual(cloneElement(element, &state), xsd.Element{}) ||
		!reflect.DeepEqual(cloneAttribute(attribute, &state), xsd.Attribute{}) ||
		!reflect.DeepEqual(cloneDocument(document, &state), Document{}) ||
		cloneNamespaceMap(namespaces, &state) != nil || cloneAnnotation(annotation, &state) != nil {
		t.Fatal("canceled owner published a clone")
	}
	assertInputsUnchanged()
	assertCopies()
	publicElement, elementOK := set.Element(elementName)
	publicAttribute, attributeOK := set.Attribute(attributeName)
	publicDocument, documentOK := set.Document(rootURI)
	if !elementOK || !attributeOK || !documentOK || !reflect.DeepEqual(publicElement, element) ||
		!reflect.DeepEqual(publicAttribute, attribute) || !reflect.DeepEqual(publicDocument, document) {
		t.Fatal("owner cancellation changed context-free public accessors")
	}
	publicElement.Annotation.Documentation[0].Content = "changed"
	publicAttribute.ValueNamespaces["xs"] = "urn:changed"
	publicDocument.Dependencies[0] = "urn:changed"
	againElement, _ := set.Element(elementName)
	againAttribute, _ := set.Attribute(attributeName)
	againDocument, _ := set.Document(rootURI)
	if !reflect.DeepEqual(againElement, element) || !reflect.DeepEqual(againAttribute, attribute) || !reflect.DeepEqual(againDocument, document) {
		t.Fatal("public accessor returned an alias of immutable Set storage")
	}
	assertInputsUnchanged()
}
