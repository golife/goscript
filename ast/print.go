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

func printIdens(idents []*Ident) {
	for _, item := range idents {
		printIdent(item)
		fmt.Print(" ")
	}
}

func printIdent(ident *Ident) {
	if ident == nil {
		fmt.Print("nil")
		return
	}
	fmt.Print(ident.Name)
}

func printPrefix(level int)  {
	for i := 0; i < level; i++ {
		fmt.Print(".  ")
	}
}

func print(node Node, level int) {
	printPrefix(level)

	switch n := node.(type) {
	case *BasicLit:
		fmt.Println("BasicType:", "pos:", n.LitPos, "value:", n.Lit)
	case *BinaryExpr:
		fmt.Printf("BinaryExpr: opPos: %v, op: %v, source pos: %v-%v\n", n.OpPos, n.Op.String(), n.Pos(), n.End())
		if n.X != nil {
			print(n.X, level+1)
		}
		if n.Y != nil {
			print(n.Y, level+1)
		}
	case *File:
		fmt.Println("File:")
	case *VarStmt:
		fmt.Print("Var stmt: ", n.Tok.String(), " ")
		printIdens(n.Names)
		printIdent(n.Type)
		if n.Values != nil {
			fmt.Println(" = ")
			for _, v := range n.Values {
				print(v, level+1)
			}
		}
		fmt.Println("")

	case *Ident:
		fmt.Println("ident", n.Name)
	case *AssignStmt:
		fmt.Println("asign:", n.Tok.String())
		printPrefix(level)
		fmt.Println("left: ")
		for _, v := range n.Lhs {
			print(v, level+1)
		}
		//fmt.Println("")
		printPrefix(level)
		fmt.Println("right: ")
		for _, v := range n.Rhs {
			print(v, level+1)
		}
	case *CallExpr:
		fmt.Println("call:")
		print(n.Fun, level)
		printPrefix(level)
		fmt.Println("args: ")
		for _, v := range n.Args {
			print(v, level+1)
		}
	case *ExprStmt:
		print(n.X, level)
	default:
		fmt.Println("dunno what I got...that can't be good")
		fmt.Println(n)
	}
}
