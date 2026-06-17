package filter

import "fmt"

type Expr interface {
	exprNode()
}

type AndExpr struct {
	Left, Right Expr
}

type OrExpr struct {
	Left, Right Expr
}

type NotExpr struct {
	Inner Expr
}

type CompareExpr struct {
	Field string
	Op    string
	Value string
}

type InExpr struct {
	Field  string
	Values []string
}

func (*AndExpr) exprNode()     {}
func (*OrExpr) exprNode()      {}
func (*NotExpr) exprNode()     {}
func (*CompareExpr) exprNode() {}
func (*InExpr) exprNode()      {}

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens}
}

func (p *Parser) peek() Token {
	if p.pos >= len(p.tokens) {
		return Token{Kind: TokenEOF}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.peek()
	if p.pos < len(p.tokens) {
		p.pos++
	}
	return tok
}

func (p *Parser) parseAndExpr() (Expr, error) {
	left, err := p.parseAtom()
	if err != nil {
		return nil, err
	}

	for p.peek().Kind == TokenAND {
		p.advance()
		right, err := p.parseAtom()
		if err != nil {
			return nil, err
		}
		left = &AndExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) parseOrExpr() (Expr, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.peek().Kind == TokenOR {
		p.advance()
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &OrExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *Parser) ParseExpr() (Expr, error) {
	expr, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}
	if p.peek().Kind != TokenEOF {
		return nil, fmt.Errorf("unexpected trailing token %q", p.peek().Literal)
	}
	return expr, nil
}

func (p *Parser) parseComparison() (Expr, error) {
	fieldTok := p.advance()
	field := fieldTok.Literal

	opTok := p.peek()
	if opTok.Kind != TokenOp {
		return nil, fmt.Errorf("expected operator after field %q, got %q", field, opTok.Literal)
	}
	p.advance()

	if opTok.Literal == "in" {
		listTok := p.peek()
		if listTok.Kind != TokenList {
			return nil, fmt.Errorf("expected list after 'in', got %q", listTok.Literal)
		}
		p.advance()
		return &InExpr{Field: field, Values: listTok.Values}, nil
	}

	valTok := p.peek()
	if valTok.Kind != TokenValue {
		return nil, fmt.Errorf("expected value after operator %q, got %q", opTok.Literal, valTok.Literal)
	}
	p.advance()

	return &CompareExpr{Field: field, Op: opTok.Literal, Value: valTok.Literal}, nil
}

func (p *Parser) parseAtom() (Expr, error) {
	switch p.peek().Kind {

	case TokenNOT:
		p.advance()
		inner, err := p.parseAtom()
		if err != nil {
			return nil, err
		}
		return &NotExpr{Inner: inner}, nil

	case TokenLParen:
		p.advance()
		inner, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		if p.peek().Kind != TokenRParen {
			return nil, fmt.Errorf("expected ')', got %q", p.peek().Literal)
		}
		p.advance()
		return inner, nil

	case TokenField:
		return p.parseComparison()

	default:
		return nil, fmt.Errorf("unexpected token %q", p.peek().Literal)
	}
}

func Parse(input string) (Expr, error) {
	l := &Lexer{input: []rune(input)}

	var tokens []Token
	for {
		tok := l.Next()
		if tok.Kind == TokenError {
			return nil, fmt.Errorf("lex error: %s", tok.Literal)
		}
		tokens = append(tokens, tok)
		if tok.Kind == TokenEOF {
			break
		}
	}

	p := NewParser(tokens)
	expr, err := p.ParseExpr()
	if err != nil {
		return nil, fmt.Errorf("parse error: %w", err)
	}
	return expr, nil
}
