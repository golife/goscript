// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"os"
	"strconv"

	"github.com/golife/goscript/ast"
	"github.com/golife/goscript/scan"
	"github.com/golife/goscript/token"
)

// 入口
func ParseFile(filename, src string) *ast.File {
	var p parser
	p.init(filename, src)
	f := p.parseFile()
	if p.errors.Count() > 0 {
		p.errors.Print()
		return nil
	}
	return f
}

type parser struct {
	file    *token.File
	errors  scan.ErrorList
	scanner scan.Scanner

	// 当前token
	pos token.Pos
	tok token.Token
	lit string
}

func (p *parser) init(fname, src string) {
	p.file = token.NewFile(fname, src)
	p.scanner.Init(p.file, src)
}

func (p *parser) parseFile() *ast.File {
	var expr ast.Expr
	p.next()
	expr = p.parseExpression()
	if p.tok != token.EOF {
		p.addError("Expected EOF, got '" + p.lit + "'")
	}
	return &ast.File{Root: expr}
}

func (p *parser) addError(msg string) {
	p.errors.Add(p.file.Position(p.pos), msg)
	if p.errors.Count() >= 10 {
		p.errors.Print()
		os.Exit(1)
	}
}

// 期待当前token是某个token, 并返回pos, 并next()
func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.addError("Pos:" + strconv.Itoa(int(pos)) + " Expected '" + tok.String() + "' got '" + p.lit + "'")
	}
	p.next()
	return pos
}

// 下一个token
func (p *parser) next() {
	p.lit, p.tok, p.pos = p.scanner.Scan()
}

// 解析字面值
func (p *parser) parseBasicLit() *ast.BasicLit {
	return &ast.BasicLit{LitPos: p.pos, Kind: p.tok, Lit: p.lit}
}

/*
1+2+3
expression -> term { addOp term }
addOp -> "+" | "-"
term -> factor { mulop factor }
mulop -> "*" | "/"
factor -> NUM | "(" expression ")"
*/

func (p *parser) parseExpression() ast.Expr {
	// left
	left := p.parseTerm()

	// { addOp term }
	for {
		// op
		opTok := p.tok
		opPos := p.pos

		if opTok == token.ADD || opTok == token.SUB {
			p.next()
			// right
			right := p.parseTerm()

			// BinaryExpr再做为left
			// 1+2 + 3 => 1+2是一个left
			left = &ast.BinaryExpr{
				Op:    opTok,
				OpPos: opPos,
				X:     left,
				Y:     right,
			}
		} else {
			return left
		}
	}
}

/*
term -> factor { mulop factor }
mulop -> "*" | "/" | "%"
factor -> NUM | "(" expression ")"
*/
func (p *parser) parseTerm() ast.Expr {
	// left
	left := p.parseFactor()

	for {
		// op
		opTok := p.tok
		opPos := p.pos

		if opTok == token.MUL || opTok == token.QUO || opTok == token.REM {
			p.next()
			right := p.parseFactor()

			// BinaryExpr再做为left
			// 1*2 *3 3 => 1*2是一个left
			left = &ast.BinaryExpr{
				Op:    opTok,
				OpPos: opPos,
				X:     left,
				Y:     right,
			}
		} else {
			return left
		}
	}
}

// factor -> NUM | "(" expression ")"
// 会指向下一个
func (p *parser) parseFactor() ast.Expr {
	if p.tok.IsLiteral() { // NUM
		expr := p.parseBasicLit()
		p.next()
		return expr
	} else if p.tok == token.LPAREN { // (expression)
		p.next()
		expr := p.parseExpression()
		p.expect(token.RPAREN)
		return expr
	} else {
		p.addError("Expected NUM or ( but got '" + p.lit + "'")
		return nil
	}
}
