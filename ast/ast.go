// Copyright (c) 2014, Rob Thornton
// All rights reserved.
// This source code is governed by a Simplied BSD-License. Please see the
// LICENSE included in this distribution for a copy of the full license
// or, if one is not included, you may also find a copy at
// http://opensource.org/licenses/BSD-2-Clause

package ast

import (
	"github.com/golife/goscript/token"
)

type Node interface {
	Pos() token.Pos
	End() token.Pos
}

// BasicLit
// Expression(BinaryExpr)
type Expr interface {
	Node
	exprNode()
}

type Expression struct {
	Opening token.Pos
	Closing token.Pos
}

// 字面值
type BasicLit struct {
	LitPos token.Pos
	Kind   token.Token
	Lit    string
}

// 二元操作
type BinaryExpr struct {
	Expression
	Op    token.Token
	OpPos token.Pos
	List  []Expr
}

type File struct {
	Root Expr
}

func (b *BasicLit) Pos() token.Pos   { return b.LitPos }
func (e *Expression) Pos() token.Pos { return e.Opening }
func (f *File) Pos() token.Pos       { return f.Root.Pos() }

func (b *BasicLit) End() token.Pos   { return b.LitPos + token.Pos(len(b.Lit)) }
func (e *Expression) End() token.Pos { return e.Closing }
func (f *File) End() token.Pos       { return f.Root.End() }

func (b *BasicLit) exprNode()   {}
func (e *Expression) exprNode() {}
