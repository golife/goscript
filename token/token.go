// Copyright (c) 2014, Rob Thornton
// All rights reserved.
// This source code is governed by a Simplied BSD-License. Please see the
// LICENSE included in this distribution for a copy of the full license
// or, if one is not included, you may also find a copy at
// http://opensource.org/licenses/BSD-2-Clause

package token

import "strconv"

type Token int

const (
	tok_start Token = iota

	EOF
	ILLEGAL
	COMMENT

	// 字面值
	lit_start

	INT    // 1
	FLOAT  // 1.2
	CHAR   // 'a'
	STRING // "a"

	lit_end

	// 操作
	op_start

	IDENT // funcname, var name

	ADD // +
	SUB // -
	MUL // *
	QUO // /
	REM // %

	LPAREN // (
	RPAREN // )
	LBRACK // [
	RBRACK // ]
	LBRACE // {
	RBRACE // }

	COMMA // ,
	SEMICOLON // ;

	EQL // ==
	LSS // <
	GTR // >
	NOT // !

	NEQ // !=
	LEQ // <=
	GEQ // >=

	ASSIGN // =
	DEFINE // :=

	op_end

	keyword_beg

	IF
	ELSE
	FOR
	FUNC
	RETURN
	VAR

	keyword_end

	tok_end
)

var tokens = [...]string{
	ILLEGAL: "ILLEGAL",

	EOF:     "EOF",
	COMMENT: "COMMENT",

	IDENT:  "IDENT",
	INT:    "INT",
	FLOAT:  "FLOAT",
	CHAR:   "CHAR",
	STRING: "STRING",

	ADD: "+",
	SUB: "-",
	MUL: "*",
	QUO: "/",
	REM: "%",

	EQL:    "==",
	LSS:    "<",
	GTR:    ">",
	ASSIGN: "=",
	NOT:    "!",

	NEQ:    "!=",
	LEQ:    "<=",
	GEQ:    ">=",
	DEFINE: ":=",

	LPAREN: "(",
	LBRACK: "[",
	LBRACE: "{",

	RPAREN: ")",
	RBRACK: "]",
	RBRACE: "}",

	COMMA: ",",
	SEMICOLON: ":",

	ELSE: "else",
	FOR:  "for",

	FUNC: "func",
	IF:   "if",

	RETURN: "return",

	VAR: "var",
}

var keywords map[string]Token

func init() {
	keywords = make(map[string]Token)
	for i := keyword_beg + 1; i < keyword_end; i++ {
		keywords[tokens[i]] = i
	}
}

func (t Token) IsEOF() bool {
	return t == EOF
}

func (t Token) IsLiteral() bool {
	return t > lit_start && t < lit_end
}

func (t Token) IsOperator() bool {
	return t > op_start && t < op_end
}

// Lookup maps an identifier to its keyword token or IDENT (if not a keyword).
//
func Lookup(ident string) Token {
	if tok, is_keyword := keywords[ident]; is_keyword {
		return tok
	}
	return IDENT
}

func (t Token) String() string {
	s := ""
	if 0 <= t && t < Token(len(tokens)) {
		s = tokens[t]
	}
	if s == "" {
		s = "token(" + strconv.Itoa(int(t)) + ")"
	}
	return s
}

func (t Token) Valid() bool {
	return t > tok_start && t < tok_end
}
