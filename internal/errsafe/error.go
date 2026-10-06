// Package errsafe owns default error-output privacy across public boundaries.
package errsafe

import (
	"context"
	"fmt"
	"strconv"
)

// Wrap retains the original cause without evaluating any of its output hooks.
// Only the two exact standard context errors already have trusted fixed text.
func Wrap(category string, cause error) error {
	if cause == nil || cause == context.Canceled || cause == context.DeadlineExceeded {
		return cause
	}
	return &failure{category: category, cause: cause}
}

type failure struct {
	category string
	cause    error
}

func (e *failure) Error() string { return e.category }
func (e *failure) Unwrap() error { return e.cause }
func (e failure) Format(state fmt.State, verb rune) {
	text := e.category
	if verb == 'q' {
		text = strconv.Quote(text)
	}
	_, _ = state.Write([]byte(text))
}
func (e failure) MarshalJSON() ([]byte, error) { return []byte(strconv.Quote(e.category)), nil }
