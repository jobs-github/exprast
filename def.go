package exprast

import (
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

type token struct {
	tok    string
	offset int
	tt     tokenType
}

type iparser interface {
	parse(skipSign bool) ([]*token, error)
	eof() bool
	skipWhitespace()
	throw(start int)
	current() (byte, int)

	decodeVar(skipSign bool) *token
	decodeOp() *token
	decodeInteger() *token
	decodeLogic() *token
}

type nextToken func(skipSign bool, p iparser) *token
