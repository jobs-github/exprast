package exprast

import (
	"errors"
	"fmt"
)

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
