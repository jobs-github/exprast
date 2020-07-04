package exprast

import (
	"errors"
	"fmt"
)

// reference: https://releases.llvm.org/3.1/docs/tutorial/LangImpl2.html

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

type exprAST struct {
	leaf      func(tt tokenType) bool
	preced    tokPrecedence
	tokens    []*token
	source    string
	currTok   *token
	currIndex int
	err       error
}

func (this *exprAST) parseExpression() *ExprAST {
	left := this.parsePrimary()
	return this.parseBinOpRHS(0, left)
}

func (this *exprAST) mismatch() {
	this.err = errors.New(fmt.Sprintf("want ')' but get %v\n%v",
		this.currTok.tok, errPos(this.source, this.currTok.offset)))
}

func isOperator(tt tokenType) bool {
	return tokenOperator == tt
}

func (this *exprAST) parsePrimary() *ExprAST {
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

func (this *exprAST) parseBinOpRHS(execPrec int, left *ExprAST) *ExprAST {
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

func (this *exprAST) getTokPrecedence() int {
	return this.preced.get(this.currTok.tok)
}

func (this *exprAST) parseLeaf(tt tokenType) *ExprAST {
	node := newExprAST(tt, this.currTok.tok, nil, nil)
	this.getNextToken()
	return node
}

func (this *exprAST) getNextToken() *token {
	this.currIndex++
	if this.currIndex < len(this.tokens) {
		this.currTok = this.tokens[this.currIndex]
		return this.currTok
	}
	return nil
}

func buildExprAST(
	p iparser,
	preced tokPrecedence,
	expr string,
	skipSign bool,
	leaf func(tt tokenType) bool,
) (*ExprAST, error) {
	toks, err := p.parse(skipSign)
	if nil != err {
		return nil, err
	}
	if toks == nil || len(toks) == 0 {
		return nil, errors.New("empty token")
	}
	// []token -> AST Tree
	rawAst := &exprAST{
		leaf:      leaf,
		preced:    preced,
		tokens:    toks,
		source:    expr,
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
