// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"os"

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

	// 是否存在unget
	hasUnget bool
	oldPos token.Pos
	oldTok token.Token
	oldLit string
}

func (p *parser) init(fname, src string) {
	p.file = token.NewFile(fname, src)
	p.scanner.Init(p.file, src)
}

func (p *parser) parseFile() *ast.File {
	var expr ast.Expr
	expr = p.parseExpresson()
	p.getToken()
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

// 期待当前token是某个token, 并返回pos
func (p *parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.addError("Expected '" + tok.String() + "' got '" + p.lit + "'")
	}
	return pos
}

// 下一个token
func (p *parser) getToken() {
	if p.hasUnget {
		p.hasUnget = false
		p.lit, p.tok, p.pos = p.oldLit, p.oldTok, p.oldPos
		return
	}

	p.oldLit, p.oldTok, p.oldPos = p.lit, p.tok, p.pos
	p.lit, p.tok, p.pos = p.scanner.Scan()
}

// 重回
func (p *parser) ungetToken() {
	p.hasUnget = true
	t1, t2, t3 := p.lit, p.tok, p.pos
	p.lit, p.tok, p.pos = p.oldLit, p.oldTok, p.oldPos
	p.oldLit, p.oldTok, p.oldPos = t1, t2, t3
}

// 解析字面值
func (p *parser) parseBasicLit() *ast.BasicLit {
	return &ast.BasicLit{LitPos: p.pos, Kind: p.tok, Lit: p.lit}
}

/*
1+2+3
expression -> term { addOp term }*
addOp -> "+" | "-"
term -> factor { mulop factor }*
mulop -> "*" | "/"
factor -> NUM | "(" expression ")"
 */

func (p *parser) parseExpresson() ast.Expr {
	// 为了得到pos
	p.getToken()
	open := p.pos
	p.ungetToken()

	// left
	left := p.parseTerm()

	// { addOp term }*
	for ;; {

		// op
		p.getToken()
		opTok := p.tok
		opPos := p.pos

		if opTok == token.ADD || opTok == token.SUB {
			// right
			right := p.parseTerm()
			end := p.pos

			// BinaryExpr再做为left
			// 1+2 + 3 => 1+2是一个left
			left = &ast.BinaryExpr{
				Expression: ast.Expression{
					Opening: open,
					Closing: end,
				},
				Op:    opTok,
				OpPos: opPos,
				List:  []ast.Expr{left, right},
			}
		} else {
			p.ungetToken()
			return left
		}
	}
}

/*
term -> factor { mulop factor }*
mulop -> "*" | "/" | "%"
factor -> NUM | "(" expression ")"
 */
func (p *parser) parseTerm() ast.Expr {
	// 取一个, 是为了得到pos
	p.getToken()
	open := p.pos
	p.ungetToken()

	// left
	left := p.parseFactor()

	for ;; {
		// op
		p.getToken()
		opTok := p.tok
		opPos := p.pos

		if opTok == token.MUL || opTok == token.QUO || opTok == token.REM {
			right := p.parseFactor()
			end := p.pos

			// BinaryExpr再做为left
			// 1*2 *3 3 => 1*2是一个left
			left = &ast.BinaryExpr{
				Expression: ast.Expression{
					Opening: open,
					Closing: end,
				},
				Op:    opTok,
				OpPos: opPos,
				List:  []ast.Expr{left, right},
			}
		} else {
			p.ungetToken()
			return left
		}
	}
}

// factor -> NUM | "(" expression ")"
// 会指向下一个
func (p *parser) parseFactor() ast.Expr {
	p.getToken()

	if p.tok.IsLiteral() { // NUM
		expr := p.parseBasicLit()
		return expr
	} else if p.tok == token.LPAREN { // (expression)
		expr := p.parseExpresson()

		p.getToken() // )
		p.expect(token.RPAREN)
		if p.tok != token.RPAREN {
			// p.addError("Expected ) but got '" + p.lit + "'")
			return nil
		}
		return expr
	} else {
		p.addError("Expected NUM or ( but got '" + p.lit + "'")
		return nil
	}
}
