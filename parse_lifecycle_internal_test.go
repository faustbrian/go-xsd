package xsd

import (
	"context"
	"encoding/xml"
	"errors"
	"testing"
)

func TestParsePreparedOwnerCancellation(t *testing.T) {
	const source = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:annotation><xs:documentation>a</xs:documentation></xs:annotation></xs:schema>`
	for _, boundary := range []string{"initial decoder read", "model admission", "owned token"} {
		t.Run(boundary, func(t *testing.T) {
			for _, canceled := range []bool{false, true} {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				owner := newSchemaParser(ctx, []byte(source), ParseOptions{})
				if err := owner.prepareAnnotationSpans(ParseOptions{}); err != nil {
					cancel()
					t.Fatal(err)
				}
				if len(owner.annotationSpans) != 1 {
					cancel()
					t.Fatal("actual annotation span was not prepared")
				}
				decoder, start := decoderAtStart(t, source)
				if canceled {
					cancel()
				}
				switch boundary {
				case "initial decoder read", "model admission":
					var document *Document
					var err error
					if boundary == "initial decoder read" {
						document, err = parseValidated(ctx, []byte(source), ParseOptions{}, owner)
					} else {
						document, err = owner.parseDocument(decoder, start, "")
					}
					cancel()
					if canceled {
						if document != nil || !errors.Is(err, context.Canceled) {
							t.Fatal("prepared owner did not discard canceled document")
						}
					} else if err != nil || document == nil || len(document.Annotations) != 1 || len(document.Annotations[0].Documentation) != 1 || document.Annotations[0].Documentation[0].Content != "a" {
						t.Fatal("live prepared owner changed document semantics")
					}
				case "owned token":
					token, err := owner.token(decoder)
					cancel()
					if canceled {
						if token != nil || !errors.Is(err, context.Canceled) {
							t.Fatal("prepared owner did not stop token work")
						}
					} else {
						start, ok := token.(xml.StartElement)
						if err != nil || !ok || start.Name != (xml.Name{Space: Namespace, Local: "annotation"}) {
							t.Fatal("live prepared owner changed its next annotation token")
						}
					}
				}
			}
		})
	}
}

func TestParseValidatedWithoutPreparedOwner(t *testing.T) {
	const source = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:annotation><xs:appinfo id="a"/></xs:annotation></xs:schema>`
	document, err := parseValidated(context.Background(), []byte(source), ParseOptions{})
	if err != nil || document == nil || len(document.Annotations) != 1 || len(document.Annotations[0].AppInformation) != 1 || document.Annotations[0].AppInformation[0].ID != "a" {
		t.Fatal("valid private parse without an existing owner lost its document")
	}
}

func TestParseOwnerTokenPreservesAdmissionFailure(t *testing.T) {
	const source = `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:annotation/></xs:schema>`
	owner := newSchemaParser(context.Background(), []byte(source), ParseOptions{MaxModelBytes: 1})
	if owner.chargeBytes(2) || !errors.Is(owner.err, ErrLimitExceeded) {
		t.Fatal("finite owner admission did not establish a limit refusal")
	}
	decoder, _ := decoderAtStart(t, source)
	before := decoder.InputOffset()
	token, err := owner.token(decoder)
	if token != nil || err != owner.err || !errors.Is(err, ErrLimitExceeded) || decoder.InputOffset() != before {
		t.Fatal("sticky admission refusal lost its cause or consumed another token")
	}
}
