package resolve_test

import (
	"context"
	"reflect"
	"testing"

	"github.com/faustbrian/go-xsd/v2/resolve"
)

func TestChainCancellationStopsFallbackAndAllowsReuse(t *testing.T) {
	request := resolve.Request{URI: "urn:ordinary"}
	fallbackCalls := 0
	cancelChild := false
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	chain := resolve.Chain(
		resolverFunc(func(context.Context, resolve.Request) (resolve.Resource, error) {
			if cancelChild {
				cancel()
			}
			return resolve.Resource{}, resolve.ErrNotFound
		}),
		resolverFunc(func(context.Context, resolve.Request) (resolve.Resource, error) {
			fallbackCalls++
			return resolve.Resource{URI: request.URI, Content: []byte("ordinary")}, nil
		}),
	)

	resource, err := chain.Resolve(ctx, request)
	if err != nil || string(resource.Content) != "ordinary" || fallbackCalls != 1 {
		t.Fatalf("live fallback = %v, %v, calls %d", resource, err, fallbackCalls)
	}
	cancelChild = true
	fallbackCalls = 0
	resource, err = chain.Resolve(ctx, request)
	if err != context.Canceled || !reflect.DeepEqual(resource, resolve.Resource{}) || fallbackCalls != 0 {
		t.Fatalf("canceled fallback = %v, %v, calls %d; want zero resource, cancellation, zero calls", resource, err, fallbackCalls)
	}
	cancelChild = false
	fallbackCalls = 0
	resource, err = chain.Resolve(context.Background(), request)
	if err != nil || string(resource.Content) != "ordinary" || fallbackCalls != 1 {
		t.Fatalf("fresh-context reuse = %v, %v, calls %d", resource, err, fallbackCalls)
	}
}
