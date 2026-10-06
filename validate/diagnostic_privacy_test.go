package validate_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"testing"

	xsd "github.com/faustbrian/go-xsd/v2"
	"github.com/faustbrian/go-xsd/v2/validate"
)

func TestValidateGeneratedDiagnosticDefaultPrivacy(t *testing.T) {
	v, err := validate.New(validatorSet(t), validate.Options{SystemID: "synthetic-private-instance"})
	if err != nil {
		t.Fatal(err)
	}
	result, err := v.Validate(context.Background(), []byte(`<amount xmlns="urn:order">ordinary-not-a-decimal</amount>`))
	if err != nil || result.Valid || len(result.Diagnostics) != 1 {
		t.Fatal("ordinary invalid decimal did not produce one finding")
	}
	diagnostic := result.Diagnostics[0]
	if diagnostic.Severity != xsd.SeverityError || diagnostic.Code != "cvc-datatype-valid.1.2.1" || diagnostic.Message != "value is not valid for decimal" || diagnostic.Path != "/amount" || diagnostic.Location.SystemID != "synthetic-private-instance" || diagnostic.Location.Line != 1 || diagnostic.Location.Column <= 0 || diagnostic.Location.Offset <= 0 {
		t.Fatal("trusted generated finding classification/location changed")
	}
	for _, value := range []any{result, &result, result.Diagnostics, diagnostic, &diagnostic, diagnostic.Location, &diagnostic.Location} {
		for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
			assertGeneratedPrivacy(t, fmt.Sprintf(format, value))
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		assertGeneratedPrivacy(t, string(encoded))
		for _, text := range []bool{false, true} {
			var output bytes.Buffer
			var handler slog.Handler = slog.NewJSONHandler(&output, nil)
			if text {
				handler = slog.NewTextHandler(&output, nil)
			}
			slog.New(handler).Info("operation", slog.Any("result", value))
			assertGeneratedPrivacy(t, output.String())
		}
	}
	// Detailed reporting is a deliberate field projection, not default struct output.
	trusted := struct{ Code, Message, SystemID string }{diagnostic.Code, diagnostic.Message, diagnostic.Location.SystemID}
	encoded, err := json.Marshal(trusted)
	if err != nil || !strings.Contains(string(encoded), "synthetic-private-instance") || !strings.Contains(string(encoded), diagnostic.Code) {
		t.Fatal("explicit trusted reporting projection unavailable")
	}
}

func assertGeneratedPrivacy(t *testing.T, output string) {
	t.Helper()
	for _, private := range []string{"synthetic-private-instance", "cvc-datatype-valid.1.2.1", "value is not valid for decimal", "/amount", "ordinary-not-a-decimal", "SystemID", "Column", "Offset"} {
		if strings.Contains(output, private) {
			t.Error("default generated finding output exposed a field")
			return
		}
	}
	if !strings.Contains(output, "xsd: diagnostic") && !strings.Contains(output, "xsd: location") {
		t.Error("default generated finding output lost category")
	}
}
