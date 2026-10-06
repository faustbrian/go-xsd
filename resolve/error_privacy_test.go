package resolve_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/faustbrian/go-xsd/resolve"
)

type privateResolverCause struct{ calls int }

func (e *privateResolverCause) Error() string { e.calls++; return "synthetic-private-cause" }
func (e *privateResolverCause) Format(s fmt.State, _ rune) {
	e.calls++
	_, _ = s.Write([]byte("synthetic-private-cause"))
}

func TestResolverPrivacyPreservesFallbackAndClassification(t *testing.T) {
	request := resolve.Request{URI: "urn:ordinary"}
	for _, cause := range []error{resolve.ErrNotFound, resolve.ErrAccessDenied, &privateResolverCause{}} {
		fallbackCalls := 0
		resolver := resolve.Chain(privateResolver{err: cause}, privateResolver{resource: resolve.Resource{URI: request.URI, Content: []byte("ordinary")}, calls: &fallbackCalls})
		resource, err := resolver.Resolve(context.Background(), request)
		if cause == resolve.ErrNotFound {
			if err != nil || string(resource.Content) != "ordinary" || fallbackCalls != 1 {
				t.Fatal("not-found fallback changed")
			}
		} else {
			if !errors.Is(err, cause) || fallbackCalls != 0 || !reflect.DeepEqual(resource, resolve.Resource{}) {
				t.Fatal("operational/access error did not stop fallback")
			}
			assertResolverSafeError(t, err)
		}
	}
	for _, kind := range []resolve.Kind{resolve.KindInclude, resolve.KindImport} {
		catalog, err := resolve.NewCatalog(nil, resolve.Deny())
		if err != nil {
			t.Fatal(err)
		}
		resource, err := catalog.Resolve(context.Background(), resolve.Request{Namespace: "urn:synthetic-private-namespace", Kind: kind})
		if !errors.Is(err, resolve.ErrNotFound) || !reflect.DeepEqual(resource, resolve.Resource{}) {
			t.Fatal("catalog missing mapping classification changed")
		}
		assertResolverSafeError(t, err)
	}
	catalog, err := resolve.NewCatalog(map[string]string{"urn:ordinary": request.URI}, privateResolver{resource: resolve.Resource{URI: "urn:synthetic-private-wrong"}})
	if err != nil {
		t.Fatal(err)
	}
	resource, err := catalog.Resolve(context.Background(), resolve.Request{Namespace: "urn:ordinary", Kind: resolve.KindImport})
	if err == nil || !reflect.DeepEqual(resource, resolve.Resource{}) {
		t.Fatal("catalog wrong identity published a resource")
	}
	assertResolverSafeError(t, err)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	for _, resolver := range []resolve.Resolver{resolve.Deny(), resolve.Chain(privateResolver{resource: resolve.Resource{URI: request.URI}}), catalog} {
		resource, err := resolver.Resolve(ctx, request)
		if err != context.Canceled || !reflect.DeepEqual(resource, resolve.Resource{}) {
			t.Fatal("standard canceled resolver cause changed")
		}
	}
}

func TestFileDefaultPrivacyPreservesTrustedOSCause(t *testing.T) {
	root := t.TempDir()
	if resolver, err := resolve.NewFile(resolve.FileOptions{Root: filepath.Join(root, "synthetic-private-missing-root")}); resolver != nil || err == nil {
		t.Fatal("missing root accepted")
	} else {
		assertResolverSafeError(t, err)
		var inspected *os.PathError
		if !errors.As(err, &inspected) || !errors.Is(err, os.ErrNotExist) {
			t.Fatal("file constructor lost OS cause")
		}
	}
	resolver, err := resolve.NewFile(resolve.FileOptions{Root: root})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resolver.Close() })
	request := resolve.Request{URI: (&url.URL{Scheme: "file", Path: filepath.Join(root, "synthetic-private-missing.xsd")}).String()}
	resource, err := resolver.Resolve(context.Background(), request)
	var inspected *os.PathError
	if !reflect.DeepEqual(resource, resolve.Resource{}) || !errors.Is(err, resolve.ErrNotFound) || !errors.Is(err, os.ErrNotExist) || !errors.As(err, &inspected) {
		t.Fatal("file missing-resource classification or OS cause changed")
	}
	assertResolverSafeError(t, err)
	if err := resolver.Close(); err != nil {
		t.Fatal("ordinary file close failed")
	}
	resource, err = resolver.Resolve(context.Background(), request)
	if !reflect.DeepEqual(resource, resolve.Resource{}) || !errors.Is(err, resolve.ErrAccessDenied) || !errors.Is(err, os.ErrClosed) {
		t.Fatal("closed-root operational classification or OS cause changed")
	}
	assertResolverSafeError(t, err)
	var absent *resolve.File
	if err := absent.Close(); err != nil {
		t.Fatal("nil file close changed")
	}
}
func (e *privateResolverCause) MarshalJSON() ([]byte, error) {
	e.calls++
	return []byte(`"synthetic-private-cause"`), nil
}
func (e *privateResolverCause) LogValue() slog.Value {
	e.calls++
	return slog.StringValue("synthetic-private-cause")
}

type privateResolver struct {
	resource resolve.Resource
	err      error
	calls    *int
}

func (r privateResolver) Resolve(context.Context, resolve.Request) (resolve.Resource, error) {
	if r.calls != nil {
		*r.calls++
	}
	return r.resource, r.err
}

func TestResolverDefaultErrorPrivacy(t *testing.T) {
	request := resolve.Request{URI: "urn:synthetic-private-resource"}
	memory, err := resolve.NewMemory(nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, resolver := range []resolve.Resolver{resolve.Deny(), memory, resolve.Chain(memory)} {
		resource, err := resolver.Resolve(context.Background(), request)
		if !reflect.DeepEqual(resource, resolve.Resource{}) {
			t.Fatal("resolver returned partial failure resource")
		}
		assertResolverSafeError(t, err)
	}
	if _, err := resolve.NewMemory(map[string][]byte{"synthetic-private-relative": nil}); err == nil {
		t.Fatal("invalid identity accepted")
	} else {
		assertResolverSafeError(t, err)
	}
	if _, err := resolve.NewCatalog(map[string]string{"urn:ordinary": "synthetic-private-relative"}, memory); err == nil {
		t.Fatal("invalid catalog identity accepted")
	} else {
		assertResolverSafeError(t, err)
	}
	cause := &privateResolverCause{}
	catalog, err := resolve.NewCatalog(map[string]string{"urn:ordinary": request.URI}, privateResolver{resource: resolve.Resource{URI: request.URI, Content: []byte("partial")}, err: cause})
	if err != nil {
		t.Fatal(err)
	}
	for _, resolver := range []resolve.Resolver{resolve.Chain(privateResolver{err: cause}), catalog} {
		resource, err := resolver.Resolve(context.Background(), request)
		if !reflect.DeepEqual(resource, resolve.Resource{}) {
			t.Error("delegated failure leaked partial resource")
		}
		assertResolverSafeError(t, err)
		var inspected *privateResolverCause
		if !errors.Is(err, cause) || !errors.As(err, &inspected) || inspected != cause {
			t.Error("delegated cause inspection changed")
		}
	}
	if cause.calls != 0 {
		t.Fatal("default resolver boundary evaluated cause callbacks")
	}
}

func assertResolverSafeError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected failure")
	}
	if err.Error() != "xsd resolve: failed" {
		t.Error("resolver Error did not return fixed category")
	}
	for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		want := "xsd resolve: failed"
		if format == "%q" {
			want = `"xsd resolve: failed"`
		}
		if fmt.Sprintf(format, err) != want {
			t.Error("resolver default formatting did not return fixed category")
		}
	}
	encoded, e := json.Marshal(err)
	if e != nil || string(encoded) != `"xsd resolve: failed"` {
		t.Error("resolver JSON did not return fixed category")
	}
	for _, text := range []bool{false, true} {
		var output bytes.Buffer
		var handler slog.Handler = slog.NewJSONHandler(&output, nil)
		if text {
			handler = slog.NewTextHandler(&output, nil)
		}
		slog.New(handler).Info("operation", slog.Any("error", err))
		if !strings.Contains(output.String(), "xsd resolve: failed") || strings.Contains(output.String(), "synthetic-private-") {
			t.Error("resolver logging exposed details or lost category")
		}
	}
}
