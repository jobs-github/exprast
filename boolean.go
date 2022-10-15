package exprast

import (
	"fmt"
)

type BooleanVarInterpreter func(key string) (error, error)
type booleanOpInterpreter func(op string, left error, right error) error

type IBooleanExprAst interface {
	Build(exp string, skipSign bool) (*ExprAST, error)
	Interpret(node *ExprAST, interpretVar BooleanVarInterpreter) (error, error)
}

type booleanExprAst struct {
	astContext
}

func NewBooleanExprAst(preced tokPrecedence, next nextToken) IBooleanExprAst {
	return &booleanExprAst{astContext{preced, next}}
}

func DefaultBooleanExprAst() IBooleanExprAst {
	return &booleanExprAst{
		astContext{
			tokPrecedence{"||": 20, "&&": 40},
			nextBooleanToken,
		},
	}
}

func (this *booleanExprAst) Build(expr string, skipSign bool) (*ExprAST, error) {
	if len(expr) < 2 {
		return nil, fmt.Errorf("invalid expr: %v", expr)
	}
	if expr[0] != '(' {
		return nil, fmt.Errorf("expr not start with `(`")
	}
	if expr[len(expr)-1] != ')' {
		return nil, fmt.Errorf("expr not end with `)`")
	}
	return buildExprAST(defaultParser(expr, this.next), this.preced, expr, skipSign, func(tt tokenType) bool {
		return tokenVariable == tt
	})
}

func (this *booleanExprAst) doAnd(lhs *ExprAST, rhs *ExprAST, interpretVar BooleanVarInterpreter) (error, error) {
	l, err := this.Interpret(lhs, interpretVar)
	if nil != err {
		return nil, err
	}
	if nil != l {
		return l, nil
	}
	return this.Interpret(rhs, interpretVar)
}

func (this *booleanExprAst) doOr(lhs *ExprAST, rhs *ExprAST, interpretVar BooleanVarInterpreter) (error, error) {
	l, err := this.Interpret(lhs, interpretVar)
	if nil != err {
		return nil, err
	}
	if nil == l {
		return l, nil
	}
	return this.Interpret(rhs, interpretVar)
}

func (this *booleanExprAst) Interpret(node *ExprAST, interpretVar BooleanVarInterpreter) (error, error) {
	switch node.Type {
	case tokenOperator:
		if "&&" == node.Key {
			return this.doAnd(node.Left, node.Right, interpretVar)
		} else if "||" == node.Key {
			return this.doOr(node.Left, node.Right, interpretVar)
		} else {
			return nil, fmt.Errorf("undefined op: %v", node.Key)
		}
	case tokenVariable:
		if nil == interpretVar {
			return nil, fmt.Errorf("interpretVar is nil")
		}
		return interpretVar(node.Key)
	default:
		return nil, fmt.Errorf("unknown token type: `%v`", node.Type)
	}
}

func nextBooleanToken(skipSign bool, p iparser) *token {
	if p.eof() {
		return nil
	}
	p.skipWhitespace()
	ch, offset := p.current()
	switch ch {
	case '(', ')':
		return p.decodeOp()
	case '&', '|':
		return p.decodeLogic()
	case '$':
		return p.decodeVar(skipSign)
	default:
		p.throw(offset)
		return nil
	}
}
