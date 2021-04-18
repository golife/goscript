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
type Expr interface {
	Node
	exprNode()
}

// 字面值
type BasicLit struct {
	LitPos token.Pos
	Kind   token.Token
	Lit    string
}

// 二元操作 X op Y
type BinaryExpr struct {
	X     Expr
	OpPos token.Pos
	Op    token.Token
	Y     Expr
}

type File struct {
	Root Expr
}

func (b *BasicLit) Pos() token.Pos   { return b.LitPos }
func (b *BinaryExpr) Pos() token.Pos { return b.X.Pos() } // left's pos
func (f *File) Pos() token.Pos       { return f.Root.Pos() }

func (b *BasicLit) End() token.Pos   { return b.LitPos + token.Pos(len(b.Lit)) }
func (b *BinaryExpr) End() token.Pos { return b.Y.End() } // right's end
func (f *File) End() token.Pos       { return f.Root.End() }

func (b *BasicLit) exprNode()   {}
func (b *BinaryExpr) exprNode()   {}
