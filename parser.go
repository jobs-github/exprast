package exprast

import (
	"errors"
	"fmt"
)

type parser struct {
	next   nextToken
	source string
	ch     byte
	offset int
	err    error
}

func defaultParser(s string, next nextToken) iparser {
	return &parser{
		next:   next,
		source: s,
		err:    nil,
		ch:     s[0],
	}
}

func (this *parser) parse(skipSign bool) ([]*token, error) {
	toks := []*token{}
	for {
		tok := this.next(skipSign, this)
		if nil == tok {
			break
		}
		toks = append(toks, tok)
	}
	return toks, this.err
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

func (this *parser) eof() bool {
	if this.offset >= len(this.source) || this.err != nil {
		return true
	}
	return false
}

func (this *parser) skipWhitespace() {
	var err error
	for isWhitespace(this.ch) && err == nil {
		err = this.nextCh()
	}
}

func (this *parser) current() (byte, int) {
	return this.ch, this.offset
}

func (this *parser) decodeOp() *token {
	start := this.offset
	tok := &token{
		tok:    string(this.ch),
		offset: start,
		tt:     tokenOperator,
	}
	this.nextCh()
	return tok
}

func (this *parser) decodeInteger() *token {
	start := this.offset
	for isDigitNum(this.ch) && this.nextCh() == nil {
	}
	tok := &token{
		tok: this.source[start:this.offset],
		tt:  tokenInteger,
	}
	tok.offset = start
	return tok
}

func (this *parser) decodeLogic() *token {
	start := this.offset
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

func (this *parser) decodeVar(skipSign bool) *token {
	start := this.offset
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
