// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package parse

import (
	"fmt"
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

	topScope *ast.Scope

	// 当前token
	pos token.Pos
	tok token.Token
	lit string
}

func (p *parser) init(fname, src string) {
	p.file = token.NewFile(fname, src)
	p.scanner.Init(p.file, src)
	p.openScope()
}


func assert(cond bool, msg string) {
	if !cond {
		panic("go/parser internal error: " + msg)
	}
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

func (p *parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if pos == p.pos {
		// the error happened at the current position;
		// make the error message more specific
		switch {
		case p.tok == token.SEMICOLON && p.lit == "\n":
			msg += ", found newline"
		case p.tok.IsLiteral():
			// print 123 rather than 'INT', etc.
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}
	p.addError(msg)
}

// 下一个token
func (p *parser) next() {
	p.lit, p.tok, p.pos = p.scanner.Scan()
}

// ----------------------------------------------------------------------------
// Scoping support

func (p *parser) openScope() {
	p.topScope = ast.NewScope(p.topScope)
}

func (p *parser) closeScope() {
	p.topScope = p.topScope.Outer
}

// File = { (Statement ";") | (FunctionDecl ";") }
// Statement = VarDecl | SimpleStmt | ReturnStmt | Block | IfStmt | ForStmt .
// SimpleStmt = EmptyStmt | ExpressionStmt | Assignment | ShortVarDecl .
// ReturnStmt = "return" [ ExpressionList ] .
// ExpressionStmt = Expression .
// EmptyStmt = . # The empty statement does nothing.
//
func (p *parser) parseFile() *ast.File {
	p.next()

	stmts := []ast.Stmt{}
	//funcs := []*ast.FuncDecl{}

	for {
		switch p.tok {
		//case token.FUNC:
		//	f := p.parseFunc()
		//	if f != nil {
		//		funcs = append(funcs, f)
		//	}
		case token.EOF:
			goto End
		default:
			fmt.Println("parseStmt:")
			stmt := p.parseStmt()
			if stmt != nil {
				stmts = append(stmts, stmt)
			}
		}
	}

End:
	//expr = p.parseExpression()
	//if p.tok != token.EOF {
	//	p.addError("Expected EOF, got '" + p.lit + "'")
	//}
	return &ast.File{Stmts: stmts}
}

// 解析语句
// var a int
// var a = 10
// a := 10
// expressionStmt a+10 (a+10)*b funcName()
func (p *parser) parseStmt () (s ast.Stmt) {
	tok := p.tok
	// lit := p.lit
	switch tok {
	case token.VAR: // var a, b = 1, 10
		s = p.parseVarStmt()
	case token.FUNC:
		return p.parseFuncStmt()
	case token.IF:
		s = p.parseIfStmt()
	case token.IDENT: // a = 10, a := 10, a, b = 10, 12, a + 10 有很多种可能
		s = p.parseSimpleStmt()
	case token.EOF:
		return nil
	default: // expression 10 + 2
		expr := p.parseExpression()
		return &ast.ExprStmt{X: expr}
	}
	return
}

// a = 10, a := 10, a, b = 10, 12, a + 10 有很多种可能
// 那怎么办? 先 parseLeft, left可能有多种情况
// a, // 有right
// a, b // 有right
// a + b // 无right
func (p *parser) parseSimpleStmt() ast.Stmt {
	x := p.parseLhsList()

	switch p.tok {
	case token.DEFINE, token.ASSIGN: // a := 10, a = 10
		pos, tok := p.pos, p.tok
		p.next()
		var y []ast.Expr // 右边的值
		y = p.parseRhsList()
		as := &ast.AssignStmt{Lhs: x, TokPos: pos, Tok: tok, Rhs: y}
		if tok == token.DEFINE {
			p.shortVarDecl(as, x)
		}
		return as
	default:
		return &ast.ExprStmt{X: x[0]}
	}
}

func (p *parser) shortVarDecl(decl *ast.AssignStmt, list []ast.Expr) {
	// Go spec: A short variable declaration may redeclare variables
	// provided they were originally declared in the same block with
	// the same type, and at least one of the non-blank variables is new.
	n := 0 // number of new variables
	for _, x := range list {
		if ident, isIdent := x.(*ast.Ident); isIdent {
			assert(ident.Obj == nil, "identifier already declared or resolved")
			obj := ast.NewObj(ast.Var, ident.Name)
			// remember corresponding assignment for other tools
			obj.Decl = decl
			ident.Obj = obj
			if ident.Name != "_" {
				if alt := p.topScope.Insert(obj); alt != nil {
					ident.Obj = alt // redeclaration
				} else {
					n++ // new declaration
				}
			}
		} else {
			p.errorExpected(x.Pos(), "identifier on left side of :=")
		}
	}

	if n == 0  {
		p.addError("no new variables on left side of :=")
	}
}

func (p *parser) parseLhsList() []ast.Expr {
	list := p.parseExprList()
	// 检查下 list 是否能成为 x TODO
	return list
}

func (p *parser) parseRhsList() []ast.Expr {
	list := p.parseExprList()
	// 检查下 list 是否能成为 y TODO
	return list
}

// var a
// var a, b
// var a = 1
// var a, b = 1, 2
// var a, b int = 1, 2
func (p *parser) parseVarStmt() *ast.VarStmt {
	pos := p.pos
	tok := p.tok
	p.next()
	names := p.parseIdentList() // a, b
	typ := p.tryType() // int or nil
	// 右边
	var values []ast.Expr
	// 如果下一个是 = 那么有 right
	if p.tok == token.ASSIGN {
		p.next()
		values = p.parseExprList()
	}

	return &ast.VarStmt{
		TokPos: pos,
		Tok: tok,
		Names: names,
		Type: typ,
		Values: values,
	}
}

func (p *parser) parseFuncStmt() *ast.FuncStmt {
	return nil
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	return nil
}

// If lhs is set, result list elements which are identifiers are not resolved.
func (p *parser) parseExprList() (list []ast.Expr) {
	list = append(list, p.parseExpression())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseExpression())
	}

	return
}

// Ident

func (p *parser) parseIdent() *ast.Ident {
	pos := p.pos
	name := "_"
	if p.tok == token.IDENT {
		name = p.lit
		p.next()
	} else {
		p.expect(token.IDENT) // use expect() error handling
	}
	return &ast.Ident{NamePos: pos, Name: name}
}

func (p *parser) parseIdentList() (list []*ast.Ident) {
	list = append(list, p.parseIdent())
	for p.tok == token.COMMA {
		p.next()
		list = append(list, p.parseIdent())
	}

	return
}

func (p *parser) tryType() *ast.Ident {
	switch p.tok {
	case token.IDENT: // int float
		typ := p.parseIdent()
		return typ
	}
	// no type found
	return nil
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

Expression = Term { addOp Term }
Term = UnaryExpr { mulop UnaryExpr }
UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
PrimaryExpr -> BasicLit | "(" Expression ")" | MethodExpr
BasicLit    = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .

mulop = "*" | "/" | "%"
addOp = "+" | "-"
unary_op   = "+" | "-"

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

Term = UnaryExpr { mulop UnaryExpr }
UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
PrimaryExpr -> BasicLit | "(" Expression ")" | MethodExpr
BasicLit    = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .

*/
func (p *parser) parseTerm() ast.Expr {
	// left
	left := p.parseUnaryExpr()

	for {
		// op
		opTok := p.tok
		opPos := p.pos

		if opTok == token.MUL || opTok == token.QUO || opTok == token.REM {
			p.next()
			right := p.parseUnaryExpr()

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

// UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
// unary_op   = "+" | "-"
func (p *parser) parseUnaryExpr() ast.Expr {
	// 看是否unary_op
	opTok := p.tok
	opPos := p.pos
	if opTok == token.ADD || opTok == token.SUB {
		p.next()
		x := p.parseUnaryExpr()
		return &ast.UnaryExpr{
			Op: opTok,
			OpPos: opPos,
			X: x,
		}
	} else {
		return p.parsePrimaryExpr()
	}
}

// factor -> NUM | "(" expression ")"
// UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .
// PrimaryExpr -> BasicLit | "(" Expression ")" | identifier | MethodExpr
// BasicLit    = int_lit | float_lit | imaginary_lit | rune_lit | string_lit .
// 会指向下一个
func (p *parser) parsePrimaryExpr() ast.Expr {
	if p.tok.IsLiteral() { // NUM
		expr := p.parseBasicLit()
		p.next()
		return expr
	} else if p.tok == token.LPAREN { // (expression)
		p.next()
		expr := p.parseExpression()
		p.expect(token.RPAREN)
		return expr
	} else if p.tok == token.IDENT { // a 或 a()
		ident := p.parseIdent()
		switch p.tok {
		case token.LPAREN:
			return p.parseCall(ident)
		default:
			return ident
		}
	} else {
		p.addError("Expected NUM or ( but got '" + p.lit + "'")
		return nil
	}
}

// a(b, c, d) 当前是 (
func (p *parser) parseCall(fun ast.Expr) *ast.CallExpr {
	lparen := p.expect(token.LPAREN)
	var list []ast.Expr
	for p.tok != token.RPAREN && p.tok != token.EOF {
		list = append(list, p.parseExpression())
		if !p.atComma("argument list", token.RPAREN) {
			break
		}
		p.next()
	}
	rparen := p.expectClosing(token.RPAREN, "argument list")

	return &ast.CallExpr{Fun: fun, Lparen: lparen, Args: list, Rparen: rparen}
}

func (p *parser) atComma(context string, follow token.Token) bool {
	if p.tok == token.COMMA {
		return true
	}
	if p.tok != follow {
		msg := "missing ','"
		if p.tok == token.SEMICOLON && p.lit == "\n" {
			msg += " before newline"
		}
		p.addError(msg+" in "+context)
		return true // "insert" comma and continue
	}
	return false
}


// expectClosing is like expect but provides a better error message
// for the common case of a missing comma before a newline.
//
func (p *parser) expectClosing(tok token.Token, context string) token.Pos {
	if p.tok != tok && p.tok == token.SEMICOLON && p.lit == "\n" {
		p.addError("missing ',' before newline in "+context)
		p.next()
	}
	return p.expect(tok)
}

