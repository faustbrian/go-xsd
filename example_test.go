package xsd_test

import (
	"context"
	"fmt"

	"github.com/faustbrian/go-xsd/compile"
	"github.com/faustbrian/go-xsd/validate"
)

func Example() {
	ctx := context.Background()
	schema := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
  targetNamespace="urn:example" elementFormDefault="qualified">
  <xs:element name="order" type="xs:string"/>
</xs:schema>`)

	compiler, err := compile.New(compile.Options{})
	if err != nil {
		panic(err)
	}
	set, err := compiler.Compile(ctx, compile.Source{
		URI:     "https://example.test/order.xsd",
		Content: schema,
	})
	if err != nil {
		panic(err)
	}
	validator, err := validate.New(set, validate.Options{})
	if err != nil {
		panic(err)
	}
	result, err := validator.Validate(ctx, []byte(`<order xmlns="urn:example">ready</order>`))
	if err != nil {
		panic(err)
	}

	fmt.Println(result.Valid)
	// Output: true
}
