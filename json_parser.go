package partialjson

/*
 * Copyright (c) 2025 shado1111w.
 * Licensed under the MIT License.
 * See LICENSE file in the project root for full license information.
 */

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var (
	// ErrIncompleteString is returned when the string is incomplete
	ErrIncompleteString = errors.New("incomplete string")
	// ErrUnexpectedToken is returned when an unexpected token is encountered
	ErrUnexpectedToken = errors.New("unexpected token")
	// ErrIncompleteNum is returned when the number is incomplete
	ErrIncompleteNum = errors.New("incomplete num")
)

var (
	res = []struct {
		regexp *regexp.Regexp
		repl   func(string) string
	}{
		{
			regexp: regexp.MustCompile(`,\{\}[\]\}]+$`),
			repl:   func(s string) string { return strings.ReplaceAll(s, ",{}", "") },
		},
		{
			regexp: regexp.MustCompile(`\[\{\}\][\]\}]+$`),
			repl:   func(s string) string { return strings.ReplaceAll(s, "[{}]", "null") },
		},
	}
)

// JSONParser is a parser for JSON data
type JSONParser struct {
	strict       bool
	parsers      map[rune]func(string) (any, string, error)
	onExtraToken func(string, any, string)
}

// NewJSONParser creates a JSONParser
func NewJSONParser(strict bool, opts ...ParserOption) *JSONParser {
	parser := &JSONParser{
		strict:  strict,
		parsers: make(map[rune]func(string) (any, string, error)),
	}

	for _, opt := range opts {
		opt(parser)
	}
	parser.parsers[' '] = parser.parseSpace
	parser.parsers['\r'] = parser.parseSpace
	parser.parsers['\n'] = parser.parseSpace
	parser.parsers['\t'] = parser.parseSpace
	parser.parsers['['] = parser.parseArray
	parser.parsers['{'] = parser.parseObject
	parser.parsers['"'] = parser.parseString
	parser.parsers['t'] = parser.parseTrue
	parser.parsers['f'] = parser.parseFalse
	parser.parsers['n'] = parser.parseNull
	for _, c := range "0123456789.-" {
		parser.parsers[c] = parser.parseNumber
	}

	return parser
}

// ParserOption is a function that sets an option on a JSONParser
type ParserOption func(*JSONParser)

// WithOnExtraToken sets the onExtraToken function on a JSONParser
func WithOnExtraToken(fn func(text string, data any, remaining string)) ParserOption {
	return func(p *JSONParser) {
		p.onExtraToken = fn
	}
}

// WithDefaultOnExtraToken sets the default onExtraToken function on a JSONParser
func WithDefaultOnExtraToken() ParserOption {
	return func(p *JSONParser) {
		p.onExtraToken = p.defaultOnExtraToken
	}
}

// Unmarshal unmarshal JSON data into a value
func (p *JSONParser) Unmarshal(data []byte, v any) error {
	jsonData, err := p.EnsureJSON(string(data))
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(jsonData), v)
}

// FastUnmarshal unmarshal JSON data into a value
func (p *JSONParser) FastUnmarshal(data []byte, v any) error {
	jsonData, err := p.FastEnsureJSON(string(data))
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(jsonData), v)
}

// EnsureJSON return a valid JSON string
func (p *JSONParser) EnsureJSON(s string) (string, error) {
	data, err := p.parse(s)
	if err != nil {
		return "", err
	}

	b, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// FastEnsureJSON return a valid JSON string
func (p *JSONParser) FastEnsureJSON(s string) (ret string, err error) {
	if len(s) == 0 {
		err = ErrUnexpectedToken
		return
	}

	defer func() {
		if err == nil {
			for _, re := range res {
				ret = re.regexp.ReplaceAllStringFunc(ret, re.repl)
			}
		}
	}()

	var leftDelimIndexes []int
	isInQuotes := false
	src := []rune(s)
	for i, char := range src {
		if char == '"' && (i == 0 || src[i-1] != '\\') {
			isInQuotes = !isInQuotes
		}

		if !isInQuotes {
			if char == '{' || char == '[' {
				leftDelimIndexes = append(leftDelimIndexes, i)
			}

			if char == '}' || char == ']' {
				if len(leftDelimIndexes) == 0 || src[leftDelimIndexes[len(leftDelimIndexes)-1]] != getReverseDelim(char) {
					err = ErrUnexpectedToken
					return
				}

				leftDelimIndexes = leftDelimIndexes[:len(leftDelimIndexes)-1]
			}
		}
	}

	if len(leftDelimIndexes) == 0 {
		ret = string(src)
		return
	}

	start := len(leftDelimIndexes) - 1
	remaining := string(src[leftDelimIndexes[start]:])
	jsonData, err := p.EnsureJSON(remaining)
	if err != nil {
		return
	}

	src = append(src[:leftDelimIndexes[start]], []rune(jsonData)...)
	leftDelimIndexes = leftDelimIndexes[:start]
	if len(leftDelimIndexes) == 0 {
		ret = string(src)
		return
	}

	delims := make([]rune, 0, len(leftDelimIndexes))
	for i := len(leftDelimIndexes) - 1; i >= 0; i-- {
		d := leftDelimIndexes[i]
		delims = append(delims, getReverseDelim(src[d]))
	}
	src = append(src, delims...)

	ret = string(src)
	return
}

// parse parses a JSON string
func (p *JSONParser) parse(s string) (any, error) {
	if len(s) == 0 {
		return nil, ErrUnexpectedToken
	}

	if !(strings.HasPrefix(s, "{") || strings.HasPrefix(s, "[")) {
		return nil, ErrUnexpectedToken
	}

	if strings.HasSuffix(s, "}") || strings.HasSuffix(s, "]") {
		data := make(map[string]any)
		err := json.Unmarshal([]byte(s), &data)
		if err == nil {
			return data, nil
		}
	}

	data, reminding, err := p.parseAny(s)
	if p.onExtraToken != nil && reminding != "" {
		p.onExtraToken(s, data, reminding)
	}
	if err != nil {
		return nil, err
	}

	return data, nil
}

func getReverseDelim(char int32) int32 {
	var result int32 = 0
	switch char {
	case '{':
		result = '}'
	case '}':
		result = '{'
	case '[':
		result = ']'
	case ']':
		result = '['
	}

	return result
}

func (p *JSONParser) parseAny(s string) (any, string, error) {
	if len(s) == 0 {
		return nil, "", nil
	}

	parser, exists := p.parsers[rune(s[0])]
	if !exists {
		return nil, s, ErrUnexpectedToken
	}

	return parser(s)
}

func (p *JSONParser) parseSpace(s string) (any, string, error) {
	return p.parseAny(strings.TrimSpace(s))
}

func (p *JSONParser) parseArray(s string) (any, string, error) {
	s = s[1:]
	var acc []any
	s = strings.TrimSpace(s)
	var err error

	for len(s) > 0 {
		if s[0] == ']' {
			s = s[1:]
			break
		}

		var remaining string
		var res any
		res, remaining, err = p.parseAny(s)
		if err != nil {
			if errors.Is(err, ErrIncompleteString) {
				err = nil
			}

			s = strings.TrimSpace(remaining)
			break
		}

		acc = append(acc, res)
		s = strings.TrimSpace(remaining)
		if strings.HasPrefix(s, ",") {
			s = strings.TrimSpace(s[1:])
		}
	}

	if len(acc) > 0 {
		if val, ok := acc[len(acc)-1].(map[string]any); ok && len(val) == 0 {
			acc = acc[:len(acc)-1]
		}
	}

	if len(acc) == 0 {
		return nil, s, err
	}

	return acc, s, err
}

func (p *JSONParser) parseObject(s string) (any, string, error) {
	s = s[1:]
	acc := make(map[string]any)
	s = strings.TrimSpace(s)
	var err error

	for len(s) > 0 {
		if s[0] == '}' {
			s = s[1:]
			break
		}

		if !p.strict && !p.containCompleteKey(s) {
			break
		}

		var key any
		var remaining string
		key, remaining, err = p.parseAny(s)
		if err != nil {
			if errors.Is(err, ErrIncompleteString) {
				err = nil
			}

			s = strings.TrimSpace(remaining)
			break
		}
		keyStr, ok := key.(string)
		if !ok {
			s = strings.TrimSpace(remaining)
			err = ErrUnexpectedToken
			break
		}

		s = strings.TrimSpace(remaining)
		if len(s) == 0 || s[0] == '}' {
			acc[keyStr] = nil
			break
		}
		if s[0] != ':' {
			err = ErrUnexpectedToken
			break
		}
		s = strings.TrimSpace(s[1:]) // skip ':'
		if len(s) == 0 || s[0] == '}' {
			acc[keyStr] = nil
			break
		}

		var value any
		value, remaining, err = p.parseAny(s)
		if err != nil {
			if errors.Is(err, ErrIncompleteString) {
				acc[keyStr] = nil
				err = nil
			}

			s = strings.TrimSpace(remaining)
			break
		}

		acc[keyStr] = value
		s = strings.TrimSpace(remaining)
		if strings.HasPrefix(s, ",") {
			s = strings.TrimSpace(s[1:])
		}
	}

	return acc, s, err
}

func (p *JSONParser) containCompleteKey(s string) bool {
	s = strings.TrimSpace(s)

	end := strings.Index(s[1:], "\"") + 1
	for end > 0 && s[end-1] == '\\' {
		if nextEnd := strings.Index(s[end+1:], "\""); nextEnd >= 0 {
			end = nextEnd + end + 1
		} else {
			return false
		}
	}

	if end == 0 {
		return false
	}

	return true
}

func (p *JSONParser) parseString(s string) (any, string, error) {
	end := strings.Index(s[1:], "\"") + 1
	for end > 0 && s[end-1] == '\\' {
		if nextEnd := strings.Index(s[end+1:], "\""); nextEnd >= 0 {
			end = nextEnd + end + 1
		} else {
			if !p.strict {
				return s[1:], "", nil
			}
			return nil, "", ErrIncompleteString
		}
	}

	if end == 0 {
		if !p.strict {
			return s[1:], "", nil
		}
		return nil, "", ErrIncompleteString
	}
	strVal := s[:end+1]
	s = s[end+1:]

	var result string
	err := json.Unmarshal([]byte(strVal), &result)
	return result, s, err
}

func (p *JSONParser) parseNumber(s string) (any, string, error) {
	i := 0
	if i < len(s) && s[i] == '-' {
		i++
	}

	hasDigits := false
	for i < len(s) && unicode.IsDigit(rune(s[i])) {
		hasDigits = true
		i++
	}

	if i < len(s) && s[i] == '.' {
		i++
		for i < len(s) && unicode.IsDigit(rune(s[i])) {
			hasDigits = true
			i++
		}
	}

	if !hasDigits {
		return nil, s, ErrIncompleteNum
	}

	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i < len(s) && (s[i] == '-' || s[i] == '+') {
			i++
		}
		hasExponent := false
		for i < len(s) && unicode.IsDigit(rune(s[i])) {
			hasExponent = true
			i++
		}
		if !hasExponent {
			return nil, s, ErrIncompleteNum
		}
	}

	numStr := s[:i]
	remaining := s[i:]

	num, err := strconv.ParseFloat(numStr, 64)
	if err != nil {
		return nil, s, ErrIncompleteNum
	}

	return num, remaining, nil
}

func (p *JSONParser) parseTrue(s string) (any, string, error) {
	if strings.HasPrefix(s, "true") {
		return true, s[4:], nil
	}
	return nil, s, ErrUnexpectedToken
}

func (p *JSONParser) parseFalse(s string) (any, string, error) {
	if strings.HasPrefix(s, "false") {
		return false, s[5:], nil
	}
	return nil, s, ErrUnexpectedToken
}

func (p *JSONParser) parseNull(s string) (any, string, error) {
	if strings.HasPrefix(s, "null") {
		return nil, s[4:], nil
	}
	return nil, s, ErrUnexpectedToken
}

func (p *JSONParser) defaultOnExtraToken(text string, data any, remaining string) {
	fmt.Printf("Parsed JSON with extra tokens. text: %s, data: %v, remaining: %s\n", text, data, remaining)
}
