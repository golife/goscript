// Copyright (c) 2014, Rob Thornton
// All rights reserved.
// This source code is governed by a Simplied BSD-License. Please see the
// LICENSE included in this distribution for a copy of the full license
// or, if one is not included, you may also find a copy at
// http://opensource.org/licenses/BSD-2-Clause

package scan

import (
	"unicode"

	"github.com/golife/goscript/token"
)

type Scanner struct {
	ch       rune // current character
	offset   int  // current character offset
	rdOffset int  // reading offset (position after current character) = offset + 1
	src      string
	file     *token.File
}

func (s *Scanner) Init(file *token.File, src string) {
	s.file = file
	s.offset, s.rdOffset = 0, 0
	s.src = src
	s.ch = rune(0)

	// 读一个字符
	s.next()
}

func (s *Scanner) Scan() (lit string, tok token.Token, pos token.Pos) {
	s.skipWhitespace()

	if unicode.IsDigit(s.ch) {
		return s.scanNumber()
	}

	lit, pos = string(s.ch), s.file.Pos(s.offset)
	switch s.ch {
	case '(':
		tok = token.LPAREN
	case ')':
		tok = token.RPAREN
	case '{':
		tok = token.LBRACE
	case '}':
		tok = token.RBRACE
	case '[':
		tok = token.LBRACK
	case ']':
		tok = token.RBRACK
	case ',':
		tok = token.COMMA
	case '+':
		tok = token.ADD
	case '-':
		tok = token.SUB
	case '*':
		tok = token.MUL
	case '/':
		tok = token.QUO
	case '%':
		tok = token.REM
	case ';':
		tok = token.SEMICOLON
	case '=':
		s.next()
		switch s.ch {
		case '=': // ==
			tok = token.EQL
			lit = "=="
		default: // =
			tok = token.ASSIGN
			return
		}
	case '<':
		s.next()
		switch s.ch {
		case '=': // <=
			tok = token.LEQ
			lit = "<="
		default: // <
			tok = token.LSS
			return
		}
	case '>':
		s.next()
		switch s.ch {
		case '=': // >=
			tok = token.GEQ
			lit = ">="
		default: // >
			tok = token.GTR
			return
		}
	case '!':
		s.next()
		switch s.ch {
		case '=': // !=
			tok = token.NEQ
			lit = "!="
		default: // !
			tok = token.NOT
			return
		}
	case ':': // :=
		s.next()
		switch s.ch {
		case '=':
			tok = token.DEFINE
			lit = ":="
		default:
			tok = token.ILLEGAL
		}
	case '"': // "life" string
		tok = token.STRING
		start := s.offset
		for {
			s.next()
			if s.ch != '"' { // TODO EOF 考虑
			} else {
				lit = s.src[start:s.offset+1]
				break
			}
		}
	case rune(-1):
		tok = token.EOF
		return
	default: // 其它字符 IDENT a b ab ab123
		tok = token.IDENT
		start := s.offset
		for {
			s.next()
			if s.ch >= 'a' && s.ch <= 'z' || s.ch >= 'A' && s.ch <= 'Z' { // TODO EOF 考虑
				//fmt.Println(string(s.ch))
			} else {
				if s.ch == rune(-1) {
					lit = s.src[start:s.offset+1]
				} else {
					lit = s.src[start:s.offset]
				}
				tok = token.Lookup(lit) // 如果是 func var for 之类的, 就变成相应的关键词
				return
			}
		}
	}

	s.next()

	return
}

// 如果之前的位置是\n, 则加一行
// 读位置为rdOffset的一个字符, 读后 rdOffset++
func (s *Scanner) next() {
	if s.rdOffset < len(s.src) {
		s.offset = s.rdOffset
		// 上一个位置是\n且还有下一个位置
		if s.ch == '\n' {
			s.file.AddLine(s.rdOffset) // s.offset+1
		}

		s.ch = rune(s.src[s.offset])
		// 下一个offset
		s.rdOffset++
	} else {
		s.ch = rune(-1) // EOF
	}
}

func (s *Scanner) scanNumber() (string, token.Token, token.Pos) {
	start := s.offset

	for unicode.IsDigit(s.ch) {
		s.next()
	}
	offset := s.offset
	if s.ch == rune(-1) {
		offset++
	}
	return s.src[start:offset], token.INT, s.file.Pos(start)
}

func (s *Scanner) skipComment() {
	for s.ch != '\n' && s.offset < len(s.src)-1 {
		s.next()
	}
}

func (s *Scanner) skipWhitespace() {
	for s.ch == ' ' || s.ch == '\t' || s.ch == '\n' || s.ch == '\r' {
		s.next()
	}
}
