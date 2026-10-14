package xsdtest_test

import (
	"context"
	"encoding/xml"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/faustbrian/go-xsd/v2/xsdtest"
)

// Run reads caller-trusted metadata before its group cancellation boundary.
// This is a bounded startup/outcome characterization, not hostile ingestion
// qualification or a new official-suite conformance result.
func TestR32TrustedMetadataStartupAndGroupCancellation(t *testing.T) {
	for _, malformed := range []bool{false, true} {
		root := t.TempDir()
		metadata := `<testSet xmlns="http://www.w3.org/XML/2004/xml-schema-test-suite/"><testGroup name="cancel"/></testSet>`
		if malformed {
			metadata = `<testSet><testGroup`
		}
		if err := os.WriteFile(filepath.Join(root, "small.testSet"), []byte(metadata), 0600); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		report, err := xsdtest.Run(ctx, root, "small.testSet")
		if malformed {
			var syntax *xml.SyntaxError
			if !errors.As(err, &syntax) || errors.Is(err, context.Canceled) {
				t.Fatalf("trusted startup syntax cause missing: %v", err)
			}
		} else if !errors.Is(err, context.Canceled) {
			t.Fatalf("group cancellation cause missing: %v", err)
		}
		if report.Passed != 0 || report.Failed != 0 || report.Skipped != 0 || report.Excluded != 0 || len(report.Cases) != 0 {
			t.Fatal("startup/cancellation published case outcomes")
		}
	}
}
