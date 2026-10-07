package datatype

import (
	"cmp"
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"
)

const (
	maxPatternSourceBytes     = 1 << 20
	maxTranslatedPatternBytes = 8 << 20
	maxPatternClassDepth      = 256
)

// CompilePattern compiles an XML Schema 1.0 regular expression. The
// translation retains Go's linear-time regular-expression execution while
// implementing XML Schema character classes and whole-literal matching.
func CompilePattern(pattern string) (*regexp.Regexp, error) {
	return CompilePatternContext(context.Background(), pattern)
}

// CompilePatternContext cooperatively cancels owned translation and character
// set work. Class-subtraction depth is limited to 256, counting the outer class
// as one, independently of source and translated byte bounds. Standard-library
// regexp compilation and subsequent matching cannot be interrupted in flight.
func CompilePatternContext(ctx context.Context, pattern string) (*regexp.Regexp, error) {
	if err := patternContextError(ctx); err != nil {
		return nil, err
	}
	if !withinByteLimit(len(pattern), maxPatternSourceBytes) {
		return nil, fmt.Errorf("xsd: pattern exceeds %d bytes", maxPatternSourceBytes)
	}
	translator := patternTranslator{source: pattern, ctx: ctx}
	translated, err := translator.translate()
	if canceled := patternContextError(ctx); canceled != nil {
		return nil, canceled
	}
	if err != nil {
		return nil, err
	}
	compiled, err := regexp.Compile(`^(?:` + translated + `)$`)
	if canceled := patternContextError(ctx); canceled != nil {
		return nil, canceled
	}
	return compiled, err
}

// Private context-free callers retain the same set algebra (including init).
func patternContextError(contexts ...context.Context) error {
	if len(contexts) != 0 && contexts[0] != nil {
		return contexts[0].Err()
	}
	return nil
}

type patternTranslator struct {
	source     string
	index      int
	ctx        context.Context
	classDepth int
}

func (translator *patternTranslator) translate() (string, error) {
	var translated strings.Builder
	for translator.index < len(translator.source) {
		if err := patternContextError(translator.ctx); err != nil {
			return "", err
		}
		character, size := utf8.DecodeRuneInString(translator.source[translator.index:])
		if character == utf8.RuneError && size == 1 {
			return "", fmt.Errorf("xsd: pattern is not valid UTF-8 at byte %d", translator.index)
		}
		switch character {
		case '[':
			class, err := translator.parseClass()
			if err != nil {
				return "", err
			}
			translated.WriteString(class.regexp(translator.ctx))
		case '\\':
			class, literal, isLiteral, err := translator.parseEscape()
			if err != nil {
				return "", err
			}
			if isLiteral {
				translated.WriteString(regexp.QuoteMeta(string(literal)))
			} else {
				translated.WriteString(class.regexp(translator.ctx))
			}
		case '.':
			translated.WriteString(xmlUniverse.subtract(runeSet{{'\n', '\n'}, {'\r', '\r'}}, translator.ctx).regexp(translator.ctx))
			translator.index += size
		case '^', '$':
			translated.WriteByte('\\')
			translated.WriteRune(character)
			translator.index += size
		default:
			translated.WriteRune(character)
			translator.index += size
		}
		if !withinByteLimit(translated.Len(), maxTranslatedPatternBytes) {
			return "", fmt.Errorf("xsd: translated pattern exceeds %d bytes", maxTranslatedPatternBytes)
		}
	}
	return translated.String(), nil
}

func (translator *patternTranslator) parseClass() (runeSet, error) {
	if err := patternContextError(translator.ctx); err != nil {
		return nil, err
	}
	if translator.classDepth >= maxPatternClassDepth {
		return nil, fmt.Errorf("xsd: pattern character class depth exceeds %d", maxPatternClassDepth)
	}
	translator.classDepth++
	defer func() { translator.classDepth-- }()
	start := translator.index
	translator.index++
	negated := translator.consume('^')
	var result runeSet
	hasMember := false
	for translator.index < len(translator.source) && translator.source[translator.index] != ']' {
		if err := patternContextError(translator.ctx); err != nil {
			return nil, err
		}
		if strings.HasPrefix(translator.source[translator.index:], "-[") {
			if !hasMember {
				return nil, fmt.Errorf("xsd: empty character group at byte %d", start)
			}
			if translator.classDepth >= maxPatternClassDepth {
				return nil, fmt.Errorf("xsd: pattern character class depth exceeds %d", maxPatternClassDepth)
			}
			translator.index++
			excluded, err := translator.parseClass()
			if err != nil {
				return nil, err
			}
			if translator.index >= len(translator.source) {
				return nil, fmt.Errorf("xsd: character class subtraction at byte %d is not final", start)
			}
			if translator.source[translator.index] != ']' {
				return nil, fmt.Errorf("xsd: character class subtraction at byte %d is not final", start)
			}
			translator.index++
			if negated {
				result = xmlUniverse.subtract(result, translator.ctx)
			}
			return result.subtract(excluded, translator.ctx), patternContextError(translator.ctx)
		}

		member, singleton, singletonMember, err := translator.parseClassMember()
		if err != nil {
			return nil, err
		}
		hasMember = true
		if translator.index < len(translator.source) && translator.source[translator.index] == '-' &&
			!strings.HasPrefix(translator.source[translator.index:], "-[") {
			if !singletonMember {
				return nil, fmt.Errorf("xsd: character range at byte %d has a class endpoint", translator.index)
			}
			translator.index++
			_, endSingleton, endIsSingleton, rangeErr := translator.parseClassMember()
			if rangeErr != nil {
				return nil, rangeErr
			}
			if !endIsSingleton {
				return nil, fmt.Errorf("xsd: invalid character range at byte %d", translator.index)
			}
			if singleton > endSingleton {
				return nil, fmt.Errorf("xsd: invalid character range at byte %d", translator.index)
			}
			member = runeSet{{singleton, endSingleton}}
		}
		result = result.union(member, translator.ctx)
	}
	if translator.index >= len(translator.source) {
		return nil, fmt.Errorf("xsd: unterminated or empty character class at byte %d", start)
	}
	if !hasMember {
		return nil, fmt.Errorf("xsd: unterminated or empty character class at byte %d", start)
	}
	translator.index++
	if negated {
		result = xmlUniverse.subtract(result, translator.ctx)
	}
	return result, patternContextError(translator.ctx)
}

func (translator *patternTranslator) parseClassMember() (runeSet, rune, bool, error) {
	if err := patternContextError(translator.ctx); err != nil {
		return nil, 0, false, err
	}
	if translator.index == len(translator.source) {
		return nil, 0, false, fmt.Errorf("xsd: missing character class member at byte %d", translator.index)
	}
	if translator.source[translator.index] == ']' {
		return nil, 0, false, fmt.Errorf("xsd: missing character class member at byte %d", translator.index)
	}
	if translator.source[translator.index] == '[' || translator.source[translator.index] == '-' {
		return nil, 0, false, fmt.Errorf("xsd: unescaped character at byte %d", translator.index)
	}
	if translator.source[translator.index] == '\\' {
		class, literal, isLiteral, err := translator.parseEscape()
		if err != nil {
			return nil, 0, false, err
		}
		if !isLiteral {
			return class, 0, false, nil
		}
		return runeSet{{literal, literal}}, literal, true, nil
	}
	character, size := utf8.DecodeRuneInString(translator.source[translator.index:])
	if character == utf8.RuneError && size == 1 {
		return nil, 0, false, fmt.Errorf("xsd: pattern is not valid UTF-8 at byte %d", translator.index)
	}
	translator.index += size
	return runeSet{{character, character}}, character, true, nil
}

func (translator *patternTranslator) parseEscape() (runeSet, rune, bool, error) {
	if err := patternContextError(translator.ctx); err != nil {
		return nil, 0, false, err
	}
	start := translator.index
	translator.index++
	if translator.index == len(translator.source) {
		return nil, 0, false, fmt.Errorf("xsd: incomplete escape at byte %d", start)
	}
	escape, size := utf8.DecodeRuneInString(translator.source[translator.index:])
	translator.index += size
	switch escape {
	case 'n':
		return nil, '\n', true, nil
	case 'r':
		return nil, '\r', true, nil
	case 't':
		return nil, '\t', true, nil
	case '\\', '|', '.', '-', '^', '?', '*', '+', '(', ')', '{', '}', '[', ']':
		return nil, escape, true, nil
	case 's':
		return runeSet{{'\t', '\n'}, {'\r', '\r'}, {' ', ' '}}, 0, false, nil
	case 'S':
		return xmlUniverse.subtract(runeSet{{'\t', '\n'}, {'\r', '\r'}, {' ', ' '}}, translator.ctx), 0, false, nil
	case 'i':
		return xmlNameStartSet, 0, false, nil
	case 'I':
		return xmlUniverse.subtract(xmlNameStartSet, translator.ctx), 0, false, nil
	case 'c':
		return xmlNameChar, 0, false, nil
	case 'C':
		return xmlUniverse.subtract(xmlNameChar, translator.ctx), 0, false, nil
	case 'd':
		return categorySet("Nd", translator.ctx), 0, false, nil
	case 'D':
		return xmlUniverse.subtract(categorySet("Nd", translator.ctx), translator.ctx), 0, false, nil
	case 'w':
		return xmlWord, 0, false, nil
	case 'W':
		return xmlUniverse.subtract(xmlWord, translator.ctx), 0, false, nil
	case 'p', 'P':
		property, err := translator.parseProperty(start)
		if err != nil {
			return nil, 0, false, err
		}
		if escape == 'P' {
			property = xmlUniverse.subtract(property, translator.ctx)
		}
		return property, 0, false, nil
	default:
		return nil, 0, false, fmt.Errorf("xsd: invalid escape \\%c at byte %d", escape, start)
	}
}

func withinByteLimit(length int, limit int) bool { return length <= limit }

func (translator *patternTranslator) parseProperty(start int) (runeSet, error) {
	if !translator.consume('{') {
		return nil, fmt.Errorf("xsd: property escape at byte %d is missing '{'", start)
	}
	end := strings.IndexByte(translator.source[translator.index:], '}')
	if end == -1 {
		return nil, fmt.Errorf("xsd: unterminated property escape at byte %d", start)
	}
	name := translator.source[translator.index : translator.index+end]
	translator.index += end + 1
	if strings.HasPrefix(name, "Is") {
		block, ok := xsdUnicodeBlocks[name[2:]]
		if !ok {
			return nil, fmt.Errorf("xsd: unknown Unicode block %q at byte %d", name[2:], start)
		}
		return block, nil
	}
	category, ok := unicode.Categories[name]
	if !ok {
		return nil, fmt.Errorf("xsd: unknown Unicode category %q at byte %d", name, start)
	}
	return rangeTableSet(category, translator.ctx), patternContextError(translator.ctx)
}

func (translator *patternTranslator) consume(character byte) bool {
	if translator.index < len(translator.source) && translator.source[translator.index] == character {
		translator.index++
		return true
	}
	return false
}

type runeRange struct {
	first rune
	last  rune
}

type runeSet []runeRange

var xmlUniverse = runeSet{
	{0x9, 0xA}, {0xD, 0xD}, {0x20, 0xD7FF}, {0xE000, 0xFFFD}, {0x10000, utf8.MaxRune},
}

var xmlNameStartSet = runeSet{
	{':', ':'}, {'A', 'Z'}, {'_', '_'}, {'a', 'z'}, {0xC0, 0xD6}, {0xD8, 0xF6},
	{0xF8, 0x2FF}, {0x370, 0x37D}, {0x37F, 0x1FFF}, {0x200C, 0x200D},
	{0x2070, 0x218F}, {0x2C00, 0x2FEF}, {0x3001, 0xD7FF}, {0xF900, 0xFDCF},
	{0xFDF0, 0xFFFD}, {0x10000, 0xEFFFF},
}

var xmlNameChar = xmlNameStartSet.union(runeSet{
	{'-', '.'}, {'0', '9'}, {0xB7, 0xB7}, {0x300, 0x36F}, {0x203F, 0x2040},
})

var xmlWord = xmlUniverse.subtract(categorySet("P").union(categorySet("Z")).union(categorySet("C")))

func categorySet(name string, contexts ...context.Context) runeSet {
	return rangeTableSet(unicode.Categories[name], contexts...)
}

func rangeTableSet(table *unicode.RangeTable, contexts ...context.Context) runeSet {
	if patternContextError(contexts...) != nil {
		return nil
	}
	result := make(runeSet, 0, len(table.R16)+len(table.R32))
	for _, item := range table.R16 {
		if patternContextError(contexts...) != nil {
			return nil
		}
		low, high, stride := uint32(item.Lo), uint32(item.Hi), uint32(item.Stride)
		count, ok := boundedRangeCount(low, high, stride, 1<<16)
		if !ok {
			return nil
		}
		for offset := range count {
			if patternContextError(contexts...) != nil {
				return nil
			}
			value := low + offset*stride
			result = append(result, runeRange{rune(value), rune(value)}) // #nosec G115 -- R16 stride expansion remains at or below its uint16 Hi endpoint, within rune range.
		}
	}
	for _, item := range table.R32 {
		if patternContextError(contexts...) != nil {
			return nil
		}
		count, ok := boundedRangeCount(
			item.Lo,
			item.Hi,
			item.Stride,
			0x110000,
		)
		if !ok {
			return nil
		}
		for offset := range count {
			if patternContextError(contexts...) != nil {
				return nil
			}
			value := item.Lo + offset*item.Stride
			result = append(result, runeRange{rune(value), rune(value)}) // #nosec G115 -- Callers use standard unicode.Categories tables; stride expansion stays at or below Hi <= utf8.MaxRune.
		}
	}
	return result.normalized(contexts...).intersect(xmlUniverse, contexts...)
}

func boundedRangeCount(low, high, stride, limit uint32) (uint32, bool) {
	if stride == 0 || high < low {
		return 0, false
	}
	count := (high-low)/stride + 1
	if count > limit {
		return 0, false
	}
	return count, true
}

func (set runeSet) union(other runeSet, contexts ...context.Context) runeSet {
	if patternContextError(contexts...) != nil {
		return nil
	}
	combined := append(append(runeSet(nil), set...), other...)
	return combined.normalized(contexts...)
}

func (set runeSet) intersect(other runeSet, contexts ...context.Context) runeSet {
	set = set.normalized(contexts...)
	other = other.normalized(contexts...)
	result := make(runeSet, 0)
	for len(set) != 0 && len(other) != 0 {
		if patternContextError(contexts...) != nil {
			return nil
		}
		first := max(set[0].first, other[0].first)
		last := min(set[0].last, other[0].last)
		if first <= last {
			result = append(result, runeRange{first, last})
		}
		switch cmp.Compare(set[0].last, other[0].last) {
		case -1:
			set = set[1:]
		case 1:
			other = other[1:]
		default:
			set = set[1:]
			other = other[1:]
		}
	}
	return result
}

func (set runeSet) subtract(other runeSet, contexts ...context.Context) runeSet {
	set = set.normalized(contexts...)
	other = other.normalized(contexts...)
	result := make(runeSet, 0, len(set))
	for _, candidate := range set {
		if patternContextError(contexts...) != nil {
			return nil
		}
		cursor := candidate.first
		for _, excluded := range other {
			if patternContextError(contexts...) != nil {
				return nil
			}
			if excluded.last >= cursor && excluded.first <= candidate.last {
				if excluded.first > cursor {
					result = append(result, runeRange{cursor, excluded.first - 1})
				}
				cursor = excluded.last + 1
			}
		}
		if cursor <= candidate.last {
			result = append(result, runeRange{cursor, candidate.last})
		}
	}
	return result
}

func (set runeSet) normalized(contexts ...context.Context) runeSet {
	if patternContextError(contexts...) != nil {
		return nil
	}
	if len(set) < 2 {
		return append(runeSet(nil), set...)
	}
	result := append(runeSet(nil), set...)
	sort.Slice(result, func(left, right int) bool {
		return cmp.Or(
			cmp.Compare(result[left].first, result[right].first),
			cmp.Compare(result[left].last, result[right].last),
		) == -1
	})
	merged := result[:0]
	for _, item := range result {
		if patternContextError(contexts...) != nil {
			return nil
		}
		if len(merged) == 0 || int64(item.first) > int64(merged[len(merged)-1].last)+1 {
			merged = append(merged, item)
			continue
		}
		merged[len(merged)-1].last = max(item.last, merged[len(merged)-1].last)
	}
	return merged
}

func (set runeSet) regexp(contexts ...context.Context) string {
	var expression strings.Builder
	expression.WriteByte('[')
	for _, item := range set.normalized(contexts...) {
		if patternContextError(contexts...) != nil {
			return ""
		}
		fmt.Fprintf(&expression, `\x{%X}`, item.first)
		if item.last != item.first {
			fmt.Fprintf(&expression, `-\x{%X}`, item.last)
		}
	}
	expression.WriteByte(']')
	return expression.String()
}
