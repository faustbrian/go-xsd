package xsd_test

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
)

func TestParseAppInfoMetadataNamespaces(t *testing.T) {
	for _, test := range []struct {
		name, attributes, id, source string
	}{
		{"foreign only", `e:id="a" e:source="b"`, "", ""},
		{"foreign after standard", `id="a" source="b" e:id="bb" e:source="cc"`, "a", "b"},
		{"foreign before standard", `e:id="bb" e:source="cc" id="a" source="b"`, "a", "b"},
		{"standard only", `id="a" source="b"`, "a", "b"},
	} {
		t.Run(test.name, func(t *testing.T) {
			source := `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:annotation><xs:appinfo xmlns:e="urn:extension" ` + test.attributes + `><tool/></xs:appinfo></xs:annotation></xs:schema>`
			document, err := xsd.Parse(context.Background(), []byte(source), xsd.ParseOptions{})
			if err != nil || document == nil || len(document.Annotations) != 1 || len(document.Annotations[0].AppInformation) != 1 {
				t.Fatal("accepted appinfo did not produce one annotation")
			}
			appInfo := document.Annotations[0].AppInformation[0]
			if appInfo.ID != test.id || appInfo.Source != test.source || appInfo.Content != "<tool/>" {
				t.Fatal("foreign attributes changed standard appinfo metadata or content")
			}
		})
	}
}

func TestParseAppInfoMetadataAdmission(t *testing.T) {
	for _, attributes := range []string{`id="a"`, `source="b"`, `id="a" source="b"`} {
		t.Run(attributes, func(t *testing.T) {
			source := `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:annotation><xs:appinfo ` + attributes + `/></xs:annotation></xs:schema>`
			metadataBytes := int64(1)
			if attributes == `id="a" source="b"` {
				metadataBytes = 2
			}
			maximum := int64(2+len(xsd.Namespace)) + metadataBytes
			for _, limit := range []int64{maximum - 1, maximum} {
				document, err := xsd.Parse(context.Background(), []byte(source), xsd.ParseOptions{MaxModelBytes: limit})
				if limit < maximum {
					if document != nil || !errors.Is(err, xsd.ErrLimitExceeded) {
						t.Fatal("standard appinfo metadata bypassed its byte allowance")
					}
				} else if err != nil || document == nil || len(document.Annotations) != 1 || len(document.Annotations[0].AppInformation) != 1 || document.Annotations[0].AppInformation[0].ID+document.Annotations[0].AppInformation[0].Source == "" {
					t.Fatal("exact appinfo metadata allowance rejected valid empty content")
				}
			}
		})
	}
	for _, attributes := range []string{``, `xmlns="urn:opaque"`, `xmlns:e="urn:extension" e:id="a" e:source="b"`} {
		source := `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:annotation><xs:appinfo ` + attributes + `/></xs:annotation></xs:schema>`
		document, err := xsd.Parse(context.Background(), []byte(source), xsd.ParseOptions{MaxModelBytes: int64(2 + len(xsd.Namespace)), MaxNamespaceEntries: 1})
		if err != nil || document == nil || len(document.Annotations) != 1 || len(document.Annotations[0].AppInformation) != 1 || document.Annotations[0].AppInformation[0] != (xsd.AppInfo{}) {
			t.Fatal("ignored appinfo declarations or extension metadata consumed model allowance")
		}
	}
}

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
