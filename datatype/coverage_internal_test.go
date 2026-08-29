package datatype

import "testing"

func TestOrderedWrappersCoverCalendarFormsAndNegativeDurations(t *testing.T) {
	if comparison, comparable := CompareOrdered("duration", "-P1D", "P1D"); !comparable || comparison >= 0 {
		t.Fatalf("CompareOrdered(duration) = %d, %t", comparison, comparable)
	}
	if comparison, comparable := CompareOrdered("date", "invalid", "2000-01-01"); comparable || comparison != 0 {
		t.Fatalf("CompareOrdered(invalid date) = %d, %t", comparison, comparable)
	}

	for _, test := range []struct {
		kind, lexical string
		zoned         bool
	}{
		{kind: "dateTime", lexical: "2000-01-01T00:00:00"},
		{kind: "time", lexical: "12:30:00Z", zoned: true},
		{kind: "date", lexical: "2000-01-01+01:00", zoned: true},
		{kind: "gYearMonth", lexical: "2000-01"},
		{kind: "gYear", lexical: "2000"},
		{kind: "gMonthDay", lexical: "--01-02"},
		{kind: "gDay", lexical: "---02"},
		{kind: "gMonth", lexical: "--02"},
	} {
		if _, ok := parseOrderedCalendar(test.kind, test.lexical); !ok {
			t.Fatalf("parseOrderedCalendar(%q, %q) failed", test.kind, test.lexical)
		}
		canonical, ok := CanonicalOrderedValue(test.kind, test.lexical)
		if !ok || (test.zoned && len(canonical) < len("zoned:")) ||
			(!test.zoned && len(canonical) < len("local:")) {
			t.Fatalf("CanonicalOrderedValue(%q, %q) = %q, %t", test.kind, test.lexical, canonical, ok)
		}
	}
}

func TestPatternEscapesCoverComplementClasses(t *testing.T) {
	for _, source := range []string{`\s`, `\S`, `\I`, `\C`, `\d`, `\D`, `\w`, `\W`} {
		translator := patternTranslator{source: source}
		set, literal, isLiteral, err := translator.parseEscape()
		if err != nil || set == nil || literal != 0 || isLiteral {
			t.Fatalf("parseEscape(%q) = %v, %q, %t, %v", source, set, literal, isLiteral, err)
		}
	}
}
