package filter

import (
	"fmt"
	"strings"
	"unicode"
)

type TokenKind int

const (
	TokenField TokenKind = iota // type, source, confidence, tags, value
	TokenOp                     // =, !=, contains, in, >=, <=, >,
	TokenValue                  // "urlhaus", 75, [url, domain]
	TokenAND
	TokenOR
	TokenNOT
	TokenLParen
	TokenRParen
	TokenList
	TokenError
	TokenEOF
)

type Token struct {
	Kind    TokenKind
	Literal string // raw text of this token
	Values  []string
}

type Lexer struct {
	input []rune
	pos   int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: []rune(input)}
}

func (l *Lexer) skipWhitespace() {
	for {

		if l.pos >= len(l.input) {
			return
		}

		if !unicode.IsSpace(l.input[l.pos]) {
			return
		}

		l.pos++
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	return l.input[l.pos]
}

func (l *Lexer) advance() {
	if l.pos >= len(l.input) {
		return
	}

	l.pos++
}

func (l *Lexer) readWord() string {
	start := l.pos

	for l.pos < len(l.input) && isWordChar(l.peek()) {
		l.advance()
	}

	return string(l.input[start:l.pos])
}

func (l *Lexer) readList() Token {
	l.advance() // consume '['

	var values []string

	for {
		l.skipWhitespace()

		if l.pos >= len(l.input) {
			break
		}

		if l.peek() == ']' {
			l.advance() // consume ']'
			break
		}

		if l.peek() == ',' {
			l.advance() // consume ',' between items
			continue
		}

		word := l.readWord()
		if word != "" {
			values = append(values, word)
		}
	}

	return Token{Kind: TokenList, Literal: strings.Join(values, ","), Values: values}
}

func (l *Lexer) readQuotedString() Token {
	quote := l.peek()
	l.advance() // consume the opening quote

	start := l.pos

	for l.pos < len(l.input) && l.peek() != quote {
		l.advance()
	}

	content := string(l.input[start:l.pos])

	if l.pos >= len(l.input) {
		return Token{Kind: TokenError, Literal: fmt.Sprintf("unterminated string starting with %c", quote)}
	}

	l.advance() // consume the closing quote

	return Token{Kind: TokenValue, Literal: content}
}

func (l *Lexer) readOperator() Token {
	startRune := l.peek()
	l.advance()

	switch startRune {
	case '=':
		return Token{Kind: TokenOp, Literal: "="}

	case '!':
		if l.peek() == '=' {
			l.advance()
			return Token{Kind: TokenOp, Literal: "!="}
		}
		return Token{Kind: TokenError, Literal: "expected '=' after '!'"}
	case '>':
		if l.peek() == '=' {
			l.advance()
			return Token{Kind: TokenOp, Literal: ">="}
		}
		return Token{Kind: TokenOp, Literal: ">"}

	case '<':
		if l.peek() == '=' {
			l.advance()
			return Token{Kind: TokenOp, Literal: "<="}
		}
		return Token{Kind: TokenOp, Literal: "<"}
	}

	return Token{Kind: TokenError, Literal: fmt.Sprintf("unexpected operator character %q", startRune)}
}

func (l *Lexer) Next() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Kind: TokenEOF}
	}

	switch {
	case l.peek() == '(':
		l.advance()
		return Token{Kind: TokenLParen, Literal: "("}
	case l.peek() == ')':
		l.advance()
		return Token{Kind: TokenRParen, Literal: ")"}
	case l.peek() == '[':
		return l.readList()
	case l.peek() == '"' || l.peek() == '\'':
		return l.readQuotedString()
	case isOperatorStart(l.peek()):
		return l.readOperator()
	default:
		word := l.readWord()
		return l.classifyWord(word)
	}
}

func (l *Lexer) classifyWord(word string) Token {
	switch strings.ToUpper(word) {
	case "AND":
		return Token{Kind: TokenAND, Literal: word}
	case "OR":
		return Token{Kind: TokenOR, Literal: word}
	case "NOT":
		return Token{Kind: TokenNOT, Literal: word}
	case "CONTAINS":
		return Token{Kind: TokenOp, Literal: "contains"}
	case "IN":
		return Token{Kind: TokenOp, Literal: "in"}
	case "=", "!=", ">=", "<=", ">", "<":
		return Token{Kind: TokenOp, Literal: word}
	case "TYPE", "SOURCE", "VALUE", "TAGS", "CONFIDENCE", "ID":
		return Token{Kind: TokenField, Literal: strings.ToLower(word)}
	default:
		return Token{Kind: TokenValue, Literal: word}
	}
}

func isWordChar(r rune) bool {
	switch {
	case unicode.IsLetter(r):
		return true
	case unicode.IsDigit(r):
		return true
	case r == '_' || r == '.' || r == '-':
		return true
	default:
		return false
	}
}

func isOperatorStart(r rune) bool {
	return r == '!' || r == '<' || r == '>' || r == '='
}
