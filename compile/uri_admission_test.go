package compile_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-xsd/v2/compile"
	"github.com/faustbrian/go-xsd/v2/resolve"
)

func TestCompilerDefaultURIAdmission(t *testing.T) {
	compiler, err := compile.New(compile.Options{})
	if err != nil {
		t.Fatal(err)
	}
	const maximum = 64 << 10
	content := []byte(`<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`)
	exact := "urn:" + strings.Repeat("a", maximum-len("urn:"))
	if set, err := compiler.Compile(t.Context(), compile.Source{URI: exact, Content: content}); err != nil || set == nil {
		t.Fatal("exact default URI allowance rejected")
	}
	if set, err := compiler.Compile(t.Context(), compile.Source{URI: exact + "a", Content: content}); set != nil || !errors.Is(err, compile.ErrLimitExceeded) {
		t.Fatal("URI above default allowance must fail with a typed limit and no Set")
	}
}

type uriAdmissionResolver func(context.Context, resolve.Request) (resolve.Resource, error)

func (f uriAdmissionResolver) Resolve(ctx context.Context, request resolve.Request) (resolve.Resource, error) {
	return f(ctx, request)
}

func TestCompilerURIAdmissionBoundaries(t *testing.T) {
	const maximum = 16
	const minimal = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`
	exact := "urn:" + strings.Repeat("a", maximum-len("urn:"))
	if compiler, err := compile.New(compile.Options{Limits: compile.Limits{MaxURIBytes: -1}}); compiler != nil || err == nil {
		t.Fatal("negative URI allowance accepted")
	}
	for _, test := range []struct {
		name, root, body, returned string
		calls                      int
		cancel                     bool
	}{
		{"root one over", exact + "a", minimal, "", 0, false},
		{"malformed root one over", strings.Repeat("%", maximum+1), minimal, "", 0, false},
		{"reference one over", "urn:root", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:include schemaLocation="` + exact + `a"/></xs:schema>`, "", 0, false},
		{"returned explicit one over", "urn:root", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:include schemaLocation="urn:child"/></xs:schema>`, exact + "a", 1, false},
		{"returned namespace one over", "urn:root", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:import namespace="urn:import"/></xs:schema>`, exact + "a", 1, false},
		{"canceled returned one over", "urn:root", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:include schemaLocation="urn:child"/></xs:schema>`, exact + "a", 1, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			calls := 0
			compiler, err := compile.New(compile.Options{
				Limits: compile.Limits{MaxURIBytes: maximum},
				Resolver: uriAdmissionResolver(func(context.Context, resolve.Request) (resolve.Resource, error) {
					calls++
					if test.cancel {
						cancel()
					}
					return resolve.Resource{URI: test.returned, Content: []byte(minimal)}, nil
				}),
			})
			if err != nil {
				t.Fatal(err)
			}
			set, err := compiler.Compile(ctx, compile.Source{URI: test.root, Content: []byte(test.body)})
			want := compile.ErrLimitExceeded
			if test.cancel {
				want = context.Canceled
			}
			if set != nil || !errors.Is(err, want) || calls != test.calls || errors.Is(err, compile.ErrResourceIdentity) {
				t.Fatal("URI admission lost priority, dispatched prematurely or published a partial Set")
			}
			if set, err := compiler.Compile(t.Context(), compile.Source{URI: exact, Content: []byte(minimal)}); set == nil || err != nil {
				t.Fatal("fresh invocation at exact allowance failed after refusal")
			}
		})
	}
	compiler, err := compile.New(compile.Options{Limits: compile.Limits{MaxURIBytes: maximum}})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if set, err := compiler.Compile(ctx, compile.Source{URI: exact + "a", Content: []byte(minimal)}); set != nil || !errors.Is(err, context.Canceled) {
		t.Fatal("entry cancellation did not precede URI refusal")
	}
	expired, stop := context.WithDeadline(t.Context(), time.Unix(1, 0))
	defer stop()
	if set, err := compiler.Compile(expired, compile.Source{URI: exact + "a", Content: []byte(minimal)}); set != nil || !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("expired deadline did not precede URI refusal")
	}
}

func TestCompilerResolvedURIAtExactAllowance(t *testing.T) {
	const maximum = 16
	exact := "urn:" + strings.Repeat("a", maximum-len("urn:"))
	for _, namespaceOnly := range []bool{false, true} {
		body := `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:include schemaLocation="` + exact + `"/></xs:schema>`
		child := `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"/>`
		if namespaceOnly {
			body = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:import namespace="urn:import"/></xs:schema>`
			child = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema" targetNamespace="urn:import"/>`
		}
		calls := 0
		compiler, err := compile.New(compile.Options{
			Limits: compile.Limits{MaxURIBytes: maximum},
			Resolver: uriAdmissionResolver(func(context.Context, resolve.Request) (resolve.Resource, error) {
				calls++
				return resolve.Resource{URI: exact, Content: []byte(child)}, nil
			}),
		})
		if err != nil {
			t.Fatal(err)
		}
		if set, err := compiler.Compile(t.Context(), compile.Source{URI: "urn:root", Content: []byte(body)}); err != nil || set == nil || calls != 1 {
			t.Fatal("exact reference/resolved URI allowance rejected")
		}
	}
}
