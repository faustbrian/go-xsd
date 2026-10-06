package xsd_test

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
)

func TestParseIncompleteAnnotationsPreserveLocatedSyntaxCause(t *testing.T) {
	t.Parallel()

	for _, kind := range []string{"appinfo", "documentation"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()

			const systemID = "urn:example:annotation"
			lastLine := fmt.Sprintf("<xs:%s>notes", kind)
			source := "<xs:schema xmlns:xs=\"http://www.w3.org/2001/XMLSchema\">\n<xs:annotation>\n" + lastLine
			options := xsd.ParseOptions{SystemID: systemID}
			document, err := xsd.Parse(context.Background(), []byte(source), options)
			if document != nil {
				t.Fatal("incomplete annotation published a document")
			}
			var syntax *xml.SyntaxError
			if !errors.As(err, &syntax) || syntax.Line != 3 {
				t.Fatal("incomplete annotation lost its XML syntax cause")
			}
			var located *xsd.ParseError
			if !errors.As(err, &located) || located.Location != (xsd.Location{
				SystemID: systemID,
				Line:     3,
				Column:   len(lastLine) + 1,
				Offset:   int64(len(source)),
			}) || err.Error() != "xsd: parse failed" {
				t.Fatal("incomplete annotation lost its trusted location or safe category")
			}

			completed := source + fmt.Sprintf("</xs:%s></xs:annotation></xs:schema>", kind)
			document, err = xsd.Parse(context.Background(), []byte(completed), options)
			if err != nil || document == nil || len(document.Annotations) != 1 {
				t.Fatal("completed annotation did not produce one annotation")
			}
			annotation := document.Annotations[0]
			if kind == "appinfo" {
				if len(annotation.AppInformation) != 1 || annotation.AppInformation[0].Content != "notes" || len(annotation.Documentation) != 0 {
					t.Fatal("completed appinfo content changed")
				}
			} else if len(annotation.Documentation) != 1 || annotation.Documentation[0].Content != "notes" || annotation.Documentation[0].Markup != "notes" || len(annotation.AppInformation) != 0 {
				t.Fatal("completed documentation content changed")
			}
		})
	}
}
