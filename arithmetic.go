package exprast

import (
	"fmt"
	"strconv"
)

type ArithmeticVarInterpreter func(key string) (int64, error)
type arithmeticOpInterpreter func(op string, left int64, right int64) (int64, error)

type IArithmeticExprAst interface {
	Build(exp string, skipSign bool) (*ExprAST, error)
	Interpret(node *ExprAST, interpretVar ArithmeticVarInterpreter) (int64, error)
}

type arithmeticExprAst struct {
	astContext
	interpretOp arithmeticOpInterpreter
}

func NewArithmeticExprAst(preced tokPrecedence, next nextToken, interpretOp arithmeticOpInterpreter) IArithmeticExprAst {
	return &arithmeticExprAst{astContext{preced, next}, interpretOp}
}

func DefaultArithmeticExprAst() IArithmeticExprAst {
	return &arithmeticExprAst{
		astContext{
			tokPrecedence{"+": 20, "-": 20, "*": 40, "/": 40, "%": 40},
			nextArithmeticToken,
		},
		interpretArithmetic}
}

func (this *arithmeticExprAst) Build(exp string, skipSign bool) (*ExprAST, error) {
	if len(exp) < 1 {
		return nil, fmt.Errorf("invalid expr: %v", exp)
	}
	return buildExprAST(this.preced, exp, skipSign, func(tt tokenType) bool {
		return tokenVariable == tt || tokenInteger == tt
	}, this.next)
}

func (this *arithmeticExprAst) Interpret(node *ExprAST, interpretVar ArithmeticVarInterpreter) (int64, error) {
	switch node.Type {
	case tokenOperator:
		if nil == this.interpretOp {
			return 0, fmt.Errorf("interpretOp is nil")
		}
		l, err := this.Interpret(node.Left, interpretVar)
		if nil != err {
			return 0, err
		}
		r, err := this.Interpret(node.Right, interpretVar)
		if nil != err {
			return 0, err
		}
		return this.interpretOp(node.Key, l, r)
	case tokenVariable:
		if nil == interpretVar {
			return 0, fmt.Errorf("interpretVar is nil")
		}
		return interpretVar(node.Key)
	case tokenInteger:
		v, err := strconv.ParseInt(node.Key, 10, 64)
		if nil != err {
			return 0, err
		}
		return v, nil
	default:
		return 0, fmt.Errorf("unknown token type: `%v`", node.Type)
	}
}

func nextArithmeticToken(skipSign bool, p iparser) *token {
	if p.eof() {
		return nil
	}
	p.skipWhitespace()
	ch, offset := p.current()
	switch ch {
	case
		'(',
		')',
		'+',
		'-',
		'*',
		'/',
		'^',
		'%':
		return p.decodeOp()
	case
		'0',
		'1',
		'2',
		'3',
		'4',
		'5',
		'6',
		'7',
		'8',
		'9':
		return p.decodeInteger()
	case '$':
		return p.decodeVar(skipSign)
	default:
		p.throw(offset)
		return nil
	}
}

func interpretArithmetic(op string, left int64, right int64) (int64, error) {
	switch op {
	case "+":
		return left + right, nil
	case "-":
		return left - right, nil
	case "*":
		return left * right, nil
	case "/":
		if right == 0 {
			return 0, fmt.Errorf("violation of arithmetic specification: a division by zero in InterpretExprAST: [%v/%v]", left, right)
		}
		return left / right, nil
	case "%":
		return left % right, nil
	default:
		return 0, fmt.Errorf("unknown op: `%v`", op)
	}
}
