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
	interpretOp booleanOpInterpreter
}

func NewBooleanExprAst(preced tokPrecedence, next nextToken, interpretOp booleanOpInterpreter) IBooleanExprAst {
	return &booleanExprAst{astContext{preced, next}, interpretOp}
}

func DefaultBooleanExprAst() IBooleanExprAst {
	return &booleanExprAst{
		astContext{
			tokPrecedence{"||": 20, "&&": 40},
			nextBooleanToken,
		},
		interpretBoolean,
	}
}

func (this *booleanExprAst) Build(exp string, skipSign bool) (*ExprAST, error) {
	if len(exp) < 2 {
		return nil, fmt.Errorf("invalid expr: %v", exp)
	}
	if exp[0] != '(' {
		return nil, fmt.Errorf("expr not start with `(`")
	}
	if exp[len(exp)-1] != ')' {
		return nil, fmt.Errorf("expr not end with `)`")
	}
	return buildExprAST(this.preced, exp, skipSign, func(tt tokenType) bool {
		return tokenVariable == tt
	}, this.next)
}

func (this *booleanExprAst) Interpret(node *ExprAST, interpretVar BooleanVarInterpreter) (error, error) {
	switch node.Type {
	case tokenOperator:
		if nil == this.interpretOp {
			return nil, fmt.Errorf("interpretOp is nil")
		}
		l, err := this.Interpret(node.Left, interpretVar)
		if nil != err {
			return nil, err
		}
		r, err := this.Interpret(node.Right, interpretVar)
		if nil != err {
			return nil, err
		}
		return this.interpretOp(node.Key, l, r), nil
	case tokenVariable:
		if nil == interpretVar {
			return nil, fmt.Errorf("interpretVar is nil")
		}
		return interpretVar(node.Key)
	default:
		return nil, fmt.Errorf("unknown token type: `%v`", node.Type)
	}
}

func nextBooleanToken(skipSign bool, p *parser) *token {
	if p.offset >= len(p.source) || p.err != nil {
		return nil
	}
	var err error
	for isWhitespace(p.ch) && err == nil {
		err = p.nextCh()
	}
	switch p.ch {
	case '(', ')':
		return p.decodeOp(p.offset)
	case '&', '|':
		return p.decodeLogic(p.offset)
	case '$':
		return p.decodeVar(p.offset, skipSign)
	default:
		p.throw(p.offset)
		return nil
	}
}

func interpretBoolean(op string, left error, right error) error {
	if "&&" == op {
		if nil != left {
			return left
		}
		if nil != right {
			return right
		}
		return nil
	} else if "||" == op {
		if nil == left {
			return nil
		}
		if nil == right {
			return nil
		}
		return left
	} else {
		return fmt.Errorf("undefined op: %v", op)
	}
}
