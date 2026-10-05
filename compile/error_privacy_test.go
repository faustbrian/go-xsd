package compile_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-xsd/compile"
	"github.com/faustbrian/go-xsd/resolve"
)

type privateCompileCause struct{ calls int }

func (e *privateCompileCause) Error() string { e.calls++; return "synthetic-private-cause" }
func (e *privateCompileCause) Format(s fmt.State, _ rune) {
	e.calls++
	_, _ = s.Write([]byte("synthetic-private-cause"))
}
func (e *privateCompileCause) MarshalJSON() ([]byte, error) {
	e.calls++
	return []byte(`"synthetic-private-cause"`), nil
}
func (e *privateCompileCause) LogValue() slog.Value {
	e.calls++
	return slog.StringValue("synthetic-private-cause")
}

type privateCompileResolver struct {
	resource resolve.Resource
	err      error
}

func (r privateCompileResolver) Resolve(context.Context, resolve.Request) (resolve.Resource, error) {
	return r.resource, r.err
}

func TestCompilerDefaultErrorPrivacy(t *testing.T) {
	cause := &privateCompileCause{}
	c, err := compile.New(compile.Options{Resolver: privateCompileResolver{err: cause}})
	if err != nil {
		t.Fatal(err)
	}
	set, err := c.Compile(context.Background(), compile.Source{URI: "https://example.test/synthetic-private-root.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:ordinary"><xs:include schemaLocation="synthetic-private-child.xsd"/></xs:schema>`)})
	if set != nil || err == nil {
		t.Fatal("delegated include failure published a set")
	}
	assertCompilerSafeError(t, err)
	var inspected *privateCompileCause
	if !errors.Is(err, cause) || !errors.As(err, &inspected) || inspected != cause {
		t.Fatal("trusted cause inspection changed")
	}
	if cause.calls != 0 {
		t.Fatal("default boundary evaluated delegated cause callbacks")
	}
}

func TestCompilerPrivacyPreservesClassification(t *testing.T) {
	for _, test := range []struct {
		name, body string
		want       error
	}{
		{"duplicate", `<xs:element name="synthetic-private-name" type="xs:string"/><xs:element name="synthetic-private-name" type="xs:string"/>`, compile.ErrDuplicateComponent},
		{"unresolved", `<xs:element name="ordinary" type="t:synthetic-private-missing"/>`, compile.ErrUnresolvedComponent},
		{"facet", `<xs:simpleType name="ordinary"><xs:restriction base="xs:string"><xs:pattern value="["/></xs:restriction></xs:simpleType>`, compile.ErrInvalidComponent},
	} {
		t.Run(test.name, func(t *testing.T) {
			c, e := compile.New(compile.Options{})
			if e != nil {
				t.Fatal(e)
			}
			set, e := c.Compile(context.Background(), compile.Source{URI: "https://example.test/ordinary.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" xmlns:t="urn:ordinary" targetNamespace="urn:ordinary">` + test.body + `</xs:schema>`)})
			if set != nil || !errors.Is(e, test.want) {
				t.Fatal("schema refusal lost sentinel or published a set")
			}
			assertCompilerSafeError(t, e)
		})
	}
	root := compile.Source{URI: "https://example.test/root.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:ordinary"><xs:include schemaLocation="child.xsd"/></xs:schema>`)}
	for _, test := range []struct {
		resource resolve.Resource
		want     error
	}{
		{resolve.Resource{URI: "https://example.test/synthetic-private-wrong.xsd"}, compile.ErrResourceIdentity},
		{resolve.Resource{URI: "https://example.test/child.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:synthetic-private-wrong"/>`)}, compile.ErrNamespace},
	} {
		c, e := compile.New(compile.Options{Resolver: privateCompileResolver{resource: test.resource}})
		if e != nil {
			t.Fatal(e)
		}
		set, e := c.Compile(context.Background(), root)
		if set != nil || !errors.Is(e, test.want) {
			t.Fatal("resolved graph refusal changed classification")
		}
		assertCompilerSafeError(t, e)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	c, e := compile.New(compile.Options{})
	if e != nil {
		t.Fatal(e)
	}
	if set, e := c.Compile(ctx, root); set != nil || e != context.Canceled {
		t.Fatal("standard canceled compile cause changed")
	}
}

func TestPublicErrorOwnersPreserveExpiredDeadline(t *testing.T) {
	ctx, cancel := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer cancel()
	c, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	if set, err := c.Compile(ctx, compile.Source{URI: "https://example.test/ordinary.xsd", Content: []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`)}); set != nil || err != context.DeadlineExceeded {
		t.Fatal("standard expired compiler deadline changed")
	}
	memory, err := resolve.NewMemory(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, resolver := range []resolve.Resolver{resolve.Deny(), memory, resolve.Chain(memory)} {
		resource, err := resolver.Resolve(ctx, resolve.Request{URI: "urn:ordinary"})
		if resource.URI != "" || resource.Content != nil || err != context.DeadlineExceeded {
			t.Fatal("standard expired resolver deadline changed")
		}
	}
}

func assertCompilerSafeError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected failure")
	}
	if err.Error() != "xsd compile: failed" {
		t.Error("compiler Error did not return fixed category")
	}
	for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		want := "xsd compile: failed"
		if format == "%q" {
			want = `"xsd compile: failed"`
		}
		if fmt.Sprintf(format, err) != want {
			t.Error("compiler default formatting did not return fixed category")
		}
	}
	encoded, e := json.Marshal(err)
	if e != nil || string(encoded) != `"xsd compile: failed"` {
		t.Error("compiler JSON did not return fixed category")
	}
	for _, text := range []bool{false, true} {
		var output bytes.Buffer
		var handler slog.Handler = slog.NewJSONHandler(&output, nil)
		if text {
			handler = slog.NewTextHandler(&output, nil)
		}
		slog.New(handler).Info("operation", slog.Any("error", err))
		if !strings.Contains(output.String(), "xsd compile: failed") || strings.Contains(output.String(), "synthetic-private-") {
			t.Error("compiler logging exposed details or lost category")
		}
	}
}
