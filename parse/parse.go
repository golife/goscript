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

	topScope *ast.Scope // top-most scope;
	// Label scopes
	// (maintained by open/close LabelScope)
	labelScope  *ast.Scope     // label scope for current function
	targetStack [][]*ast.Ident // stack of unresolved labels

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
		//panic("----")
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

func (p *parser) openLabelScope() {
	p.labelScope = ast.NewScope(p.labelScope)
	p.targetStack = append(p.targetStack, nil)
}

func (p *parser) closeLabelScope() {
	// resolve labels
	n := len(p.targetStack) - 1
	scope := p.labelScope
	for _, ident := range p.targetStack[n] {
		ident.Obj = scope.Lookup(ident.Name)
		if ident.Obj == nil /*&& p.mode&DeclarationErrors != 0*/ {
			p.addError(fmt.Sprintf("label %s undefined", ident.Name))
		}
	}
	// pop label scope
	p.targetStack = p.targetStack[0:n]
	p.labelScope = p.labelScope.Outer
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
// a > 10
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

// funcName () () {}
func (p *parser) parseFuncStmt() *ast.FuncStmt {
	pos := p.expect(token.FUNC)
	scope := ast.NewScope(p.topScope) // function scope

	// func name
	ident := p.parseIdent()

	params, results := p.parseSignature(scope)

	var body *ast.BlockStmt
	if p.tok == token.LBRACE { // {
		body = p.parseBody(scope)
		//p.expectSemi()
	} else if p.tok == token.SEMICOLON { // ;
		p.next()
		if p.tok == token.LBRACE { // {
			// opening { of function declaration on next line
			p.addError("unexpected semicolon or newline before {")
			body = p.parseBody(scope)
			p.expectSemi()
		}
	} else {
		p.expectSemi()
	}

	return &ast.FuncStmt{
		Name: ident,
		Type: &ast.FuncType{
			Func:    pos,
			Params:  params,
			Results: results,
		},
		Body: body,
	}
	//if recv == nil {
	//	// Go spec: The scope of an identifier denoting a constant, type,
	//	// variable, or function (but not method) declared at top level
	//	// (outside any function) is the package block.
	//	//
	//	// init() functions cannot be referred to and there may
	//	// be more than one - don't put them in the pkgScope
	//	if ident.Name != "init" {
	//		//p.declare(decl, nil, p.pkgScope, ast.Fun, ident)
	//	}
	//}
}

func (p *parser) parseBody(scope *ast.Scope) *ast.BlockStmt {
	lbrace := p.expect(token.LBRACE)
	p.topScope = scope // open function scope
	p.openLabelScope()
	list := p.parseStmtList()
	p.closeLabelScope()
	p.closeScope()
	rbrace := p.expect(token.RBRACE)

	return &ast.BlockStmt{Lbrace: lbrace, List: list, Rbrace: rbrace}
}

//func a (a, b int, c, d float32) {
//}

// ()
// (a int)
// (a, b int)
// (a, b int, c, d float)
// (a int, b float)
func (p *parser) parseParameterList(scope *ast.Scope) (params []*ast.Field) {
	// 1st ParameterDecl
	// A list of identifiers looks like a list of type names.
	var list []ast.Expr
	for {
		list = append(list, p.tryType()) // a b
		if p.tok != token.COMMA { // , 当到int时就会break
			break
		}
		p.next()
		if p.tok == token.RPAREN { // )
			break
		}
	}

	// list = []
	// list = [name1]
	// list = [name1, name2] 无 type

	// analyze case
	// 当有type时
	if typ := p.tryType(); typ != nil {
		// IdentifierList Type
		idents := p.makeIdentList(list) // 转成 idents
		field := &ast.Field{Names: idents, Type: typ}
		params = append(params, field)
		// Go spec: The scope of an identifier denoting a function
		// parameter or result variable is the function body.
		//p.declare(field, nil, scope, ast.Var, idents...)
		//p.resolve(typ)
		if !p.atComma("parameter list", token.RPAREN) { // 下一个, 如果不是 , 则证明参数没了
			return
		}
		// 还有参数组
		p.next()
		for p.tok != token.RPAREN && p.tok != token.EOF {
			idents := p.parseIdentList()
			typ := p.tryType()
			field := &ast.Field{Names: idents, Type: typ}
			params = append(params, field)
			// Go spec: The scope of an identifier denoting a function
			// parameter or result variable is the function body.
			//p.declare(field, nil, scope, ast.Var, idents...)
			//p.resolve(typ)
			if !p.atComma("parameter list", token.RPAREN) {
				break
			}
			p.next()
		}
		return
	}

	// 或没有参数
	// 匿名, 只有 type,  (int, int)
	// Type { "," Type } (anonymous parameters)
	params = make([]*ast.Field, len(list))
	for i, typ := range list {
		//p.resolve(typ)
		params[i] = &ast.Field{Type: typ}
	}
	return
}

// Expr -> Ident
func (p *parser) makeIdentList(list []ast.Expr) []*ast.Ident {
	idents := make([]*ast.Ident, len(list))
	for i, x := range list {
		ident, isIdent := x.(*ast.Ident)
		if !isIdent {
			if _, isBad := x.(*ast.BadExpr); !isBad {
				// only report error if it's a new one
				p.errorExpected(x.Pos(), "identifier")
			}
			ident = &ast.Ident{NamePos: x.Pos(), Name: "_"}
		}
		idents[i] = ident
	}
	return idents
}

func (p *parser) parseParameters(scope *ast.Scope) *ast.FieldList {
	var params []*ast.Field
	lparen := p.expect(token.LPAREN)
	if p.tok != token.RPAREN {
		params = p.parseParameterList(scope)
	}
	rparen := p.expect(token.RPAREN)

	return &ast.FieldList{Opening: lparen, List: params, Closing: rparen}
}

func (p *parser) parseResult(scope *ast.Scope) *ast.FieldList {
	if p.tok == token.LPAREN {
		return p.parseParameters(scope)
	}

	typ := p.tryType()
	if typ != nil {
		list := make([]*ast.Field, 1)
		list[0] = &ast.Field{Type: typ}
		return &ast.FieldList{List: list}
	}

	return nil
}

// signature = parameters results
func (p *parser) parseSignature(scope *ast.Scope) (params, results *ast.FieldList) {
	params = p.parseParameters(scope)
	results = p.parseResult(scope)
	return
}

// if a:=false; a {} if 初始;条件{}
// 返回 init, condition
func (p *parser) parseIfHeader() (init ast.Stmt, cond ast.Expr) {
	if p.tok == token.LBRACE { // { 没有条件, 直接 block
		p.addError("missing condition in if statement")
		cond = &ast.BadExpr{From: p.pos, To: p.pos}
		return
	}

	// 初始化
	if p.tok != token.SEMICOLON { // ;
		// accept potential variable declaration but complain
		if p.tok == token.VAR {
			p.next()
			p.addError(fmt.Sprintf("var declaration not allowed in 'IF' initializer"))
		}
		init = p.parseSimpleStmt()
	}

	// 条件
	var condStmt ast.Stmt
	var semi struct {
		pos token.Pos
		lit string // ";" or "\n"; valid if pos.IsValid()
	}
	// 如果不是 { 那么, 肯定还有条件, 否则 init就是条件
	if p.tok != token.LBRACE { // {
		if p.tok == token.SEMICOLON {
			semi.pos = p.pos
			semi.lit = p.lit
			p.next()
		} else {
			p.expect(token.SEMICOLON)
		}
		if p.tok != token.LBRACE {
			condStmt = p.parseSimpleStmt()
		}
	} else {
		condStmt = init
		init = nil
	}

	if condStmt != nil {
		cond = p.makeExpr(condStmt, "boolean expression") // 条件 必须是 expr, 不能是语句之类的
	} else if semi.pos.IsValid() {
		if semi.lit == "\n" {
			p.addError("unexpected newline, expecting { after if clause")
		} else {
			p.addError("missing condition in if statement")
		}
	}

	// make sure we have a valid AST
	if cond == nil {
		cond = &ast.BadExpr{From: p.pos, To: p.pos}
	}

	return
}

func (p *parser) parseIfStmt() *ast.IfStmt {
	pos := p.expect(token.IF)
	p.openScope()
	defer p.closeScope()

	init, cond := p.parseIfHeader()
	body := p.parseBlockStmt()

	var else_ ast.Stmt
	if p.tok == token.ELSE {
		p.next()
		switch p.tok {
		case token.IF:
			else_ = p.parseIfStmt()
		case token.LBRACE: // {
			else_ = p.parseBlockStmt()
			//p.expectSemi()
		default:
			p.errorExpected(p.pos, "if statement or block")
			else_ = &ast.BadStmt{From: p.pos, To: p.pos}
		}
	} else {
		//p.expectSemi()
	}

	return &ast.IfStmt{If: pos, Init: init, Cond: cond, Body: body, Else: else_}
}

func (p *parser) parseBlockStmt() *ast.BlockStmt {
	lbrace := p.expect(token.LBRACE)
	p.openScope()
	list := p.parseStmtList()
	p.closeScope()
	rbrace := p.expect(token.RBRACE)

	return &ast.BlockStmt{Lbrace: lbrace, List: list, Rbrace: rbrace}
}

// ----------------------------------------------------------------------------
// Blocks

func (p *parser) parseStmtList() (list []ast.Stmt) {
	for p.tok != token.RBRACE && p.tok != token.EOF {
		list = append(list, p.parseStmt())
	}
	return
}

// 必须是ExprStmt, 不能是其它的
func (p *parser) makeExpr(s ast.Stmt, want string) ast.Expr {
	if s == nil {
		return nil
	}
	if es, isExpr := s.(*ast.ExprStmt); isExpr {
		return es.X
		//return p.checkExpr(es.X)
	}
	found := "simple statement"
	if _, isAss := s.(*ast.AssignStmt); isAss {
		found = "assignment"
	}
	p.addError(fmt.Sprintf("expected %s, found %s (missing parentheses around composite literal?)", want, found))
	return &ast.BadExpr{From: s.Pos(), To: s.End()}
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

Expression = UnaryExpr | Expression binary_op Expression .
UnaryExpr  = PrimaryExpr | unary_op UnaryExpr .

binary_op  = "||" | "&&" | rel_op | add_op | mul_op .
rel_op     = "==" | "!=" | "<" | "<=" | ">" | ">=" .
add_op     = "+" | "-" | "|" | "^" .
mul_op     = "*" | "/" | "%" | "<<" | ">>" | "&" | "&^" .

unary_op   = "+" | "-" | "!" | "^" | "*" | "&" | "<-" .

*/
func (p *parser) parseExpression() ast.Expr {
	return p._parseExpression(0)
}

// lastOpPrecedence 上一个操作符的优级极
//
/*
1 + 2 * 3 =>
  +
1  *
  2 3

1 * 2 + 3 // 当到2时, 因为 * > + 所以直接返回2, 然后 1*2作为 left
    +
  *  3
1   2

 */
func (p *parser) _parseExpression(lastOpPrecedence int) ast.Expr {
	// left
	left := p.parseUnaryExpr()

	// { addOp term }
	for {
		// op
		opTok := p.tok
		opPos := p.pos

		curOpPrecendence := opTok.Precedence()

		// 如果上一个是 * 现在是 + 那么直接返回
		if lastOpPrecedence > curOpPrecendence {
			return left
		}

		switch opTok {
		case token.ADD, token.SUB, // + -
			token.MUL, token.QUO, token.REM, // * / %
			token.GTR, token.EQL, token.LSS, token.GEQ, token.LEQ, token.NEQ, // > < >= <= != ==
			token.LOR, token.LAND: // || &&
			p.next()
			// right
			right := p._parseExpression(curOpPrecendence)

			// BinaryExpr再做为left
			// 1+2 + 3 => 1+2是一个left
			left = &ast.BinaryExpr{
				Op:    opTok,
				OpPos: opPos,
				X:     left,
				Y:     right,
			}
		default:
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

func (p *parser) expectSemi() {
	// semicolon is optional before a closing ')' or '}'
	if p.tok != token.RPAREN && p.tok != token.RBRACE && p.tok != token.EOF {
		switch p.tok {
		case token.COMMA:
			// permit a ',' instead of a ';' but complain
			p.errorExpected(p.pos, "';'")
			fallthrough
		case token.SEMICOLON:
			p.next()
		default:
			p.errorExpected(p.pos, "';'")
			// p.advance(stmtStart)
		}
	}
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

