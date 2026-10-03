package validate

import (
	"context"
	"errors"
	"reflect"
	"testing"

	xsd "github.com/faustbrian/go-xsd"
)

func TestValidationOwnerStopsCanceledWork(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	state := validationState{ctx: ctx}
	// These finite seams need no schema or input: cancellation must stop the
	// owner before assessment, matching, or value-space work starts.
	if err := state.validateElement(nil, xsd.Element{}, "/root"); !errors.Is(err, context.Canceled) {
		t.Fatalf("element cancellation = %v", err)
	}
	if _, _, err := state.matchGroup(nil, nil, 0, "", "/root"); !errors.Is(err, context.Canceled) {
		t.Fatalf("particle cancellation = %v", err)
	}
	if state.facetsValid(xsd.SimpleType{}, "value") {
		t.Fatal("canceled facets accepted value")
	}
	if state.simpleLexicalValid(xsd.QName{}, "value") {
		t.Fatal("canceled lexical assessment accepted value")
	}
	if err := state.validateIdentityConstraints(nil, nil, "/root"); !errors.Is(err, context.Canceled) {
		t.Fatalf("identity cancellation = %v", err)
	}
	if got := state.identityDescendants(nil); len(got) != 0 {
		t.Fatalf("canceled identity traversal = %#v", got)
	}
	if err := state.validateIDReferences(); !errors.Is(err, context.Canceled) {
		t.Fatalf("ID-reference cancellation = %v", err)
	}
	validator := &Validator{}
	result, err := validator.validateRoot(ctx, nil)
	if !errors.Is(err, context.Canceled) || !reflect.DeepEqual(result, Result{}) {
		t.Fatalf("owner cancellation = %#v, %v", result, err)
	}
}
