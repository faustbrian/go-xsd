package validate_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/compile"
	"github.com/faustbrian/go-xsd/v2/validate"
)

func TestValidateTreeOwnedPayloadAdmission(t *testing.T) {
	node := validate.Node{Name: xsd.QName{Namespace: "urn:order", Local: "amount"}, Text: "1"}
	// Each retained string occurrence contributes its byte length.
	payload := int64(len(node.Name.Namespace) + len(node.Name.Local) + len(node.Text))
	for _, allowance := range []int64{payload - 1, payload} {
		validator, err := validate.New(validatorSet(t), validate.Options{Limits: validate.Limits{MaxBytes: allowance}})
		if err != nil {
			t.Fatal(err)
		}
		result, err := validator.ValidateTree(context.Background(), node)
		if allowance < payload {
			if !errors.Is(err, validate.ErrLimitExceeded) || !reflect.DeepEqual(result, validate.Result{}) {
				t.Fatalf("owned payload refusal = %#v, %v", result, err)
			}
		} else if err != nil || !result.Valid {
			t.Fatalf("exact owned payload = %#v, %v", result, err)
		}
	}
}

func TestValidateTreePayloadOccurrences(t *testing.T) {
	c, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	set, err := c.Compile(context.Background(), compile.Source{URI: "https://example.test/payload.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:order" attributeFormDefault="qualified"><xs:element name="amount"><xs:complexType><xs:simpleContent><xs:extension base="xs:decimal"><xs:attribute name="flag" type="xs:string"/></xs:extension></xs:simpleContent></xs:complexType></xs:element></xs:schema>`)})
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name  string
		node  validate.Node
		bytes int64
	}{
		{"attribute", validate.Node{Name: xsd.QName{Namespace: "urn:order", Local: "amount"}, Text: "1", Attributes: map[xsd.QName]string{{Namespace: "urn:order", Local: "flag"}: "yes"}}, 32},
		{"namespace", validate.Node{Name: xsd.QName{Namespace: "urn:order", Local: "amount"}, Text: "1", Namespaces: map[string]string{"p": "urn:order"}}, 26},
		{"location", validate.Node{Name: xsd.QName{Namespace: "urn:order", Local: "amount"}, Text: "1", Location: xsd.Location{SystemID: "fixture"}}, 23},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, limit := range []int64{test.bytes - 1, test.bytes} {
				v, err := validate.New(set, validate.Options{Limits: validate.Limits{MaxBytes: limit}})
				if err != nil {
					t.Fatal(err)
				}
				result, err := v.ValidateTree(context.Background(), test.node)
				if limit < test.bytes {
					if !errors.Is(err, validate.ErrLimitExceeded) || !reflect.DeepEqual(result, validate.Result{}) {
						t.Fatalf("refusal = %#v, %v", result, err)
					}
				} else if err != nil || !result.Valid {
					t.Fatalf("exact allowance = %#v, %v", result, err)
				}
			}
		})
	}
}

func TestValidateNamespaceAdmissionParity(t *testing.T) {
	c, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	set, err := c.Compile(context.Background(), compile.Source{URI: "https://example.test/admission.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:order" elementFormDefault="qualified"><xs:element name="order"><xs:complexType><xs:sequence><xs:element name="id" type="xs:string"/></xs:sequence><xs:attribute name="status" type="xs:string"/></xs:complexType></xs:element></xs:schema>`)})
	if err != nil {
		t.Fatal(err)
	}
	source := []byte(`<order xmlns="urn:order" status="new"><id>A</id></order>`)
	tree := validate.Node{Name: xsd.QName{Namespace: "urn:order", Local: "order"}, Namespaces: map[string]string{"": "urn:order"}, Attributes: map[xsd.QName]string{{Local: "status"}: "new"}, Children: []validate.Node{{Name: xsd.QName{Namespace: "urn:order", Local: "id"}, Namespaces: map[string]string{"": "urn:order"}, Text: "A"}}}
	for _, limit := range []int{1, 2, 0} {
		v, err := validate.New(set, validate.Options{Limits: validate.Limits{MaxNamespaceEntries: limit, MaxAttributes: 1}})
		if err != nil {
			t.Fatal(err)
		}
		for _, call := range []struct {
			name string
			run  func() (validate.Result, error)
		}{
			{"bytes", func() (validate.Result, error) { return v.Validate(context.Background(), source) }},
			{"reader", func() (validate.Result, error) {
				return v.ValidateReader(context.Background(), bytes.NewReader(source))
			}},
			{"tree", func() (validate.Result, error) { return v.ValidateTree(context.Background(), tree) }},
		} {
			t.Run(call.name+fmt.Sprint(limit), func(t *testing.T) {
				result, err := call.run()
				if limit == 1 {
					if !errors.Is(err, validate.ErrLimitExceeded) || !reflect.DeepEqual(result, validate.Result{}) {
						t.Fatalf("refusal = %#v, %v", result, err)
					}
				} else if err != nil || !result.Valid {
					t.Fatalf("accepted = %#v, %v", result, err)
				}
			})
		}
	}
	if _, err := validate.New(set, validate.Options{Limits: validate.Limits{MaxNamespaceEntries: -1}}); err == nil {
		t.Fatal("negative namespace policy accepted")
	}
	// Namespace maps are caller-owned complete scopes, not inherited tree data.
	if tree.Namespaces[""] != "urn:order" || tree.Children[0].Namespaces[""] != "urn:order" || tree.Attributes[xsd.QName{Local: "status"}] != "new" {
		t.Fatal("caller tree changed")
	}
}

func TestValidateNamespaceRebindingWork(t *testing.T) {
	source := []byte(`<amount xmlns="urn:order" xmlns:p="urn:first" xmlns:q="urn:second">1</amount>`)
	for _, limit := range []int{2, 3} {
		v, err := validate.New(validatorSet(t), validate.Options{Limits: validate.Limits{MaxNamespaceEntries: limit}})
		if err != nil {
			t.Fatal(err)
		}
		result, err := v.Validate(context.Background(), source)
		if limit == 2 {
			if !errors.Is(err, validate.ErrLimitExceeded) || !reflect.DeepEqual(result, validate.Result{}) {
				t.Fatalf("declaration refusal = %#v, %v", result, err)
			}
		} else if err != nil || !result.Valid {
			t.Fatalf("declarations = %#v, %v", result, err)
		}
	}
	// A child's replacement declaration is charged in addition to its inherited scope.
	source = []byte(`<complex xmlns="urn:order"><child xmlns="urn:other"/></complex>`)
	v, err := validate.New(validatorSet(t), validate.Options{Limits: validate.Limits{MaxNamespaceEntries: 2}})
	if err != nil {
		t.Fatal(err)
	}
	if result, err := v.Validate(context.Background(), source); !errors.Is(err, validate.ErrLimitExceeded) || !reflect.DeepEqual(result, validate.Result{}) {
		t.Fatalf("rebinding refusal = %#v, %v", result, err)
	}
}

func TestValidateReaderOwnedPayloadAndCancellation(t *testing.T) {
	source := []byte(`<amount xmlns="urn:order">1</amount>`)
	v, err := validate.New(validatorSet(t), validate.Options{SystemID: "https://example.test/a-long-instance-identity", Limits: validate.Limits{MaxBytes: int64(len(source))}})
	if err != nil {
		t.Fatal(err)
	}
	for _, read := range []bool{false, true} {
		var result validate.Result
		if read {
			result, err = v.ValidateReader(context.Background(), bytes.NewReader(source))
		} else {
			result, err = v.Validate(context.Background(), source)
		}
		if !errors.Is(err, validate.ErrLimitExceeded) || !reflect.DeepEqual(result, validate.Result{}) {
			t.Fatalf("owned reader payload = %#v, %v", result, err)
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if result, err := v.Validate(ctx, source); !errors.Is(err, context.Canceled) || !reflect.DeepEqual(result, validate.Result{}) {
		t.Fatalf("cancellation precedence = %#v, %v", result, err)
	}
	if result, err := v.ValidateReader(ctx, nil); !errors.Is(err, context.Canceled) || !reflect.DeepEqual(result, validate.Result{}) {
		t.Fatalf("canceled reader admission = %#v, %v", result, err)
	}
}

func TestValidateTreeExplicitQNameScope(t *testing.T) {
	c, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	set, err := c.Compile(context.Background(), compile.Source{URI: "https://example.test/qname-admission.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:p="urn:value" targetNamespace="urn:order" elementFormDefault="qualified"><xs:element name="order"><xs:complexType><xs:sequence><xs:element name="id" type="xs:QName" fixed="p:item"/></xs:sequence></xs:complexType></xs:element></xs:schema>`)})
	if err != nil {
		t.Fatal(err)
	}
	v, err := validate.New(set, validate.Options{Limits: validate.Limits{MaxNamespaceEntries: 4}})
	if err != nil {
		t.Fatal(err)
	}
	source := []byte(`<order xmlns="urn:order" xmlns:p="urn:value"><id>p:item</id></order>`)
	if result, err := v.ValidateReader(context.Background(), bytes.NewReader(source)); err != nil || !result.Valid {
		t.Fatalf("inherited reader QName = %#v, %v", result, err)
	}
	tree := validate.Node{Name: xsd.QName{Namespace: "urn:order", Local: "order"}, Namespaces: map[string]string{"": "urn:order", "p": "urn:value"}, Children: []validate.Node{{Name: xsd.QName{Namespace: "urn:order", Local: "id"}, Text: "p:item"}}}
	if result, err := v.ValidateTree(context.Background(), tree); err != nil || result.Valid {
		t.Fatalf("tree unexpectedly inherited QName scope = %#v, %v", result, err)
	}
	tree.Children[0].Namespaces = map[string]string{"": "urn:order", "p": "urn:value"}
	if result, err := v.ValidateTree(context.Background(), tree); err != nil || !result.Valid {
		t.Fatalf("explicit tree QName = %#v, %v", result, err)
	}
	if tree.Children[0].Text != "p:item" || tree.Children[0].Namespaces["p"] != "urn:value" {
		t.Fatal("caller QName scope changed")
	}
}

func TestValidateReaderTextPrefixWork(t *testing.T) {
	identity := "https://example.test/ordinary-instance-document"
	flat := []byte(`<amount xmlns="urn:order">1.5</amount>`)
	segmented := []byte(`<amount xmlns="urn:order">1<![CDATA[.]]>5</amount>`)
	// Names, one default declaration and the location are each owned once.
	base := int64(len("urn:order") + len("amount") + len("urn:order") + len(identity))
	for _, test := range []struct {
		name      string
		source    []byte
		allowance int64
		refused   bool
	}{
		{"flat", flat, base + 3, false},
		{"recopied-prefix", segmented, base + 3, true},
		{"exact-prefix-work", segmented, base + 1 + 2 + 3, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			v, err := validate.New(validatorSet(t), validate.Options{SystemID: identity, Limits: validate.Limits{MaxBytes: test.allowance, MaxTextBytes: 3}})
			if err != nil {
				t.Fatal(err)
			}
			result, err := v.ValidateReader(context.Background(), bytes.NewReader(test.source))
			if test.refused {
				if !errors.Is(err, validate.ErrLimitExceeded) || !reflect.DeepEqual(result, validate.Result{}) {
					t.Fatalf("prefix work refusal = %#v, %v", result, err)
				}
			} else if err != nil || !result.Valid {
				t.Fatalf("exact work admission = %#v, %v", result, err)
			}
		})
	}
}

func TestValidateTreeChildAdmissionBeforeTraversal(t *testing.T) {
	validator, err := validate.New(validatorSet(t), validate.Options{Limits: validate.Limits{MaxNodes: 2}})
	if err != nil {
		t.Fatal(err)
	}
	node := validate.Node{Name: xsd.QName{Namespace: "urn:order", Local: "amount"}, Children: []validate.Node{{}, {Name: xsd.QName{Local: "child"}}}}
	result, err := validator.ValidateTree(context.Background(), node)
	if !errors.Is(err, validate.ErrLimitExceeded) || !reflect.DeepEqual(result, validate.Result{}) {
		t.Fatalf("known child count refusal = %#v, %v", result, err)
	}
	validator, err = validate.New(validatorSet(t), validate.Options{Limits: validate.Limits{MaxNodes: 3}})
	if err != nil {
		t.Fatal(err)
	}
	node.Children[0].Name = xsd.QName{Local: "child"}
	if _, err := validator.ValidateTree(context.Background(), node); err != nil {
		t.Fatalf("exact node allowance double charged visits: %v", err)
	}
}
