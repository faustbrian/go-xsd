package xsd_test

import (
	"testing"
	"time"

	xsd "github.com/faustbrian/go-xsd"
)

func TestNormalizeIdentityXPathWhitespace(t *testing.T) {
	t.Parallel()

	for _, test := range []struct {
		input string
		want  string
	}{
		{input: "  . // p:item / @ id | child  ", want: ".//p:item/@id|child"},
		{input: "a b", want: "a b"},
		{input: "p :name", want: "p :name"},
		{input: "\t.\n", want: "."},
	} {
		done := make(chan string, 1)
		go func() { done <- xsd.NormalizeIdentityXPath(test.input) }()
		var got string
		select {
		case got = <-done:
		case <-time.After(time.Second):
			t.Fatalf("NormalizeIdentityXPath(%q) did not terminate", test.input)
		}
		if got != test.want {
			t.Fatalf("NormalizeIdentityXPath(%q) = %q, want %q", test.input, got, test.want)
		}
	}
}
