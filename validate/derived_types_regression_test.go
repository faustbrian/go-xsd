package validate_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/faustbrian/go-xsd/v2/compile"
	"github.com/faustbrian/go-xsd/v2/validate"
)

func TestValidateNamedTypeSemantics(t *testing.T) {
	for _, test := range []struct {
		name, types, root, valid, invalid string
	}{
		{
			name:  "union identity",
			types: `<xs:simpleType name="Union"><xs:union memberTypes="xs:decimal xs:string"/></xs:simpleType>`,
			root: `<xs:element name="root"><xs:complexType><xs:sequence>
 <xs:element name="value" type="t:Union" maxOccurs="unbounded"/>
</xs:sequence></xs:complexType><xs:unique name="values">
 <xs:selector xpath="t:value"/><xs:field xpath="."/>
</xs:unique></xs:element>`,
			valid: `<value>1</value><value>2</value>`, invalid: `<value>01</value><value>1</value>`,
		},
		{
			name:  "derived ID",
			types: `<xs:simpleType name="MyID"><xs:restriction base="xs:ID"/></xs:simpleType>`,
			root: `<xs:element name="root"><xs:complexType><xs:sequence>
 <xs:element name="value" type="t:MyID" maxOccurs="unbounded"/>
</xs:sequence></xs:complexType></xs:element>`,
			valid: `<value>a</value><value>b</value>`, invalid: `<value>a</value><value>a</value>`,
		},
		{
			name: "derived binary length",
			types: `<xs:simpleType name="Hex"><xs:restriction base="xs:hexBinary"/></xs:simpleType>
<xs:simpleType name="OneByte"><xs:restriction base="t:Hex"><xs:length value="1"/></xs:restriction></xs:simpleType>`,
			root:  `<xs:element name="root" type="t:OneByte"/>`,
			valid: `0A`, invalid: `0A0B`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			compiler, err := compile.New(compile.Options{})
			if err != nil {
				t.Fatal(err)
			}
			set, err := compiler.Compile(context.Background(), compile.Source{
				URI: "https://example.test/derived-types.xsd",
				Content: []byte(fmt.Sprintf(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 xmlns:t="urn:derived-types" targetNamespace="urn:derived-types" elementFormDefault="qualified">
%s%s</xs:schema>`, test.types, test.root)),
			})
			if err != nil {
				t.Fatal(err)
			}
			validator, err := validate.New(set, validate.Options{})
			if err != nil {
				t.Fatal(err)
			}
			for _, instance := range []struct {
				content   string
				wantValid bool
			}{{test.valid, true}, {test.invalid, false}} {
				result, err := validator.Validate(context.Background(), []byte(`<root xmlns="urn:derived-types">`+instance.content+`</root>`))
				if err != nil || result.Valid != instance.wantValid || (!instance.wantValid && len(result.Diagnostics) == 0) {
					t.Fatalf("Validate(%q) = %#v, %v; want valid=%t", instance.content, result, err, instance.wantValid)
				}
			}
		})
	}
}
