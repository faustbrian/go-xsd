package datatype

import (
	"errors"
	"fmt"
	"math/big"
	"strings"
)

var (
	// ErrInvalidLexical identifies malformed datatype lexical input; detailed errors
	// may echo that input.
	ErrInvalidLexical = errors.New("xsd datatype: invalid lexical value")
	// ErrFacetViolation identifies a value rejected by a constraining facet.
	ErrFacetViolation = errors.New("xsd datatype: facet violation")
	// ErrLimitExceeded identifies refusal of a configured resource or supported-size
	// bound.
	ErrLimitExceeded = errors.New("xsd datatype: resource limit exceeded")
)

const maxDecimalLexicalBytes = 1 << 20

// Decimal is an immutable arbitrary-precision XML Schema decimal value.
// Its zero value denotes zero. Copies may share a private coefficient, but
// arithmetic helpers copy it before mutation; concurrent reads are safe.
type Decimal struct {
	coefficient *big.Int
	scale       int
}

// ParseDecimal applies decimal's fixed collapse whitespace rule and parses
// its lexical space without floating-point conversion. Inputs above 1 MiB are
// refused before trimming; exponent notation is rejected. Trailing fractional
// zeros are removed exactly and signed zero normalizes to zero. Errors may echo
// malformed lexical input. No rounding or machine-width approximation occurs.
func ParseDecimal(lexical string) (Decimal, error) {
	if len(lexical) > maxDecimalLexicalBytes {
		return Decimal{}, fmt.Errorf("%w: decimal exceeds %d bytes", ErrLimitExceeded, maxDecimalLexicalBytes)
	}
	lexical = trimXMLSpace(lexical)
	if lexical == "" {
		return Decimal{}, invalidDecimal(lexical)
	}

	negative := false
	switch lexical[0] {
	case '+':
		lexical = lexical[1:]
	case '-':
		negative = true
		lexical = lexical[1:]
	}
	if lexical == "" {
		return Decimal{}, invalidDecimal(lexical)
	}

	dot := strings.IndexByte(lexical, '.')
	if strings.Count(lexical, ".") > 1 {
		return Decimal{}, invalidDecimal(lexical)
	}
	integerPart := lexical
	fractionPart := ""
	if dot != -1 {
		integerPart = lexical[:dot]
		fractionPart = lexical[dot+1:]
	}
	if integerPart == "" && fractionPart == "" {
		return Decimal{}, invalidDecimal(lexical)
	}
	if !asciiDigits(integerPart) || !asciiDigits(fractionPart) {
		return Decimal{}, invalidDecimal(lexical)
	}

	digits := integerPart + fractionPart
	coefficient, _ := new(big.Int).SetString(digits, 10)
	scale := len(strings.TrimRight(fractionPart, "0"))
	coefficient.Quo(coefficient, powerOfTen(len(fractionPart)-scale))
	if coefficient.Sign() == 0 {
		scale = 0
		negative = false
	}
	if negative {
		coefficient.Neg(coefficient)
	}
	return Decimal{coefficient: coefficient, scale: scale}, nil
}

func invalidDecimal(lexical string) error {
	return fmt.Errorf("%w: %q is not decimal", ErrInvalidLexical, lexical)
}

func asciiDigits(value string) bool {
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

func trimXMLSpace(value string) string {
	return strings.TrimFunc(value, func(character rune) bool {
		return character == ' ' || character == '\t' || character == '\n' || character == '\r'
	})
}

// String returns the XML Schema canonical representation of the value.
func (d Decimal) String() string {
	if coefficient(d).Sign() == 0 {
		return "0.0"
	}
	absolute := new(big.Int).Abs(d.coefficient).String()
	negative := ""
	if d.coefficient.Sign() == -1 {
		negative = "-"
	}
	if d.scale == 0 {
		return negative + absolute + ".0"
	}
	if len(absolute) <= d.scale {
		absolute = strings.Repeat("0", d.scale-len(absolute)+1) + absolute
	}
	split := len(absolute) - d.scale
	return negative + absolute[:split] + "." + absolute[split:]
}

// Compare returns -1, 0, or 1 according to the decimal value ordering.
func (d Decimal) Compare(other Decimal) int {
	left := coefficient(d)
	right := coefficient(other)
	scaleDifference := d.scale - other.scale
	switch big.NewInt(int64(scaleDifference)).Sign() {
	case -1:
		left.Mul(left, powerOfTen(-scaleDifference))
	case 1:
		right.Mul(right, powerOfTen(scaleDifference))
	}
	return left.Cmp(right)
}

func coefficient(value Decimal) *big.Int {
	if value.coefficient == nil {
		return new(big.Int)
	}
	return new(big.Int).Set(value.coefficient)
}

func powerOfTen(exponent int) *big.Int {
	return new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(exponent)), nil)
}

// TotalDigits reports the smallest number of decimal digits needed for the
// value, excluding sign and decimal point.
func (d Decimal) TotalDigits() int {
	digits := new(big.Int).Abs(coefficient(d)).String()
	if digits == "0" {
		return 1
	}
	return len(digits)
}

// FractionDigits reports the smallest number of fractional digits needed for
// the value.
func (d Decimal) FractionDigits() int { return d.scale }

// DecimalFacets contains decimal's ordered and digit constraining facets.
// Nonpositive TotalDigits disables its check; nil FractionDigits is absent,
// while a pointer to zero requires an integral value. Referenced bounds and
// slices are borrowed during Validate and must not be mutated concurrently.
type DecimalFacets struct {
	// MinInclusive is an optional inclusive lower bound; nil means absent.
	MinInclusive *Decimal
	// MinExclusive is an optional exclusive lower bound; nil means absent.
	MinExclusive *Decimal
	// MaxInclusive is an optional inclusive upper bound; nil means absent.
	MaxInclusive *Decimal
	// MaxExclusive is an optional exclusive upper bound; nil means absent.
	MaxExclusive *Decimal
	// TotalDigits is the maximum digit count; nonpositive disables this check.
	TotalDigits int
	// FractionDigits is an optional maximum scale; nil means absent and a pointer to
	// zero requires an integral value. A negative pointed value rejects every value.
	FractionDigits *int
	// Enumeration admits numerically equal values; an empty slice disables
	// enumeration.
	Enumeration []Decimal
}

// Validate checks a decimal value against every configured facet.
func (f DecimalFacets) Validate(value Decimal) error {
	if f.MinInclusive != nil && value.Compare(*f.MinInclusive) < 0 {
		return facetViolation("minInclusive", value)
	}
	if f.MinExclusive != nil && value.Compare(*f.MinExclusive) <= 0 {
		return facetViolation("minExclusive", value)
	}
	if f.MaxInclusive != nil && value.Compare(*f.MaxInclusive) > 0 {
		return facetViolation("maxInclusive", value)
	}
	if f.MaxExclusive != nil && value.Compare(*f.MaxExclusive) >= 0 {
		return facetViolation("maxExclusive", value)
	}
	if f.TotalDigits > 0 && value.TotalDigits() > f.TotalDigits {
		return facetViolation("totalDigits", value)
	}
	if f.FractionDigits != nil && value.FractionDigits() > *f.FractionDigits {
		return facetViolation("fractionDigits", value)
	}
	if len(f.Enumeration) > 0 {
		for _, allowed := range f.Enumeration {
			if value.Compare(allowed) == 0 {
				return nil
			}
		}
		return facetViolation("enumeration", value)
	}
	return nil
}

func facetViolation(name string, value Decimal) error {
	return fmt.Errorf("%w: %s rejects %s", ErrFacetViolation, name, value.String())
}
