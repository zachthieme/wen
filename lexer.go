package wen

import (
	"strconv"
	"strings"
	"time"
)

var weekdays = map[string]time.Weekday{
	"sunday": time.Sunday, "monday": time.Monday, "tuesday": time.Tuesday,
	"wednesday": time.Wednesday, "thursday": time.Thursday, "friday": time.Friday,
	"saturday": time.Saturday,
	"sun": time.Sunday, "mon": time.Monday, "tue": time.Tuesday,
	"wed": time.Wednesday, "thu": time.Thursday, "fri": time.Friday,
	"sat": time.Saturday,
}

var months = map[string]time.Month{
	"january": time.January, "february": time.February, "march": time.March,
	"april": time.April, "may": time.May, "june": time.June,
	"july": time.July, "august": time.August, "september": time.September,
	"october": time.October, "november": time.November, "december": time.December,
	"jan": time.January, "feb": time.February, "mar": time.March,
	"apr": time.April, "jun": time.June, "jul": time.July,
	"aug": time.August, "sep": time.September, "oct": time.October,
	"nov": time.November, "dec": time.December,
}

var modifiers = map[string]bool{"this": true, "next": true, "last": true}

var prepositions = map[string]bool{
	"in": true, "from": true, "of": true, "at": true, "ago": true, "now": true,
}

var units = map[string]string{
	"day": "day", "days": "day",
	"week": "week", "weeks": "week",
	"month": "month", "months": "month",
	"quarter": "quarter", "quarters": "quarter",
	"year": "year", "years": "year",
	"hour": "hour", "hours": "hour",
	"minute": "minute", "minutes": "minute",
}

var relativeDays = map[string]bool{"today": true, "tomorrow": true, "yesterday": true}
var namedTimes = map[string]bool{"noon": true, "midnight": true}
var meridiems = map[string]bool{"am": true, "pm": true}
var boundaries = map[string]bool{"beginning": true, "end": true}
var noiseWords = map[string]bool{"the": true, "a": true}

var ordinalWords = map[string]int{
	"first": 1, "second": 2, "third": 3, "fourth": 4, "fifth": 5,
	"sixth": 6, "seventh": 7, "eighth": 8, "ninth": 9, "tenth": 10,
	"eleventh": 11, "twelfth": 12,
}

var cardinalWords = map[string]int{
	"one": 1, "two": 2, "three": 3, "four": 4, "five": 5,
	"six": 6, "seven": 7, "eight": 8, "nine": 9, "ten": 10,
	"eleven": 11, "twelve": 12, "twenty": 20, "thirty": 30,
}

type lexer struct {
	input  string
	lower  string
	pos    int
	tokens []token
}

func newLexer(input string) *lexer {
	return &lexer{
		input: input,
		lower: strings.ToLower(input),
	}
}

func (l *lexer) tokenize() []token {
	l.tokens = make([]token, 0, len(l.input)/4+1)
	for l.pos < len(l.lower) {
		l.skipWhitespace()
		if l.pos >= len(l.lower) {
			break
		}
		ch := l.lower[l.pos]
		switch {
		case ch == ':':
			l.tokens = append(l.tokens, token{Kind: tokenColon, Value: ":", Position: l.pos})
			l.pos++
		case ch >= '0' && ch <= '9':
			l.scanNumber()
		case isLetter(ch):
			l.scanWord()
		default:
			l.tokens = append(l.tokens, token{Kind: tokenUnknown, Value: string(ch), Position: l.pos})
			l.pos++
		}
	}
	l.tokens = append(l.tokens, token{Kind: tokenEOF, Position: l.pos})
	return l.tokens
}

func (l *lexer) skipWhitespace() {
	for l.pos < len(l.lower) {
		ch := l.lower[l.pos]
		if ch != ' ' && ch != '\t' && ch != '\r' && ch != '\n' {
			break
		}
		l.pos++
	}
}

func (l *lexer) scanNumber() {
	start := l.pos
	for l.pos < len(l.lower) && l.lower[l.pos] >= '0' && l.lower[l.pos] <= '9' {
		l.pos++
	}
	numStr := l.lower[start:l.pos]
	val := atoi(numStr)

	// Check for ordinal suffix: 1st, 2nd, 3rd, 4th, ...
	if l.pos+2 <= len(l.lower) {
		suffix := l.lower[l.pos : l.pos+2]
		if isOrdinalSuffix(suffix) && !l.followedByLetter(l.pos+2) {
			l.pos += 2
			l.tokens = append(l.tokens, token{Kind: tokenOrdinal, Value: numStr + suffix, IntVal: val, Position: start})
			return
		}
	}

	// Check for attached meridiem: 3pm, 11am
	if l.pos+2 <= len(l.lower) {
		suffix := l.lower[l.pos : l.pos+2]
		if meridiems[suffix] && !l.followedByLetter(l.pos+2) {
			l.tokens = append(l.tokens,
				token{Kind: tokenNumber, Value: numStr, IntVal: val, Position: start},
				token{Kind: tokenMeridiem, Value: suffix, Position: l.pos},
			)
			l.pos += 2
			return
		}
	}

	l.tokens = append(l.tokens, token{Kind: tokenNumber, Value: numStr, IntVal: val, Position: start})
}

func (l *lexer) scanWord() {
	start := l.pos
	for l.pos < len(l.lower) && isLetter(l.lower[l.pos]) {
		l.pos++
	}
	word := l.lower[start:l.pos]

	if w, ok := weekdays[word]; ok {
		l.tokens = append(l.tokens, token{Kind: tokenWeekday, Value: word, Weekday: w, Position: start})
		return
	}
	// Plural weekdays: "mondays" -> Monday
	if strings.HasSuffix(word, "s") {
		if w, ok := weekdays[word[:len(word)-1]]; ok {
			l.tokens = append(l.tokens, token{Kind: tokenWeekday, Value: word[:len(word)-1], Weekday: w, Position: start})
			return
		}
	}
	if m, ok := months[word]; ok {
		l.tokens = append(l.tokens, token{Kind: tokenMonth, Value: word, Month: m, Position: start})
		return
	}
	if modifiers[word] {
		l.tokens = append(l.tokens, token{Kind: tokenModifier, Value: word, Position: start})
		return
	}
	if prepositions[word] {
		l.tokens = append(l.tokens, token{Kind: tokenPreposition, Value: word, Position: start})
		return
	}
	if u, ok := units[word]; ok {
		l.tokens = append(l.tokens, token{Kind: tokenUnit, Value: u, Position: start})
		return
	}
	if relativeDays[word] {
		l.tokens = append(l.tokens, token{Kind: tokenRelativeDay, Value: word, Position: start})
		return
	}
	if namedTimes[word] {
		l.tokens = append(l.tokens, token{Kind: tokenNamedTime, Value: word, Position: start})
		return
	}
	if meridiems[word] {
		l.tokens = append(l.tokens, token{Kind: tokenMeridiem, Value: word, Position: start})
		return
	}
	if v, ok := ordinalWords[word]; ok {
		l.tokens = append(l.tokens, token{Kind: tokenOrdinal, Value: word, IntVal: v, Position: start})
		return
	}
	if v, ok := cardinalWords[word]; ok {
		l.tokens = append(l.tokens, token{Kind: tokenNumber, Value: word, IntVal: v, Position: start})
		return
	}
	if boundaries[word] {
		l.tokens = append(l.tokens, token{Kind: tokenBoundary, Value: word, Position: start})
		return
	}
	if word == "every" {
		l.tokens = append(l.tokens, token{Kind: tokenEvery, Value: word, Position: start})
		return
	}
	if noiseWords[word] {
		l.tokens = append(l.tokens, token{Kind: tokenNoise, Value: word, Position: start})
		return
	}
	l.tokens = append(l.tokens, token{Kind: tokenUnknown, Value: word, Position: start})
}

func (l *lexer) followedByLetter(pos int) bool {
	return pos < len(l.lower) && isLetter(l.lower[pos])
}

// isLetter reports whether ch is a letter byte. It includes non-ASCII bytes
// (0x80+) so that multi-byte UTF-8 characters are kept together as part of a
// single word token rather than splitting on non-ASCII boundaries.
func isLetter(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || ch > 127
}

func isOrdinalSuffix(s string) bool {
	return s == "st" || s == "nd" || s == "rd" || s == "th"
}

// atoi converts a digit-only string to int. Returns 0 on overflow, which is
// harmless since no valid date component is zero. Callers must guarantee s is
// all digits (e.g., from scanNumber).
func atoi(s string) int {
	val, err := strconv.Atoi(s)
	if err != nil {
		return 0 // overflow — digit string too large for int
	}
	return val
}
