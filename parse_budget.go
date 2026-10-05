package xsd

import (
	"context"
	"encoding/xml"
	"fmt"
)

// schemaParser owns per-document admission, not decoder allocations or heap
// measurement. A refused charge is sticky: no partially built model is published.
type schemaParser struct {
	ctx             context.Context
	source          []byte
	entries         int
	bytes           int64
	err             error
	annotationSpans map[int64]int64
}

func newSchemaParser(ctx context.Context, source []byte, options ParseOptions) *schemaParser {
	entries, bytes := options.MaxNamespaceEntries, options.MaxModelBytes
	if entries == 0 {
		entries = 1000000
	}
	if bytes == 0 {
		bytes = 64 << 20
	}
	return &schemaParser{ctx: ctx, source: source, entries: entries, bytes: bytes, annotationSpans: make(map[int64]int64)}
}

func (p *schemaParser) prepareAnnotationSpans(options ParseOptions) error {
	if options.MaxDepth == 0 {
		options.MaxDepth = defaultMaxParseDepth
	}
	if options.MaxElements == 0 {
		options.MaxElements = defaultMaxElements
	}
	return validateAnnotationPlacement(p.ctx, p.source, options, p)
}

func (p *schemaParser) chargeBytes(size int) bool {
	if p.err != nil {
		return false
	}
	if err := p.ctx.Err(); err != nil {
		p.err = err
		return false
	}
	if int64(size) > p.bytes {
		p.err = fmt.Errorf("%w: model string allowance", ErrLimitExceeded)
		return false
	}
	p.bytes -= int64(size)
	return true
}

func (p *schemaParser) retain(value string) string {
	if !p.chargeBytes(len(value)) {
		return ""
	}
	return value
}

// URI serialization can escape each input byte to three bytes. Admit that
// conservative copy-work envelope before parsing/resolution/string building.
func (p *schemaParser) admitURIWork(base, reference string) bool {
	for range 3 {
		if !p.chargeBytes(len(base)) || !p.chargeBytes(len(reference)) {
			return false
		}
	}
	return true
}

func (p *schemaParser) admitNamespace(prefix, uri string) bool {
	if p.err != nil {
		return false
	}
	if p.entries == 0 {
		p.err = fmt.Errorf("%w: namespace entry allowance", ErrLimitExceeded)
		return false
	}
	if !p.chargeBytes(len(prefix)) || !p.chargeBytes(len(uri)) {
		return false
	}
	p.entries--
	return true
}

func (p *schemaParser) admitNamespaceCopy(scope map[string]string) bool {
	for prefix, uri := range scope {
		if !p.admitNamespace(prefix, uri) {
			return false
		}
	}
	return p.err == nil
}

func (p *schemaParser) admitRootNamespaces(start xml.StartElement) bool {
	for _, attribute := range start.Attr {
		if attribute.Name.Space == "xmlns" {
			if !p.admitNamespace(attribute.Name.Local, attribute.Value) {
				return false
			}
		} else if attribute.Name.Space == "" && attribute.Name.Local == "xmlns" {
			if !p.admitNamespace("", attribute.Value) {
				return false
			}
		}
	}
	return p.err == nil
}

func (p *schemaParser) token(decoder *xml.Decoder) (xml.Token, error) {
	if p.err != nil {
		return nil, p.err
	}
	if err := p.ctx.Err(); err != nil {
		return nil, err
	}
	return decoder.Token()
}

func (p *schemaParser) admitAnnotationCapture(decoder *xml.Decoder) error {
	size, known := p.annotationSpans[decoder.InputOffset()]
	if !known || size < 0 || size > int64(len(p.source)) {
		return fmt.Errorf("xsd: annotation source span unavailable")
	}
	if !p.chargeBytes(int(size)) {
		return p.err
	}
	return nil
}
