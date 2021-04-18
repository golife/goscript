// Copyright (c) 2014, Rob Thornton
// All rights reserved.
// This source code is governed by a Simplied BSD-License. Please see the
// LICENSE included in this distribution for a copy of the full license
// or, if one is not included, you may also find a copy at
// http://opensource.org/licenses/BSD-2-Clause

package ast

import (
	"fmt"
)

func Print(node Node) {
	Walk(node, 0, print)
}

func print(node Node, level int) {
	for i := 0; i < level; i++ {
		fmt.Print(".  ")
	}
	switch n := node.(type) {
	case *BasicLit:
		fmt.Println("BasicType:", "pos:", n.LitPos, "value:", n.Lit)
	case *BinaryExpr:
		fmt.Printf("BinaryExpr: opPos: %v, op: %v, source pos: %v-%v\n",  n.OpPos, n.Op.String(), n.Pos(), n.End())
	case *File:
		fmt.Println("File:")
	default:
		fmt.Println("dunno what I got...that can't be good")
		fmt.Println(n)
	}
}
