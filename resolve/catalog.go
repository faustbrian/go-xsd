package resolve

import (
	"context"
	"fmt"
	"net/url"

	"github.com/faustbrian/go-xsd/v2/internal/errsafe"
)

// Catalog maps import namespaces without schema locations to absolute resource
// identities, then delegates byte loading to another explicit resolver.
type Catalog struct {
	namespaces map[string]string
	resolver   Resolver
}

// CatalogOptions bounds constructor inputs. Zero selects finite defaults:
// 256 mappings and 64 MiB of namespace and resource identity bytes.
// Negative limits are invalid. Limits are inclusive.
type CatalogOptions struct {
	MaxMappings int
	MaxBytes    int64
}

// NewCatalog validates and copies namespace mappings. The delegated resolver
// remains responsible for the mapped resource capability and byte ownership.
func NewCatalog(namespaces map[string]string, resolver Resolver) (catalog *Catalog, err error) {
	return NewCatalogWithOptions(namespaces, resolver, CatalogOptions{})
}

// NewCatalogWithOptions copies mappings with explicitly selected input limits.
func NewCatalogWithOptions(namespaces map[string]string, resolver Resolver, options CatalogOptions) (catalog *Catalog, err error) {
	defer func() { err = errsafe.Wrap("xsd resolve: failed", err) }()
	if options.MaxMappings < 0 || options.MaxBytes < 0 {
		return nil, fmt.Errorf("xsd resolve: invalid catalog limits")
	}
	if options.MaxMappings == 0 {
		options.MaxMappings = 256
	}
	if options.MaxBytes == 0 {
		options.MaxBytes = 64 << 20
	}
	if resolver == nil {
		return nil, fmt.Errorf("xsd resolve: catalog resolver is required")
	}
	if len(namespaces) > options.MaxMappings {
		return nil, ErrLimitExceeded
	}
	remaining := options.MaxBytes
	// Complete admission precedes map capacity and URI parsing.
	for namespace, identity := range namespaces {
		if !consumeInputBytes(&remaining, len(namespace)) || !consumeInputBytes(&remaining, len(identity)) {
			return nil, ErrLimitExceeded
		}
	}
	owned := make(map[string]string, len(namespaces))
	for namespace, identity := range namespaces {
		uri, err := url.Parse(identity)
		if err != nil || !uri.IsAbs() || uri.Fragment != "" {
			return nil, fmt.Errorf("xsd resolve: invalid catalog URI %q", identity)
		}
		owned[namespace] = uri.String()
	}
	return &Catalog{namespaces: owned, resolver: resolver}, nil
}

// Resolve delegates explicit identities unchanged and maps locationless
// imports by their requested namespace.
func (c *Catalog) Resolve(ctx context.Context, request Request) (resource Resource, err error) {
	defer finishResolve(ctx, &resource, &err)
	if err := ctx.Err(); err != nil {
		return Resource{}, err
	}
	if request.URI != "" {
		return c.resolver.Resolve(ctx, request)
	}
	if request.Kind != KindImport {
		return Resource{}, fmt.Errorf("%w: locationless %s", ErrNotFound, request.Kind)
	}
	identity, ok := c.namespaces[request.Namespace]
	if !ok {
		return Resource{}, fmt.Errorf("%w: namespace %s", ErrNotFound, request.Namespace)
	}
	request.URI = identity
	resource, err = c.resolver.Resolve(ctx, request)
	if err != nil {
		return Resource{}, err
	}
	if resource.URI != identity {
		return Resource{}, fmt.Errorf(
			"xsd resolve: catalog requested %q, received %q",
			identity,
			resource.URI,
		)
	}
	return resource, nil
}
