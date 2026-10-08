package resolve_test

import (
	"context"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/faustbrian/go-xsd/v2/resolve"
)

func TestFileDefaultURIAdmission(t *testing.T) {
	resolver, err := resolve.NewFile(resolve.FileOptions{Root: t.TempDir()})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resolver.Close() })
	// Admission precedes URI syntax processing, even for an invalid request.
	resource, err := resolver.Resolve(t.Context(), resolve.Request{URI: strings.Repeat("%", (64<<10)+1)})
	if !errors.Is(err, resolve.ErrLimitExceeded) || resource.URI != "" || resource.Content != nil {
		t.Fatal("request above default URI allowance must fail with a typed limit and no resource")
	}
}

func TestFileURIAdmissionBoundaries(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "schema.xsd")
	if err := os.WriteFile(path, []byte("four"), 0o600); err != nil {
		t.Fatal(err)
	}
	identity := (&url.URL{Scheme: "file", Path: path}).String()
	if resolver, err := resolve.NewFile(resolve.FileOptions{Root: root, MaxURIBytes: -1}); resolver != nil || err == nil {
		t.Fatal("negative URI allowance accepted")
	}
	resolver, err := resolve.NewFile(resolve.FileOptions{Root: root, MaxBytes: 4, MaxURIBytes: int64(len(identity))})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = resolver.Close() })
	for _, request := range []string{identity + "a", strings.Repeat("%", len(identity)+1)} {
		resource, err := resolver.Resolve(t.Context(), resolve.Request{URI: request})
		if !errors.Is(err, resolve.ErrLimitExceeded) || errors.Is(err, resolve.ErrAccessDenied) || resource.URI != "" || resource.Content != nil {
			t.Fatal("URI admission did not precede syntax/filesystem work or published a resource")
		}
		resource, err = resolver.Resolve(t.Context(), resolve.Request{URI: identity})
		if err != nil || resource.URI != identity || string(resource.Content) != "four" {
			t.Fatal("exact URI/content allowance failed after refusal")
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	resource, err := resolver.Resolve(ctx, resolve.Request{URI: identity + "a"})
	if !errors.Is(err, context.Canceled) || resource.URI != "" || resource.Content != nil {
		t.Fatal("entry cancellation did not precede URI refusal")
	}
}
