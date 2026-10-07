package xsdtest_test

import (
	"context"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/faustbrian/go-xsd/v2/xsdtest"
)

func TestOfficialXSTS(t *testing.T) {
	root := os.Getenv("XSTS_ROOT")
	if root == "" {
		t.Skip("XSTS_ROOT is not set")
	}
	filter := os.Getenv("XSTS_FILTER")
	var metadata []string
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".testSet") &&
			(filter == "" || strings.Contains(path, filter)) {
			relative, err := filepath.Rel(root, path)
			if err != nil {
				return err
			}
			metadata = append(metadata, relative)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(metadata)
	if len(metadata) == 0 {
		t.Fatal("no XSTS .testSet metadata found")
	}
	total := xsdtest.Report{}
	loggedFailures := 0
	for _, testSet := range metadata {
		report, err := xsdtest.Run(context.Background(), root, testSet)
		if err != nil {
			t.Errorf("Run(%s) error = %v", testSet, err)
			continue
		}
		total.Passed += report.Passed
		total.Failed += report.Failed
		total.Skipped += report.Skipped
		total.Excluded += report.Excluded
		for _, testCase := range report.Cases {
			if testCase.Actual != testCase.Expected && testCase.Actual != "skipped" &&
				testCase.Actual != "excluded" &&
				loggedFailures < 50 {
				t.Logf(
					"%s %s/%s: expected %s, got %s: %v",
					testSet,
					testCase.Group,
					testCase.Name,
					testCase.Expected,
					testCase.Actual,
					testCase.Err,
				)
				loggedFailures++
			}
		}
	}
	t.Logf(
		"XSTS passed=%d failed=%d skipped=%d excluded=%d",
		total.Passed,
		total.Failed,
		total.Skipped,
		total.Excluded,
	)
	if total.Failed != 0 {
		t.Fatalf("XSTS has %d blocking failures", total.Failed)
	}
	if total.Skipped != 0 {
		t.Fatalf("XSTS has %d skipped required expectations", total.Skipped)
	}
}

func TestRunExecutesSchemaAndInstanceExpectations(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFixture(t, root, "schemas/value.xsd", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 targetNamespace="urn:test"><xs:element name="value" type="xs:decimal"/></xs:schema>`)
	writeFixture(t, root, "instances/valid.xml", `<value xmlns="urn:test">1.5</value>`)
	writeFixture(t, root, "instances/invalid.xml", `<value xmlns="urn:test">bad</value>`)
	writeFixture(t, root, "meta/tests.testSet", `<testSet xmlns="http://www.w3.org/XML/2004/xml-schema-test-suite/"
 xmlns:xlink="http://www.w3.org/1999/xlink"><testGroup name="decimal">
 <schemaTest name="schema"><schemaDocument xlink:href="../schemas/value.xsd"/>
  <expected validity="valid"/><current status="accepted"/></schemaTest>
 <instanceTest name="valid"><instanceDocument xlink:href="../instances/valid.xml"/>
  <expected validity="valid"/><current status="accepted"/></instanceTest>
 <instanceTest name="invalid"><instanceDocument xlink:href="../instances/invalid.xml"/>
  <expected validity="invalid"/><current status="accepted"/></instanceTest>
</testGroup></testSet>`)

	report, err := xsdtest.Run(context.Background(), root, "meta/tests.testSet")
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed != 3 || report.Failed != 0 || report.Skipped != 0 {
		t.Fatalf("Report = %#v", report)
	}
}

func TestRunConfinesSymlinkReadsToSuiteRoot(t *testing.T) {
	for _, target := range []string{"tests.testSet", "value.xsd", "value.xml", "included.xsd"} {
		for _, link := range []struct {
			name     string
			outside  bool
			absolute bool
		}{
			{name: "inside-relative"},
			{name: "outside-relative", outside: true},
			{name: "outside-absolute", outside: true, absolute: true},
			{name: "inside-absolute", absolute: true},
		} {
			t.Run(target+"/"+link.name, func(t *testing.T) {
				root := t.TempDir()
				fixtures := map[string]string{
					"tests.testSet": `<testSet xmlns="http://www.w3.org/XML/2004/xml-schema-test-suite/"
 xmlns:xlink="http://www.w3.org/1999/xlink"><testGroup name="links">
 <schemaTest name="schema"><schemaDocument xlink:href="value.xsd"/>
 <expected validity="valid"/></schemaTest>
 <instanceTest name="instance"><instanceDocument xlink:href="value.xml"/>
 <expected validity="valid"/></instanceTest></testGroup></testSet>`,
					"value.xsd":    `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:include schemaLocation="included.xsd"/></xs:schema>`,
					"included.xsd": `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="value" type="xs:string"/></xs:schema>`,
					"value.xml":    `<value>harmless fixture</value>`,
				}
				for path, content := range fixtures {
					if path != target {
						writeFixture(t, root, path, content)
					}
				}
				linkedRoot := root
				if link.outside {
					linkedRoot = t.TempDir()
				}
				writeFixture(t, linkedRoot, "linked-fixture", fixtures[target])
				linkTarget, err := filepath.Rel(root, filepath.Join(linkedRoot, "linked-fixture"))
				if err != nil {
					t.Fatal(err)
				}
				if link.absolute {
					linkTarget = filepath.Join(linkedRoot, "linked-fixture")
				}
				if err := os.Symlink(linkTarget, filepath.Join(root, target)); err != nil {
					t.Fatal(err)
				}
				report, err := xsdtest.Run(context.Background(), root, "tests.testSet")
				if !link.outside && !link.absolute {
					if err != nil || report.Passed != 2 || report.Failed != 0 || report.Skipped != 0 {
						t.Fatalf("in-root symlink: report=%#v error=%v", report, err)
					}
					return
				}
				if target == "tests.testSet" {
					if err == nil {
						t.Fatal("accepted metadata outside suite root")
					}
					return
				}
				if err != nil || report.Failed != 1 {
					t.Fatalf("outside fixture was not rejected: report=%#v error=%v", report, err)
				}
				for _, result := range report.Cases {
					if (result.Actual == "invalid" || result.Actual == "error") && result.Err == nil {
						t.Fatal("rejected fixture without a read or resolution error")
					}
				}
			})
		}
	}
}

func TestRunRejectsReferencesOutsideSuiteRoot(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFixture(t, root, "tests.testSet", `<testSet xmlns="http://www.w3.org/XML/2004/xml-schema-test-suite/"
 xmlns:xlink="http://www.w3.org/1999/xlink"><testGroup name="escape">
 <schemaTest name="schema"><schemaDocument xlink:href="../outside.xsd"/>
  <expected validity="valid"/><current status="accepted"/></schemaTest>
</testGroup></testSet>`)
	report, err := xsdtest.Run(context.Background(), root, "tests.testSet")
	if err != nil {
		t.Fatal(err)
	}
	if report.Failed != 1 || report.Cases[0].Err == nil {
		t.Fatalf("Report = %#v", report)
	}
}

func TestRunDoesNotTreatFixtureReadRefusalAsInvalidSchema(t *testing.T) {
	for _, target := range []string{"ordinary-invalid", "value.xsd", "included.xsd"} {
		t.Run(target, func(t *testing.T) {
			root := t.TempDir()
			invalid := `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:element name="value" type="missing"/></xs:schema>`
			writeFixture(t, root, "tests.testSet", `<testSet xmlns="http://www.w3.org/XML/2004/xml-schema-test-suite/"
 xmlns:xlink="http://www.w3.org/1999/xlink"><testGroup name="invalid">
 <schemaTest name="schema"><schemaDocument xlink:href="value.xsd"/>
 <expected validity="invalid"/></schemaTest></testGroup></testSet>`)
			if target == "ordinary-invalid" {
				writeFixture(t, root, "value.xsd", invalid)
			} else {
				if target == "included.xsd" {
					writeFixture(t, root, "value.xsd", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"><xs:include schemaLocation="included.xsd"/></xs:schema>`)
				}
				outside := t.TempDir()
				writeFixture(t, outside, "fixture.xsd", invalid)
				if err := os.Symlink(filepath.Join(outside, "fixture.xsd"), filepath.Join(root, target)); err != nil {
					t.Fatal(err)
				}
			}
			report, err := xsdtest.Run(context.Background(), root, "tests.testSet")
			if err != nil || len(report.Cases) != 1 {
				t.Fatalf("report=%#v error=%v", report, err)
			}
			if target == "ordinary-invalid" {
				if report.Passed != 1 || report.Failed != 0 || report.Cases[0].Actual != "invalid" {
					t.Fatalf("ordinary invalid schema no longer passes: %#v", report)
				}
				return
			}
			if report.Passed != 0 || report.Failed != 1 || report.Cases[0].Actual != "error" || report.Cases[0].Err == nil {
				t.Fatalf("fixture refusal counted as schema invalidity: %#v", report)
			}
		})
	}
}

func TestRunCompilesAllSchemaDocumentsInAGroup(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFixture(t, root, "schemas/types.xsd", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 targetNamespace="urn:types"><xs:simpleType name="Code"><xs:restriction base="xs:string"/></xs:simpleType></xs:schema>`)
	writeFixture(t, root, "schemas/elements.xsd", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 xmlns:t="urn:types" targetNamespace="urn:elements"><xs:element name="value" type="t:Code"/></xs:schema>`)
	writeFixture(t, root, "instances/value.xml", `<value xmlns="urn:elements">ok</value>`)
	writeFixture(t, root, "meta/tests.testSet", `<testSet xmlns="http://www.w3.org/XML/2004/xml-schema-test-suite/"
 xmlns:xlink="http://www.w3.org/1999/xlink"><testGroup name="multiple">
 <schemaTest name="schema"><schemaDocument xlink:href="../schemas/types.xsd"/>
  <schemaDocument xlink:href="../schemas/elements.xsd"/>
  <expected validity="valid"/><current status="accepted"/></schemaTest>
 <instanceTest name="instance"><instanceDocument xlink:href="../instances/value.xml"/>
  <expected validity="valid"/><current status="accepted"/></instanceTest>
</testGroup></testSet>`)

	report, err := xsdtest.Run(context.Background(), root, "meta/tests.testSet")
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed != 2 || report.Failed != 0 || report.Skipped != 0 {
		t.Fatalf("Report = %#v", report)
	}
}

func TestRunCompilesSchemasHintedByAnInstance(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFixture(t, root, "schemas/primary.xsd", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 targetNamespace="urn:primary"><xs:simpleType name="Code"><xs:restriction base="xs:string"/></xs:simpleType></xs:schema>`)
	writeFixture(t, root, "schemas/secondary.xsd", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema"
 targetNamespace="urn:secondary"><xs:element name="value" type="xs:decimal"/></xs:schema>`)
	writeFixture(t, root, "instances/value.xml", `<value xmlns="urn:secondary"
 xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
 xsi:schemaLocation="urn:secondary ../schemas/secondary.xsd">1.5</value>`)
	writeFixture(t, root, "meta/tests.testSet", `<testSet xmlns="http://www.w3.org/XML/2004/xml-schema-test-suite/"
 xmlns:xlink="http://www.w3.org/1999/xlink"><testGroup name="hinted">
 <schemaTest name="schema"><schemaDocument xlink:href="../schemas/primary.xsd"/>
  <expected validity="valid"/><current status="accepted"/></schemaTest>
 <instanceTest name="instance"><instanceDocument xlink:href="../instances/value.xml"/>
  <expected validity="valid"/><current status="accepted"/></instanceTest>
</testGroup></testSet>`)

	report, err := xsdtest.Run(context.Background(), root, "meta/tests.testSet")
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed != 2 || report.Failed != 0 || report.Skipped != 0 {
		t.Fatalf("Report = %#v", report)
	}
}

func TestRunExcludesQueriedExpectations(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFixture(t, root, "value.xsd", `<xs:schema xmlns:xs="http://www.w3.org/2001/XMLSchema">
 <xs:element name="value" type="xs:string"/>
</xs:schema>`)
	writeFixture(t, root, "tests.testSet", `<testSet xmlns="http://www.w3.org/XML/2004/xml-schema-test-suite/"
 xmlns:xlink="http://www.w3.org/1999/xlink"><testGroup name="queried">
 <schemaTest name="schema"><schemaDocument xlink:href="value.xsd"/>
  <expected validity="valid"/><current status="queried"/></schemaTest>
</testGroup></testSet>`)

	report, err := xsdtest.Run(context.Background(), root, "tests.testSet")
	if err != nil {
		t.Fatal(err)
	}
	if report.Passed != 0 || report.Failed != 0 || report.Skipped != 0 || report.Excluded != 1 {
		t.Fatalf("Report = %#v", report)
	}
}

func writeFixture(t *testing.T, root string, name string, content string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}
