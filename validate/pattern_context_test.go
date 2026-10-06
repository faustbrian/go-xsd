package validate_test

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-xsd/v2/compile"
	"github.com/faustbrian/go-xsd/v2/validate"
)

func TestPatternOwnersPreserveContextAndOrdinaryAssessment(t *testing.T) {
	c, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	source := compile.Source{URI: "https://example.test/pattern.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:order"><xs:element name="word" default="xsd"><xs:simpleType><xs:restriction base="xs:string"><xs:pattern value="[a-z-[aeiou]]+"/></xs:restriction></xs:simpleType></xs:element></xs:schema>`)}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if set, err := c.Compile(ctx, source); set != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("canceled compiler published a set or lost context cause")
	}
	set, err := c.Compile(context.Background(), source)
	if err != nil {
		t.Fatal(err)
	}
	invalidDefault := source
	invalidDefault.Content = []byte(strings.Replace(string(source.Content), `default="xsd"`, `default="schema"`, 1))
	if set, err := c.Compile(context.Background(), invalidDefault); set != nil || !errors.Is(err, compile.ErrInvalidComponent) {
		t.Fatal("pattern-invalid schema default was not rejected")
	}
	v, err := validate.New(set, validate.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if result, err := v.Validate(ctx, []byte(`<word xmlns="urn:order">xsd</word>`)); !errors.Is(err, context.Canceled) || !reflect.DeepEqual(result, validate.Result{}) {
		t.Fatal("canceled validator returned partial assessment or lost cause")
	}
	for _, test := range []struct {
		source string
		valid  bool
	}{{`<word xmlns="urn:order">xsd</word>`, true}, {`<word xmlns="urn:order">schema</word>`, false}} {
		result, err := v.Validate(context.Background(), []byte(test.source))
		if err != nil || result.Valid != test.valid {
			t.Fatal("ordinary pattern assessment changed")
		}
		if !test.valid && (len(result.Diagnostics) != 1 || result.Diagnostics[0].Code != "cvc-datatype-valid.1.2.1") {
			t.Fatal("ordinary pattern finding classification changed")
		}
	}
}
