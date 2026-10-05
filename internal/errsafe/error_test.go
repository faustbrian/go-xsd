package errsafe_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/faustbrian/go-xsd/internal/errsafe"
)

func TestDefaultOutputRetainsOnlyCategoryAndTrustedCause(t *testing.T) {
	for _, cause := range []error{nil, context.Canceled, context.DeadlineExceeded} {
		if errsafe.Wrap("safe", cause) != cause {
			t.Fatal("nil/standard context cause identity changed")
		}
	}
	cause := errors.New("synthetic-private-cause")
	err := errsafe.Wrap("safe", cause)
	if err.Error() != "safe" || !errors.Is(err, cause) || errors.Unwrap(err) != cause {
		t.Fatal("safe category or trusted cause changed")
	}
	for _, format := range []string{"%v", "%+v", "%#v", "%s", "%q"} {
		want := "safe"
		if format == "%q" {
			want = `"safe"`
		}
		if fmt.Sprintf(format, err) != want {
			t.Fatal("supported formatting lost category")
		}
	}
	encoded, e := json.Marshal(err)
	if e != nil || string(encoded) != `"safe"` {
		t.Fatal("JSON lost category")
	}
}
