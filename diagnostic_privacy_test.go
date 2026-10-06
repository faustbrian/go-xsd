package xsd_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
)

func TestDiagnosticAndLocationDefaultPrivacy(t *testing.T) {
	location := xsd.Location{SystemID: "synthetic-private-system", Line: 1731, Column: 2842, Offset: 3953}
	diagnostic := xsd.Diagnostic{Severity: "synthetic-private-severity", Code: "synthetic-private-code", Message: "synthetic-private-message", Path: "synthetic-private-path", Location: location}
	for _, test := range []struct {
		name, category string
		value          any
	}{
		{"location-value", "xsd: location", location}, {"location-pointer", "xsd: location", &location},
		{"diagnostic-value", "xsd: diagnostic", diagnostic}, {"diagnostic-pointer", "xsd: diagnostic", &diagnostic},
		{"location-zero", "xsd: location", xsd.Location{}}, {"location-zero-pointer", "xsd: location", &xsd.Location{}},
		{"diagnostic-zero", "xsd: diagnostic", xsd.Diagnostic{}}, {"diagnostic-zero-pointer", "xsd: diagnostic", &xsd.Diagnostic{}},
	} {
		t.Run(test.name, func(t *testing.T) {
			for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q", "%+q", "%#q", "%b", "%c", "%d", "%e", "%E", "%f", "%F", "%g", "%G", "%o", "%O", "%U", "%x", "%X", "%z", "%20v", "%.2v"} {
				want := test.category
				if strings.HasSuffix(format, "q") {
					want = fmt.Sprintf("%q", test.category)
				}
				if fmt.Sprintf(format, test.value) != want {
					t.Errorf("default %s did not return fixed category", format)
				}
			}
			// Intrinsic pointer formatting is supported only for pointers, not struct values.
			intrinsic := []string{"%T"}
			if reflect.TypeOf(test.value).Kind() == reflect.Pointer {
				intrinsic = append(intrinsic, "%p")
			}
			for _, format := range intrinsic {
				if strings.Contains(fmt.Sprintf(format, test.value), "synthetic-private-") {
					t.Error("intrinsic type/address verb traversed fields")
				}
			}
			encoded, err := json.Marshal(test.value)
			if err != nil || string(encoded) != fmt.Sprintf("%q", test.category) {
				t.Error("default JSON did not return fixed category")
			}
			for _, text := range []bool{false, true} {
				var output bytes.Buffer
				// Time is logging-envelope metadata, not a diagnostic coordinate.
				options := &slog.HandlerOptions{ReplaceAttr: func(groups []string, attr slog.Attr) slog.Attr {
					if len(groups) == 0 && attr.Key == slog.TimeKey {
						return slog.Attr{}
					}
					return attr
				}}
				var handler slog.Handler = slog.NewJSONHandler(&output, options)
				if text {
					handler = slog.NewTextHandler(&output, options)
				}
				slog.New(handler).Info("operation", slog.Any("finding", test.value))
				if !strings.Contains(output.String(), test.category) || strings.Contains(output.String(), "synthetic-private-") || strings.Contains(output.String(), "1731") || strings.Contains(output.String(), "2842") || strings.Contains(output.String(), "3953") {
					t.Error("default logging exposed fields or lost category")
				}
			}
		})
	}
	if location.SystemID != "synthetic-private-system" || location.Line != 1731 || location.Column != 2842 || location.Offset != 3953 || diagnostic.Severity != "synthetic-private-severity" || diagnostic.Code != "synthetic-private-code" || diagnostic.Message != "synthetic-private-message" || diagnostic.Path != "synthetic-private-path" || diagnostic.Location != location {
		t.Fatal("trusted field inspection changed")
	}
}

func TestDiagnosticAndLocationInvalidFormatBoundary(t *testing.T) {
	// Go 1.27 fmt's invalid-format diagnostics bypass Formatter with erroring set.
	// Preserve the public fields, but never treat this caller misuse as redacted output.
	for _, value := range []any{xsd.Location{SystemID: "synthetic-private-system"}, xsd.Diagnostic{Message: "synthetic-private-message"}} {
		for _, format := range []string{"%p", "%w"} {
			output := fmt.Sprintf(format, value)
			if !strings.Contains(output, "%!"+format[1:]+"(") || !strings.Contains(output, "synthetic-private-") {
				t.Fatal("Go invalid-format behavior changed; revisit documented privacy boundary")
			}
		}
	}
}

func TestDiagnosticAndLocationNilZeroAndNestedPrivacy(t *testing.T) {
	var nilLocation *xsd.Location
	var nilDiagnostic *xsd.Diagnostic
	for _, test := range []struct {
		value any
		want  string
	}{
		{nilLocation, "null"}, {nilDiagnostic, "null"}, {xsd.Location{}, `"xsd: location"`}, {xsd.Diagnostic{}, `"xsd: diagnostic"`},
		{[]xsd.Location{{SystemID: "synthetic-private-system"}}, `["xsd: location"]`},
		{[]xsd.Diagnostic{{Message: "synthetic-private-message"}}, `["xsd: diagnostic"]`},
	} {
		encoded, err := json.Marshal(test.value)
		if err != nil || string(encoded) != test.want {
			t.Error("nil, zero or nested JSON contract failed")
		}
	}
	for _, value := range []any{nilLocation, nilDiagnostic} {
		for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
			if fmt.Sprintf(format, value) != "<nil>" {
				t.Error("nil formatting lost safe placeholder")
			}
		}
	}
	location := xsd.Location{SystemID: "synthetic-private-system", Line: 1731}
	diagnostics := []xsd.Diagnostic{{Message: "synthetic-private-message", Location: location}}
	before := diagnostics[0]
	if got := fmt.Sprintf("%#v", diagnostics); strings.Contains(got, "synthetic-private-") || strings.Contains(got, "1731") {
		t.Error("nested formatting exposed fields")
	}
	if !reflect.DeepEqual(before, diagnostics[0]) {
		t.Error("formatting changed caller fields")
	}
}
