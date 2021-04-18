// Copyright (c) 2014, Rob Thornton
// All rights reserved.
// This source code is governed by a Simplied BSD-License. Please see the
// LICENSE included in this distribution for a copy of the full license
// or, if one is not included, you may also find a copy at
// http://opensource.org/licenses/BSD-2-Clause

package ast_test

import (
	"testing"

	"github.com/golife/goscript/ast"
	"github.com/golife/goscript/token"
)

func TestBasicLit(t *testing.T) {
	b := &ast.BasicLit{
		LitPos: token.Pos(1),
		Kind:   token.INTEGER,
		Lit:    "24",
	}
	pos, end := token.Pos(1), token.Pos(3)
	if b.Pos() != pos || b.End() != end {
		t.Fatal("Expected:", pos, end, "Got:", b.Pos(), b.End())
	}
}

func TestBinaryExpr(t *testing.T) {
	// 3 + 5
	x := &ast.BasicLit{
		LitPos: token.Pos(1),
		Kind:   token.INTEGER,
		Lit:    "3",
	}
	y := &ast.BasicLit{
		LitPos: token.Pos(5),
		Kind:   token.INTEGER,
		Lit:    "5",
	}
	b := &ast.BinaryExpr{
		Op:    token.ADD,
		OpPos: token.Pos(2),
		X: x,
		Y: y,
	}

	if b.Pos() != token.Pos(1) {
		t.Fatal("BinaryExpr: Expected: 1 Got:", b.Pos())
	}
	if b.End() != token.Pos(6) {
		t.Fatal("BinaryExpr: Expected: 6 Got:", b.End())
	}
}
