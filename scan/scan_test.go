// Copyright (c) 2014, Rob Thornton
// All rights reserved.
// This source code is governed by a Simplied BSD-License. Please see the
// LICENSE included in this distribution for a copy of the full license
// or, if one is not included, you may also find a copy at
// http://opensource.org/licenses/BSD-2-Clause

package scan_test

import (
	"testing"

	"github.com/golife/goscript/scan"
	"github.com/golife/goscript/token"
)

func test_handler(t *testing.T, src string, expected []token.Token) {
	var s scan.Scanner
	s.Init(token.NewFile("", src), src)
	lit, tok, pos := s.Scan()
	for i := 0; tok != token.EOF; i++ {
		if tok != expected[i] {
			t.Fatal(pos, "Expected:", expected[i], "Got:", tok, lit)
		}
		lit, tok, pos = s.Scan()
	}
}

type litTokPos struct {
	Lit string
	Tok token.Token
	Pos token.Pos
}

func test_handler2(t *testing.T, src string, expected []litTokPos) {
	var s scan.Scanner
	s.Init(token.NewFile("", src), src)
	lit, tok, pos := s.Scan()
	for i := 0; tok != token.EOF; i++ {
		item := expected[i]
		if tok != item.Tok || lit != item.Lit || pos != item.Pos {
			t.Fatal(pos, "Expected:", expected[i], "Got:", lit, tok, pos)
		}
		lit, tok, pos = s.Scan()
	}
}


func TestNumber(t *testing.T) {
	src := "-9"
	expected := []token.Token{
		token.SUB,
		token.INT,
		token.EOF,
	}
	test_handler(t, src, expected)
}


func TestScan1(t *testing.T) {
	src := `abc+b+1`
	expected := []litTokPos{
		{"abc",token.IDENT, 1},
		{"+",token.ADD, 4},
		{"b",token.IDENT, 5},
		{"+",token.ADD, 6},
		{"1",token.INT, 7},
		{"",token.EOF, 8},
	}
	test_handler2(t, src, expected)
}

func TestScan2(t *testing.T) {
	src := `var a = 1
cefA := a+2
if(a > cefA) {
	print("yes")
}
`
	expected := []litTokPos{
		{"var",token.VAR, 1},
		{"a",token.IDENT, 5},
		{"=",token.ASSIGN, 7},
		{"1",token.INT, 9},
		{"cefA",token.IDENT, 11},
		{":=",token.DEFINE, 16},
		{"a",token.IDENT, 19},
		{"+",token.ADD, 20},
		{"2",token.INT, 21},
		{"if",token.IF, 23},
		{"(",token.LPAREN, 25},
		{"a",token.IDENT, 26},
		{">",token.GTR, 28},
		{"cefA",token.IDENT, 30},
		{")",token.RPAREN, 34},
		{"{",token.LBRACE, 36},
		{"print",token.IDENT, 39},
		{"(",token.LPAREN, 44},
		{"\"yes\"",token.STRING, 45},
		{")",token.RPAREN, 50},
		{"}",token.RBRACE, 52},
		{"",token.EOF, 54},
	}
	test_handler2(t, src, expected)
}

func TestScan(t *testing.T) {
	src := "(+ 2 (- 4 1) (* 6 5) (% 10 2) (/ 9 3)); comment"
	expected := []token.Token{
		token.LPAREN,
		token.ADD,
		token.INT,
		token.LPAREN,
		token.SUB,
		token.INT,
		token.INT,
		token.RPAREN,
		token.LPAREN,
		token.MUL,
		token.INT,
		token.INT,
		token.RPAREN,
		token.LPAREN,
		token.REM,
		token.INT,
		token.INT,
		token.RPAREN,
		token.LPAREN,
		token.QUO,
		token.INT,
		token.INT,
		token.RPAREN,
		token.RPAREN,
		token.EOF,
	}
	test_handler(t, src, expected)
}

func TestScanAllTokens(t *testing.T) {
	src := "()+-*/% 1 12\t 12345 123456789 | a as ! \\ \r ;"
	expected := []token.Token{
		token.LPAREN,
		token.RPAREN,
		token.ADD,
		token.SUB,
		token.MUL,
		token.QUO,
		token.REM,
		token.INT,
		token.INT,
		token.INT,
		token.INT,
		token.ILLEGAL,
		token.ILLEGAL,
		token.ILLEGAL,
		token.ILLEGAL,
		token.ILLEGAL,
		token.ILLEGAL,
		token.EOF,
	}
	test_handler(t, src, expected)
}
