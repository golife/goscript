// Copyright 2021 The Goscript Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ast

import (
	"github.com/golife/goscript/token"
)

// Node: Expr
type Node interface {
	Pos() token.Pos
	End() token.Pos
}

// Expr: BasicLit, BinaryExpr
// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// All statement nodes implement the Stmt interface.
type Stmt interface {
	Node
	stmtNode()
}

// All declaration nodes implement the Decl interface.
type Decl interface {
	Node
	declNode()
}

//-----------
// Node

// 字面值
type BasicLit struct {
	LitPos token.Pos
	Kind   token.Token // token.INT, token.FLOAT, token.IMAG, token.CHAR, or token.STRING
	Lit    string
}

// 二元操作 X op Y
type BinaryExpr struct {
	X     Expr
	OpPos token.Pos
	Op    token.Token
	Y     Expr
}

// 一元操作 +X -X
type UnaryExpr struct {
	OpPos token.Pos   // position of Op
	Op    token.Token // operator
	X     Expr        // operand
}

type BadExpr struct {
	From, To token.Pos // position range of bad expression
}

// An Ident node represents an identifier.
type Ident struct {
	NamePos token.Pos // identifier position
	Name    string    // identifier name
	Obj     *Object   // denoted object; or nil
}

// A CallExpr node represents an expression followed by an argument list.
type CallExpr struct {
	Fun      Expr      // function expression
	Lparen   token.Pos // position of "("
	Args     []Expr    // function arguments; or nil
	Rparen   token.Pos // position of ")"
}

// A FuncLit node represents a function literal.
type FuncLit struct {
	Type *FuncType  // function type
	Body *BlockStmt // function body
}

type File struct {
	Root Expr

	Stmts []Stmt // 语句
	//Funcs []*FuncDecl // 函数

	Scope *Scope // 作用范围
}

func (b *BasicLit) Pos() token.Pos   { return b.LitPos }
func (b *UnaryExpr) Pos() token.Pos  { return b.OpPos }
func (b *BinaryExpr) Pos() token.Pos { return b.X.Pos() } // left's pos
func (f *File) Pos() token.Pos       { return f.Root.Pos() }
func (x *Ident) Pos() token.Pos      { return x.NamePos }
func (x *CallExpr) Pos() token.Pos   { return x.Fun.Pos() }
func (x *BadExpr) Pos() token.Pos  { return x.From }
func (x *FuncLit) Pos() token.Pos  { return x.Type.Pos() }
func (x *FuncType) Pos() token.Pos {
	if x.Func.IsValid() || x.Params == nil { // see issue 3870
		return x.Func
	}
	return x.Params.Pos() // interface method declarations have no "func" keyword
}

func (b *BasicLit) End() token.Pos   { return b.LitPos + token.Pos(len(b.Lit)) }
func (b *UnaryExpr) End() token.Pos  { return b.X.End() }
func (b *BinaryExpr) End() token.Pos { return b.Y.End() } // right's end
func (f *File) End() token.Pos       { return f.Root.End() }
func (x *Ident) End() token.Pos      { return token.Pos(int(x.NamePos) + len(x.Name)) }
func (x *CallExpr) End() token.Pos   { return x.Rparen + 1 }
func (x *BadExpr) End() token.Pos  { return x.To }
func (x *FuncLit) End() token.Pos   { return x.Body.End() }
func (x *FuncType) End() token.Pos {
	if x.Results != nil {
		return x.Results.End()
	}
	return x.Params.End()
}

func (b *BasicLit) exprNode()   {}
func (b *BinaryExpr) exprNode() {}
func (b *UnaryExpr) exprNode()  {}
func (b *Ident) exprNode()      {}
func (b *CallExpr) exprNode()    {}
func (b *BadExpr) exprNode()     {}
func (b *FuncLit) exprNode()     {}
func (*FuncType) exprNode()      {}

//-----------
// field

// A Field represents a Field declaration list in a struct type,
// a method list in an interface type, or a parameter/result declaration
// in a signature.
// Field.Names is nil for unnamed parameters (parameter lists which only contain types)
// and embedded struct fields. In the latter case, the field name is the type name.
//
type Field struct {
	Names []*Ident  // field/method/parameter names; or nil
	Type  Expr      // field/method/parameter type
	Tag   *BasicLit // field tag; or nil
}

func (f *Field) Pos() token.Pos {
	if len(f.Names) > 0 {
		return f.Names[0].Pos()
	}
	return f.Type.Pos()
}

func (f *Field) End() token.Pos {
	if f.Tag != nil {
		return f.Tag.End()
	}
	return f.Type.End()
}

// A FieldList represents a list of Fields, enclosed by parentheses or braces.
type FieldList struct {
	Opening token.Pos // position of opening parenthesis/brace, if any
	List    []*Field  // field list; or nil
	Closing token.Pos // position of closing parenthesis/brace, if any
}

func (f *FieldList) Pos() token.Pos {
	if f.Opening.IsValid() {
		return f.Opening
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if len(f.List) > 0 {
		return f.List[0].Pos()
	}
	return token.NoPos
}

func (f *FieldList) End() token.Pos {
	if f.Closing.IsValid() {
		return f.Closing + 1
	}
	// the list should not be empty in this case;
	// be conservative and guard against bad ASTs
	if n := len(f.List); n > 0 {
		return f.List[n-1].End()
	}
	return token.NoPos
}

// NumFields returns the number of parameters or struct fields represented by a FieldList.
func (f *FieldList) NumFields() int {
	n := 0
	if f != nil {
		for _, g := range f.List {
			m := len(g.Names)
			if m == 0 {
				m = 1
			}
			n += m
		}
	}
	return n
}

//-----------
// type

// A type is represented by a tree consisting of one
// or more of the following type-specific expression
// nodes.
//
type (
	// A FuncType node represents a function type.
	FuncType struct {
		Func    token.Pos  // position of "func" keyword (token.NoPos if there is no "func")
		Params  *FieldList // (incoming) parameters; non-nil
		Results *FieldList // (outgoing) results; or nil
	}
)
