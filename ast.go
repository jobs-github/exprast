package exprast

import (
	"errors"
	"fmt"
	"strings"
)

// reference: https://releases.llvm.org/3.1/docs/tutorial/LangImpl2.html

type astContext struct {
	preced tokPrecedence
	next   nextToken
}

type tokPrecedence map[string]int

func (this *tokPrecedence) get(tok string) int {
	if p, ok := (*this)[tok]; ok {
		return p
	}
	return -1
}

type tokenType = uint

const (
	tokenOperator tokenType = 1
	tokenVariable tokenType = 2
	tokenInteger  tokenType = 3
)

func errPos(s string, pos int) string {
	r := strings.Repeat("-", len(s)) + "\n"
	s += "\n"
	for i := 0; i < pos; i++ {
		s += " "
	}
	s += "^\n"
	return r + s + r
}

func isWhitespace(c byte) bool {
	return c == ' ' ||
		c == '\t' ||
		c == '\n' ||
		c == '\v' ||
		c == '\f' ||
		c == '\r'
}

func isLiteral(c byte) bool {
	return ('a' <= c && c <= 'z') ||
		('A' <= c && c <= 'Z') ||
		'0' <= c && c <= '9' ||
		('_' == c)
}

func isDigitNum(c byte) bool {
	return '0' <= c && c <= '9'
}

type nextToken func(skipSign bool, p *parser) *token

func parse(s string, skipSign bool, next nextToken) ([]*token, error) {
	p := &parser{
		source: s,
		err:    nil,
		ch:     s[0],
	}
	toks := p.parse(skipSign, next)
	if p.err != nil {
		return nil, p.err
	}
	return toks, nil
}

type token struct {
	tok    string
	offset int
	tt     tokenType
}

type parser struct {
	source string
	ch     byte
	offset int
	err    error
}

func (this *parser) parse(skipSign bool, next nextToken) []*token {
	toks := make([]*token, 0)
	for {
		tok := next(skipSign, this)
		if nil == tok {
			break
		}
		toks = append(toks, tok)
	}
	return toks
}

func (this *parser) nextCh() error {
	this.offset++
	if this.offset < len(this.source) {
		this.ch = this.source[this.offset]
		return nil
	}
	return errors.New("EOF")
}

func (this *parser) throw(start int) {
	s := fmt.Sprintf("symbol error: unkown '%v', pos [%v:]\n%s",
		string(this.ch),
		start,
		errPos(this.source, start))
	this.err = errors.New(s)
}

func (this *parser) decodeOp(start int) *token {
	tok := &token{
		tok:    string(this.ch),
		offset: start,
		tt:     tokenOperator,
	}
	this.nextCh()
	return tok
}

func (this *parser) decodeInteger(start int) *token {
	for isDigitNum(this.ch) && this.nextCh() == nil {
	}
	tok := &token{
		tok: this.source[start:this.offset],
		tt:  tokenInteger,
	}
	tok.offset = start
	return tok
}

func (this *parser) decodeLogic(start int) *token {
	startCh := this.ch
	if err := this.nextCh(); nil != err {
		this.throw(start)
		return nil
	} else {
		if startCh == this.ch {
			tok := &token{
				tok: this.source[start:(this.offset + 1)],
				tt:  tokenOperator,
			}
			tok.offset = start
			err = this.nextCh()
			return tok
		} else {
			this.throw(start)
			return nil
		}
	}
}

func (this *parser) decodeVar(start int, skipSign bool) *token {
	if err := this.nextCh(); nil != err {
		this.throw(start)
		return nil
	} else {
		for isLiteral(this.ch) && this.nextCh() == nil {
		}
		begin := start
		if skipSign {
			begin = start + 1
		}
		tok := &token{
			tok: this.source[begin:this.offset],
			tt:  tokenVariable,
		}
		tok.offset = start
		return tok
	}
}

type ExprAST struct {
	Type  tokenType `json:"type"`
	Key   string    `json:"key"`
	Left  *ExprAST  `json:"left,omitempty"`
	Right *ExprAST  `json:"right,omitempty"`
}

func newExprAST(tt tokenType, key string, left *ExprAST, right *ExprAST) *ExprAST {
	return &ExprAST{
		Type:  tt,
		Key:   key,
		Left:  left,
		Right: right,
	}
}

type AST struct {
	leaf      func(tt tokenType) bool
	preced    tokPrecedence
	tokens    []*token
	source    string
	currTok   *token
	currIndex int
	err       error
}

func (this *AST) parseExpression() *ExprAST {
	left := this.parsePrimary()
	return this.parseBinOpRHS(0, left)
}

func (this *AST) mismatch() {
	this.err = errors.New(fmt.Sprintf("want ')' but get %v\n%v",
		this.currTok.tok, errPos(this.source, this.currTok.offset)))
}

func isOperator(tt tokenType) bool {
	return tokenOperator == tt
}

func (this *AST) parsePrimary() *ExprAST {
	if this.leaf(this.currTok.tt) {
		return this.parseLeaf(this.currTok.tt)
	}
	if isOperator(this.currTok.tt) {
		if this.currTok.tok == "(" {
			this.getNextToken()
			e := this.parseExpression() // recursive
			if e == nil {
				return nil
			}
			if this.currTok.tok != ")" {
				this.mismatch()
				return nil
			}
			this.getNextToken()
			return e
		} else {
			return this.parseLeaf(tokenOperator)
		}
	}
	return nil
}

func (this *AST) parseBinOpRHS(execPrec int, left *ExprAST) *ExprAST {
	for {
		tokPrec := this.getTokPrecedence()
		if tokPrec < execPrec {
			return left
		}
		tt := this.currTok.tt
		op := this.currTok.tok
		if nil == this.getNextToken() {
			return left
		}
		right := this.parsePrimary()
		if nil == right {
			return nil
		}
		nextPrec := this.getTokPrecedence()
		if tokPrec < nextPrec {
			right = this.parseBinOpRHS(tokPrec+1, right) // recursive
			if nil == right {
				return nil
			}
		}
		left = newExprAST(tt, op, left, right)
	}
}

func (this *AST) getTokPrecedence() int {
	return this.preced.get(this.currTok.tok)
}

func (this *AST) parseLeaf(tt tokenType) *ExprAST {
	node := newExprAST(tt, this.currTok.tok, nil, nil)
	this.getNextToken()
	return node
}

func (this *AST) getNextToken() *token {
	this.currIndex++
	if this.currIndex < len(this.tokens) {
		this.currTok = this.tokens[this.currIndex]
		return this.currTok
	}
	return nil
}

func buildExprAST(
	preced tokPrecedence,
	exp string,
	skipSign bool,
	leaf func(tt tokenType) bool,
	next nextToken,
) (*ExprAST, error) {
	toks, err := parse(exp, skipSign, next)
	if nil != err {
		return nil, err
	}
	if toks == nil || len(toks) == 0 {
		return nil, errors.New("empty token")
	}
	// []token -> AST Tree
	rawAst := &AST{
		leaf:      leaf,
		preced:    preced,
		tokens:    toks,
		source:    exp,
		currIndex: 0,
		currTok:   toks[0],
	}

	// AST builder
	exprAst := rawAst.parseExpression()
	if rawAst.err != nil {
		return nil, err
	}
	return exprAst, nil
}
