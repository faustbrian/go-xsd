package datatype_test

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/faustbrian/go-xsd/v2/datatype"
)

func nestedSubtraction(depth int) string {
	pattern := "[b]"
	for i := 1; i < depth; i++ {
		pattern = "[a-z-" + pattern + "]"
	}
	return pattern
}

func TestCompilePatternContextCancellation(t *testing.T) {
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	expired, expire := context.WithDeadline(context.Background(), time.Unix(0, 0))
	defer expire()
	for _, test := range []struct {
		ctx  context.Context
		want error
	}{{canceled, context.Canceled}, {expired, context.DeadlineExceeded}} {
		compiled, err := datatype.CompilePatternContext(test.ctx, `[a-z-[aeiou]]+`)
		if compiled != nil || !errors.Is(err, test.want) {
			t.Fatal("context cause or zero-regexp refusal changed")
		}
	}
	compiled, err := datatype.CompilePatternContext(context.Background(), `[a-z-[aeiou]]+`)
	if err != nil || compiled == nil || !compiled.MatchString("xsd") || compiled.MatchString("schema") {
		t.Fatal("uncanceled context changed subtraction semantics")
	}
}

func TestCompilePatternClassDepthAdmission(t *testing.T) {
	for _, depth := range []int{1, 2, 3, 256, 257} {
		pattern := nestedSubtraction(depth)
		compiled, err := datatype.CompilePattern(pattern)
		if depth == 257 {
			if compiled != nil || err == nil || !strings.Contains(err.Error(), "depth exceeds 256") {
				t.Fatal("over-depth ordinary subtraction was not refused without a partial regexp")
			}
			continue
		}
		if err != nil || compiled == nil {
			t.Fatal("admitted ordinary subtraction failed")
		}
		if compiled.MatchString("a") != (depth%2 == 0) || compiled.MatchString("b") != (depth%2 == 1) {
			t.Fatal("subtraction parity changed")
		}
	}
}

func TestCompilePatternSiblingClassesDoNotAccumulateDepth(t *testing.T) {
	for _, compile := range []struct {
		name string
		fn   func(string) (*regexp.Regexp, error)
	}{
		{"ordinary", datatype.CompilePattern},
		{"context", func(pattern string) (*regexp.Regexp, error) {
			return datatype.CompilePatternContext(context.Background(), pattern)
		}},
	} {
		t.Run(compile.name, func(t *testing.T) {
			pattern := strings.Repeat("[a-z-[aeiou]]", 256)
			compiled, err := compile.fn(pattern)
			if err != nil || compiled == nil {
				t.Fatalf("independent depth-two classes were refused: %v", err)
			}
			if !compiled.MatchString(strings.Repeat("b", 256)) ||
				compiled.MatchString(strings.Repeat("b", 255)) ||
				compiled.MatchString(strings.Repeat("b", 255)+"a") {
				t.Fatal("sibling class count or subtraction semantics changed")
			}
		})
	}
}
